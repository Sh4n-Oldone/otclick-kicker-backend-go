package player

import (
	"context"
	"errors"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"net/http"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/helpers"
	"runtime"
	"slices"
	"strconv"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/constant"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/error_templates"
	pkgerr "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/errors"
)

func (s *Service) Create(ctx context.Context, request *entities.CreatePlayerRequest) (int, error) {
	logger := s.logger.With().Str("service", "player.Create").Logger()

	// если запрос от имени мастера по турнирам, то...
	if request.Creator.Role.Name == constant.TournamentMaster {
		// игрок должен быть активен...
		if request.ActivePlayer == nil || *request.ActivePlayer == false {
			err := errors.New("игрок должен быть активен")
			logger.Error().Err(err).Msg("Failed create player request")
			return 0, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}

		master, err := s.rdbOperations.GetTournamentMasterByUserId(logger, ctx, request.Creator.ID)
		if err != nil {
			return 0, err
		}

		if request.CityIdParam == "" {
			// и его cityId не должен быть указан вовсе(наиболее вероятный и желаемый сценарий)
			request.CityID = master.City.ID

		} else {
			// или его cityId должен быть равен cityId мастера по турнирам(эта проверка для подстраховки)
			cityId, err := strconv.ParseInt(request.CityIdParam, 10, 64)
			if err != nil {
				err = errors.New(pkgerr.WrongParameterError + ": " + "cityId")
				return 0, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
			}

			if cityId != master.City.ID {
				err = errors.New(pkgerr.ErrCityIdNotEqualMasterCityId)
				return 0, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
			}

			request.CityID = cityId
		}

	} else {

		if request.CityIdParam == "" {
			err := errors.New(pkgerr.EmptyParameterError + ": " + "cityId")
			return 0, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}

		cityId, err := strconv.ParseInt(request.CityIdParam, 10, 64)
		if err != nil || cityId <= 0 {
			err = errors.New(pkgerr.WrongParameterError + ": " + "cityId")
			return 0, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}

		request.CityID = cityId
	}

	id, err := s.rwdbOperations.CreatePlayer(logger, ctx, *request, &s.config.RDB)
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
	logger := s.logger.With().Str("service", "player.Find").Logger()
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
	// и возвращается []entities.FullPlayerV2

	fullPlayers := make([]entities.FullPlayerV2, len(players))

	for i, p := range players {

		var (
			pastMatches []entities.MatchV2
			leagues     []entities.PlayersLeague
			tournaments []entities.PlayersTournament
			pastGames   []entities.GameShort
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
			teams, err = s.rdbOperations.GetTeamsByPlayerID(logger, ctx, p.ID, nil)
			return err
		})

		g.Go(func() error {
			var err error
			tournaments, err = s.rdbOperations.GetTournamentListByPlayerID(logger, ctx, int64(p.ID))
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

		fullPlayer := buildFullPlayer(p, pastMatches, leagues, tournaments, teams, pastGames)

		fullPlayers[i] = fullPlayer
	}

	return entities.FindPlayersResponse{
		Players:     nil,
		FullPlayers: fullPlayers,
	}, nil
}

func (s *Service) FindV2(ctx context.Context, player entities.FindPlayersRequest) (entities.FindPlayersResponse, error) {
	logger := s.logger.With().Str("service", "player.FindV2").Logger()

	players, err := s.rdbOperations.FindPlayersV2(logger, ctx, player)
	if err != nil {
		return entities.FindPlayersResponse{}, err
	}

	if player.KeepSimple != nil && *player.KeepSimple {
		return entities.FindPlayersResponse{
			Players:     players,
			FullPlayers: nil,
		}, nil
	}

	fullPlayers := make([]entities.FullPlayerV2, len(players))

	g, gCtx := errgroup.WithContext(ctx)

	g.SetLimit(runtime.NumCPU())

	for i, p := range players {
		i, p := i, p

		g.Go(func() error {
			var (
				pastMatches            []entities.MatchV2
				leagues                []entities.PlayersLeague
				tournaments            []entities.PlayersTournament
				pastGamesOfPlayerTeams []entities.GameShort
				teams                  []entities.TeamItem
			)

			// внутренний errgroup для параллельных запросов по одному игроку
			innerG, innerCtx := errgroup.WithContext(gCtx)

			innerG.Go(func() error {
				var err error
				pastMatches, err = s.rdbOperations.GetPastMatchesByPlayerID(logger, innerCtx, p.ID)
				return err
			})

			innerG.Go(func() error {
				var err error
				leagues, err = s.rdbOperations.GetLeaguesByPlayerID(logger, innerCtx, p.ID)
				return err
			})

			innerG.Go(func() error {
				var err error
				teams, err = s.rdbOperations.GetTeamsByPlayerID(logger, innerCtx, p.ID, nil)
				return err
			})

			innerG.Go(func() error {
				var err error
				tournaments, err = s.rdbOperations.GetTournamentListByPlayerID(logger, ctx, int64(p.ID))
				return err
			})

			if err = innerG.Wait(); err != nil {
				return err
			}

			var teamIds []int
			for _, team := range teams {
				teamIds = append(teamIds, team.ID)
			}

			pastGamesOfPlayerTeams, err = s.rdbOperations.GetPastGamesByPlayersTeams(logger, gCtx, teamIds)
			if err != nil {
				return err
			}

			fullPlayer := buildFullPlayer(p, pastMatches, leagues, tournaments, teams, pastGamesOfPlayerTeams)
			fullPlayers[i] = fullPlayer

			return nil
		})
	}

	if err = g.Wait(); err != nil {
		logger.Error().Err(err).Msg("failed from g.Wait")
		return entities.FindPlayersResponse{}, err
	}

	return entities.FindPlayersResponse{
		Players:     nil,
		FullPlayers: fullPlayers,
	}, nil
}

func (s *Service) Get(ctx context.Context, id int) (entities.FullPlayerV2, error) {
	logger := s.logger.With().Interface("service", "player.Get").Logger()
	timeout, cancel := context.WithTimeout(ctx, s.config.RDB.MaxIdleConnectionTimeout)
	defer cancel()

	var (
		player      entities.Player
		pastMatches []entities.MatchV2
		leagues     []entities.PlayersLeague
		tournaments []entities.PlayersTournament
		teams       []entities.TeamItem
	)

	g, ctx := errgroup.WithContext(timeout)

	g.Go(func() error {
		var err error
		player, err = s.rdbOperations.GetPlayerByID(logger, timeout, id, nil)
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
		teams, err = s.rdbOperations.GetTeamsByPlayerID(logger, ctx, id, nil)
		return err
	})

	g.Go(func() error {
		var err error
		tournaments, err = s.rdbOperations.GetTournamentListByPlayerID(logger, ctx, int64(id))
		return err
	})

	if err := g.Wait(); err != nil {
		return entities.FullPlayerV2{}, err
	}

	var teamIds []int
	for _, _team := range teams {
		teamIds = append(teamIds, _team.ID)
	}

	// игры с участием команд игрока
	pastGamesOfPlayersTeam, err := s.rdbOperations.GetPastGamesByPlayersTeams(logger, timeout, teamIds)
	if err != nil {
		return entities.FullPlayerV2{}, err
	}

	fullPlayer := buildFullPlayer(player, pastMatches, leagues, tournaments, teams, pastGamesOfPlayersTeam)

	return fullPlayer, nil
}

func (s *Service) GetByTeamID(ctx context.Context, teamID int) ([]entities.Player, error) {
	logger := s.logger.With().Str("service", "player.GetByTeamID").Logger()
	timeout, cancel := context.WithTimeout(ctx, s.config.RDB.MaxIdleConnectionTimeout)
	defer cancel()

	players, err := s.rdbOperations.GetPlayersByTeamID(logger, timeout, teamID, nil)
	if err != nil {
		return nil, err
	}

	return players, nil
}

func (s *Service) GetTournamentPlayerList(ctx context.Context, req *entities.GetTournamentPlayerListRequest) ([]entities.TournamentPlayerItem, error) {
	logger := s.logger.With().Str("service", "player.GetTournamentPlayerList").Logger()

	items := make([]entities.TournamentPlayerItem, 0, 8)

	players, err := s.rdbOperations.GetTournamentPlayers(ctx, logger, req.TournamentID, req.WithDeleted, nil)
	if err != nil {
		return nil, err
	}

	for _, p := range players {

		var selfTeam *entities.TournamentTeam
		otherTeams := make([]entities.TournamentTeam, 0)

		selfTeamName, _ := helpers.BuildTeamName(p.Name, p.SecondName, p.LastName, p.ID)

		teams, err := s.rdbOperations.GetTeamsByPlayerID(logger, ctx, int(p.ID), nil)
		if err != nil {
			return nil, err
		}

		haveSelfTeam := false

		for _, team := range teams {
			if team.Name == selfTeamName && !haveSelfTeam {

				selfTeam = &entities.TournamentTeam{
					ID:        int64(team.ID),
					Name:      team.Name,
					ShortName: team.ShortName,
					CityID:    int64(team.CityID),
					Avatar:    team.Avatar,
				}

				haveSelfTeam = true

			} else {
				otherTeams = append(otherTeams, entities.TournamentTeam{
					ID:        int64(team.ID),
					Name:      team.Name,
					ShortName: team.ShortName,
					CityID:    int64(team.CityID),
					Avatar:    team.Avatar,
				})
			}
		}

		items = append(items, entities.TournamentPlayerItem{
			Player:     p,
			SelfTeam:   selfTeam,
			OtherTeams: otherTeams,
		})
	}

	return items, nil
}

func buildFullPlayer(player entities.Player, pastMatches []entities.MatchV2, leagues []entities.PlayersLeague, tournaments []entities.PlayersTournament, teams []entities.TeamItem, pastGamesOfPlayerTeams []entities.GameShort) entities.FullPlayerV2 {

	leagueItems := getLeaguesStat(player, leagues, teams, pastMatches, pastGamesOfPlayerTeams)
	tournamentItems := getTournamentsStat(player, tournaments, teams, pastMatches, pastGamesOfPlayerTeams)
	totalStat := getTotalStat(player, pastMatches, pastGamesOfPlayerTeams)

	return entities.FullPlayerV2{
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
		Tournament:   tournamentItems,
		TotalStat:    totalStat,
	}
}

func getLeaguesStat(player entities.Player, leagues []entities.PlayersLeague, teams []entities.TeamItem, pastMatches []entities.MatchV2, pastGamesOfPlayerTeams []entities.GameShort) []entities.LeagueItem {

	type leagueStat struct {
		goalsScoredNumber   int
		goalsConcededNumber int
		playersGames        map[int]struct{}
		playersMatches      int
	}

	leagueStats := make(map[int]leagueStat)

	for _, match := range pastMatches {
		var stat leagueStat
		stat.playersGames = make(map[int]struct{})

		if match.LeagueID != nil {
			if entry, ok := leagueStats[*match.LeagueID]; ok {
				stat = entry
			}
		}

		// заполняем игры игрока
		stat.playersGames[match.GameID] = struct{}{}

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

		if match.LeagueID != nil {
			leagueStats[*match.LeagueID] = stat
		}
	}

	leaguePlayedGames := make(map[int]int)
	for _, game := range pastGamesOfPlayerTeams {
		games := 0
		if game.LeagueID != nil {
			if entry, ok := leaguePlayedGames[*game.LeagueID]; ok {
				games = entry
			}
			games++
			leaguePlayedGames[*game.LeagueID] = games
		}
	}

	leagueItems := make([]entities.LeagueItem, len(leagues))
	for i, league := range leagues {
		leagueItems[i].ID = league.ID
		leagueItems[i].Name = league.Name
		if league.Rating != nil {
			leagueItems[i].Rating = *league.Rating
		} else {
			leagueItems[i].Rating = constant.DefaultRating
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

	return leagueItems
}

func getTournamentsStat(player entities.Player, tournaments []entities.PlayersTournament, teams []entities.TeamItem, pastMatches []entities.MatchV2, pastGamesOfPlayerTeams []entities.GameShort) []entities.TournamentItem {

	type tournamentStat struct {
		goalsScoredNumber   int
		goalsConcededNumber int
		playersGames        map[int]struct{}
		playersMatches      int
	}

	tournamentStats := make(map[int]tournamentStat)

	for _, match := range pastMatches {
		var stat tournamentStat
		stat.playersGames = make(map[int]struct{})

		if match.TournamentID != nil {
			if entry, ok := tournamentStats[*match.TournamentID]; ok {
				stat = entry
			}
		}

		// заполняем игры игрока
		stat.playersGames[match.GameID] = struct{}{}

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

		if match.TournamentID != nil {
			tournamentStats[*match.TournamentID] = stat
		}
	}

	tournamentPlayedGames := make(map[int]int)
	for _, game := range pastGamesOfPlayerTeams {
		games := 0
		if game.TournamentID != nil {
			if entry, ok := tournamentPlayedGames[*game.TournamentID]; ok {
				games = entry
			}
			games++
			tournamentPlayedGames[*game.TournamentID] = games
		}
	}

	tournamentItems := make([]entities.TournamentItem, len(tournaments))
	for i, tournament := range tournaments {
		tournamentItems[i].ID = int(tournament.ID)
		tournamentItems[i].Name = tournament.Name
		if tournament.Rating != nil {
			tournamentItems[i].Rating = int(*tournament.Rating)
		} else {
			tournamentItems[i].Rating = constant.DefaultRating
		}

		games := 0
		if entry, ok := tournamentPlayedGames[int(tournament.ID)]; ok {
			games = entry
		}

		if entry, ok := tournamentStats[int(tournament.ID)]; ok {
			tournamentItems[i].GamesPlayedNumber = len(entry.playersGames)
			tournamentItems[i].GoalsScoredNumber = entry.goalsScoredNumber
			tournamentItems[i].GoalsConcededNumber = entry.goalsConcededNumber
			tournamentItems[i].MatchesPlayed = entry.playersMatches

			if games > 0 {
				tournamentItems[i].PercentageOfParticipation = (float32(len(entry.playersGames)) / float32(games)) * 100
			}
		}

		for _, team := range teams {
			if slices.Contains(team.Tournaments, int(tournament.ID)) {
				tournamentItems[i].Teams = append(tournamentItems[i].Teams, team)
			}
		}
	}

	return tournamentItems
}

func getTotalStat(player entities.Player, pastMatches []entities.MatchV2, pastGamesOfPlayerTeams []entities.GameShort) entities.PlayerTotalStat {

	totalStat := entities.PlayerTotalStat{}

	totalStat.MatchesPlayed = len(pastMatches)
	gamesPlayedByPlayer := make(map[int]struct{})

	for _, match := range pastMatches {

		gamesPlayedByPlayer[match.GameID] = struct{}{}

		if match.Player1Team1ID == player.ID || (match.Player2Team1ID != nil && *match.Player2Team1ID == player.ID) {
			totalStat.GoalsScoredNumber += match.ScoreTeam1
			totalStat.GoalsConcededNumber += match.ScoreTeam2
		}
		if match.Player1Team2ID == player.ID || (match.Player2Team2ID != nil && *match.Player2Team2ID == player.ID) {
			totalStat.GoalsScoredNumber += match.ScoreTeam2
			totalStat.GoalsConcededNumber += match.ScoreTeam1
		}
	}

	totalStat.GamesPlayedNumber = len(gamesPlayedByPlayer)
	totalStat.PercentageOfParticipation = float32(len(gamesPlayedByPlayer)) / float32(len(pastGamesOfPlayerTeams)) * 100

	return totalStat
}
