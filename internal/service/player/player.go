package player

import (
	"context"
	"slices"

	"golang.org/x/sync/errgroup"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
)

func (s *Service) Create(ctx context.Context, player entities.CreatePlayerRequest) (int, error) {
	logger := s.logger.With().Interface("service", "player.Create").Logger()
	timeout, cancel := context.WithTimeout(ctx, s.config.RWDB.MaxIdleConnectionTimeout)
	defer cancel()

	id, err := s.rwdbOperations.CreatePlayer(logger, timeout, player)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *Service) Delete(ctx context.Context, playerID int) error {
	logger := s.logger.With().Interface("service", "player.Delete").Logger()
	timeout, cancel := context.WithTimeout(ctx, s.config.RWDB.MaxIdleConnectionTimeout)
	defer cancel()

	err := s.rwdbOperations.DeletePlayer(logger, timeout, playerID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Recover(ctx context.Context, playerID int) error {
	logger := s.logger.With().Interface("service", "player.Recover").Logger()
	timeout, cancel := context.WithTimeout(ctx, s.config.RWDB.MaxIdleConnectionTimeout)
	defer cancel()

	err := s.rwdbOperations.RecoverPlayer(logger, timeout, playerID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Update(ctx context.Context, player entities.UpdatePlayerRequest) error {
	logger := s.logger.With().Interface("service", "player.Update").Logger()
	timeout, cancel := context.WithTimeout(ctx, s.config.RWDB.MaxIdleConnectionTimeout)
	defer cancel()

	err := s.rwdbOperations.UpdatePlayer(logger, timeout, player)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Find(ctx context.Context, player entities.FindPlayersRequest) (entities.FindPlayersResponse, error) {
	logger := s.logger.With().Interface("service", "player.Find").Logger()
	timeout, cancel := context.WithTimeout(ctx, s.config.RDB.MaxIdleConnectionTimeout)
	defer cancel()

	players, err := s.rdbOperations.FindPlayers(logger, timeout, player)
	if err != nil {
		return entities.FindPlayersResponse{}, err
	}

	// если нужно краткое описание игрока(KeepSimple == true) то метод прекращает выполнение здесь
	// и возвращается []entities.Player
	if player.KeepSimple != nil && *player.KeepSimple {
		return entities.FindPlayersResponse{
			Players:     players,
			FullPlayers: nil,
		}, nil
	}

	// если нужно полное описание игрока(KeepSimple != true) то метод продолжает выполнение
	// и возвращается []entities.FullPlayer

	fullPlayers := make([]entities.FullPlayer, len(players))

	for i, p := range players {

		var (
			pastMatches []entities.Match
			leagues     []entities.PlayersLeague
			pastGames   []entities.Game
			teams       []entities.TeamItem
		)

		g, ctx := errgroup.WithContext(timeout)

		g.Go(func() error {
			var err error
			pastMatches, err = s.rdbOperations.GetPastMatchesByPlayerID(logger, ctx, p.ID)
			return err
		})

		g.Go(func() error {
			var err error
			leagues, err = s.rdbOperations.GetLeaguesByPlayerID(logger, ctx, p.ID)
			return err
		})

		g.Go(func() error {
			var err error
			teams, err = s.rdbOperations.GetTeamsByPlayerID(logger, ctx, p.ID)
			return err
		})

		if err = g.Wait(); err != nil {
			return entities.FindPlayersResponse{}, err
		}

		var teamIds []int
		for _, _team := range teams {
			teamIds = append(teamIds, _team.ID)
		}

		pastGames, err = s.rdbOperations.GetPastGamesByPlayersTeams(logger, timeout, teamIds)
		if err != nil {
			return entities.FindPlayersResponse{}, err
		}

		fullPlayer := buildFullPlayer(p, pastMatches, leagues, teams, pastGames)

		fullPlayers[i] = fullPlayer
	}

	return entities.FindPlayersResponse{
		Players:     nil,
		FullPlayers: fullPlayers,
	}, nil
}

func (s *Service) Get(ctx context.Context, id int) (entities.FullPlayer, error) {
	logger := s.logger.With().Interface("service", "player.Get").Logger()
	timeout, cancel := context.WithTimeout(ctx, s.config.RDB.MaxIdleConnectionTimeout)
	defer cancel()

	var (
		player      entities.Player
		pastMatches []entities.Match
		leagues     []entities.PlayersLeague
		teams       []entities.TeamItem
	)

	g, ctx := errgroup.WithContext(timeout)

	g.Go(func() error {
		var err error
		player, err = s.rdbOperations.GetPlayerByID(logger, timeout, id)
		return err
	})

	g.Go(func() error {
		var err error
		pastMatches, err = s.rdbOperations.GetPastMatchesByPlayerID(logger, ctx, id)
		return err
	})

	g.Go(func() error {
		var err error
		leagues, err = s.rdbOperations.GetLeaguesByPlayerID(logger, ctx, id)
		return err
	})

	g.Go(func() error {
		var err error
		teams, err = s.rdbOperations.GetTeamsByPlayerID(logger, ctx, id)
		return err
	})

	if err := g.Wait(); err != nil {
		return entities.FullPlayer{}, err
	}

	var teamIds []int
	for _, _team := range teams {
		teamIds = append(teamIds, _team.ID)
	}

	// игры с участием команд игрока
	pastGamesOfPlayersTeam, err := s.rdbOperations.GetPastGamesByPlayersTeams(logger, timeout, teamIds)
	if err != nil {
		return entities.FullPlayer{}, err
	}

	fullPlayer := buildFullPlayer(player, pastMatches, leagues, teams, pastGamesOfPlayersTeam)

	return fullPlayer, nil
}

func (s *Service) GetByTeamID(ctx context.Context, teamID int) ([]entities.Player, error) {
	logger := s.logger.With().Interface("service", "player.GetByTeamID").Logger()
	timeout, cancel := context.WithTimeout(ctx, s.config.RDB.MaxIdleConnectionTimeout)
	defer cancel()

	players, err := s.rdbOperations.GetPlayersByTeamID(logger, timeout, teamID)
	if err != nil {
		return nil, err
	}

	return players, nil
}

func buildFullPlayer(player entities.Player, pastMatches []entities.Match, leagues []entities.PlayersLeague, teams []entities.TeamItem, pastGames []entities.Game) entities.FullPlayer {
	type leagueStat struct {
		goalsScoredNumber   int
		goalsConcededNumber int
		playersGames        map[int]bool
		playersMatches      int
	}

	leagueStats := make(map[int]leagueStat)

	for _, match := range pastMatches {
		var stat leagueStat
		stat.playersGames = make(map[int]bool)
		if entry, ok := leagueStats[*match.LeagueID]; ok {
			stat = entry
		}

		// заполняем игры игрока
		stat.playersGames[match.GameID] = false

		// подсчет голов
		if match.Player1Team1ID == player.ID || (match.Player2Team1ID != nil && *match.Player2Team1ID == player.ID) {
			stat.goalsScoredNumber += match.ScoreTeam1
			stat.goalsConcededNumber += match.ScoreTeam2
			stat.playersMatches++
		}
		if match.Player1Team2ID == player.ID || (match.Player2Team2ID != nil && *match.Player2Team2ID == player.ID) {
			stat.goalsScoredNumber += match.ScoreTeam2
			stat.goalsConcededNumber += match.ScoreTeam1
			stat.playersMatches++
		}

		leagueStats[*match.LeagueID] = stat
	}

	leaguePlayedGames := make(map[int]int)
	for _, game := range pastGames {
		games := 0
		if entry, ok := leaguePlayedGames[*game.LeagueID]; ok {
			games = entry
		}
		games++
		leaguePlayedGames[*game.LeagueID] = games
	}

	leagueItems := make([]entities.LeagueItem, len(leagues))
	for i, league := range leagues {
		leagueItems[i].ID = league.ID
		leagueItems[i].Name = league.Name
		if league.Rating != nil {
			leagueItems[i].Rating = *league.Rating
		}

		games := 0
		if entry, ok := leaguePlayedGames[league.ID]; ok {
			games = entry
		}

		if entry, ok := leagueStats[league.ID]; ok {
			leagueItems[i].GamesPlayedNumber = len(entry.playersGames)
			leagueItems[i].GoalsScoredNumber = entry.goalsScoredNumber
			leagueItems[i].GoalsConcededNumber = entry.goalsConcededNumber
			leagueItems[i].MatchesPlayed = entry.playersMatches

			if games > 0 {
				leagueItems[i].PercentageOfParticipation = (float32(len(entry.playersGames)) / float32(games)) * 100
			}
		}

		for _, team := range teams {
			if slices.Contains(team.Leagues, league.ID) {
				leagueItems[i].Teams = append(leagueItems[i].Teams, team)
			}
		}
	}

	return entities.FullPlayer{
		ID:           player.ID,
		Name:         player.Name,
		SecondName:   player.SecondName,
		LastName:     player.LastName,
		Avatar:       player.Avatar,
		ActivePlayer: player.ActivePlayer,
		Deleted:      player.DeletedAt != nil,
		CityID:       player.CityID,
		CityName:     player.CityName,
		Leagues:      leagueItems,
	}
}
