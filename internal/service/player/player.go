package player

import (
	"context"
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
	if player.KeepSimple != nil && *player.KeepSimple == true {
		return entities.FindPlayersResponse{
			Players:     players,
			FullPlayers: nil,
		}, nil
	}

	// если нужно полное описание игрока(KeepSimple != true) то метод продолжает выполнение
	// и возвращается []entities.FullPlayer
	var (
		pastMatches []entities.Match
		leagues     []entities.PlayersLeague
		pastGames   []entities.Game
		teamId      int
	)

	fullPlayers := make([]entities.FullPlayer, len(players))

	for i, p := range players {

		g, ctx := errgroup.WithContext(timeout)

		if p.TeamID != nil {
			teamId = *p.TeamID
		}

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
			pastGames, err = s.rdbOperations.GetPastGamesByPlayersTeam(logger, ctx, teamId)
			return err
		})

		if err = g.Wait(); err != nil {
			return entities.FindPlayersResponse{}, err
		}

		fullPlayer := buildFullPlayer(p, pastMatches, leagues, pastGames)

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
		teamId      int
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

	if err := g.Wait(); err != nil {
		return entities.FullPlayer{}, err
	}

	if player.TeamID != nil {
		teamId = *player.TeamID
	}

	// игры с участием команды игрока
	pastGamesOfPlayersTeam, err := s.rdbOperations.GetPastGamesByPlayersTeam(logger, timeout, teamId)
	if err != nil {
		return entities.FullPlayer{}, err
	}

	fullPlayer := buildFullPlayer(player, pastMatches, leagues, pastGamesOfPlayersTeam)

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

func buildFullPlayer(player entities.Player, pastMatches []entities.Match, leagues []entities.PlayersLeague, pastGames []entities.Game) entities.FullPlayer {
	//тут будут id игр с участием игрока для подсчета
	playersGames := make(map[int]struct{})

	var (
		percentageOfParticipation float32
		goalsScoredNumber         int
		goalsConcededNumber       int
	)

	for _, match := range pastMatches {
		// заполняем игры игрока
		playersGames[match.GameID] = struct{}{}

		// подсчет голов
		if match.Player1Team1ID == player.ID || (match.Player2Team1ID != nil && *match.Player2Team1ID == player.ID) {
			goalsScoredNumber += match.ScoreTeam1
			goalsConcededNumber += match.ScoreTeam2
		}
		if match.Player1Team2ID == player.ID || (match.Player2Team2ID != nil && *match.Player2Team2ID == player.ID) {
			goalsScoredNumber += match.ScoreTeam2
			goalsConcededNumber += match.ScoreTeam1
		}
	}

	if len(pastGames) > 0 {
		//процент участия игрока в играх команды
		percentageOfParticipation = (float32(len(playersGames)) / float32(len(pastGames))) * 100
	}

	leagueItems := make([]entities.LeagueItem, len(leagues))

	for i, league := range leagues {
		leagueItems[i].ID = league.ID
		leagueItems[i].Name = league.Name
		leagueItems[i].Rating = league.Rating
	}

	return entities.FullPlayer{
		ID:                        player.ID,
		Name:                      player.Name,
		SecondName:                player.SecondName,
		LastName:                  player.LastName,
		Avatar:                    player.Avatar,
		MatchesPlayed:             len(pastMatches),
		GoalsScoredNumber:         goalsScoredNumber,
		GoalsConcededNumber:       goalsConcededNumber,
		GamesPlayedNumber:         len(playersGames),
		PercentageOfParticipation: percentageOfParticipation,
		ActivePlayer:              player.ActivePlayer,
		Deleted:                   player.DeletedAt != nil,
		TeamName:                  player.TeamName,
		TeamShortName:             player.TeamShortName,
		CityID:                    player.CityID,
		CityName:                  player.CityName,
		Leagues:                   leagueItems,
		Rating:                    player.Rating,
	}
}
