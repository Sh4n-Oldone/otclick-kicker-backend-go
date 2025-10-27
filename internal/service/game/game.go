package game

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"time"

	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/constant"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/calculator"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/error_templates"
	pkgerr "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/errors"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/helpers/pointer"
)

func (s *Service) Create(ctx context.Context, request entities.CreateGameRequest) (entities.CreateGameResponse, error) {
	logger := s.logger.With().Str("service", "game.Create").Logger()
	timeout, cancel := context.WithTimeout(ctx, s.config.RWDB.MaxIdleConnectionTimeout)
	defer cancel()

	league, err := s.rdbOperations.GetLeagueById(logger, ctx, int64(request.LeagueID), nil)
	if err != nil {
		return entities.CreateGameResponse{}, err
	}

	team1resp, err := s.rdbOperations.GetTeam(logger, ctx, int64(request.Team1ID), &s.config.RDB)
	if err != nil {
		return entities.CreateGameResponse{}, err
	}
	team2resp, err := s.rdbOperations.GetTeam(logger, ctx, int64(request.Team2ID), &s.config.RDB)
	if err != nil {
		return entities.CreateGameResponse{}, err
	}

	// проверяем, что команды из одной лиги
	leagueID := int64(request.LeagueID)
	if !bothContain(leagueID, team1resp.Leagues, team2resp.Leagues) {
		logger.Error().Stack().Err(err).Msg("failed to Create (Played Game) : leagueID is missing from one or both arrays")
		err = errors.New(pkgerr.FailedGameByTeamsLeagueMismatch)
		return entities.CreateGameResponse{}, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	// проверяем относится ли капитан к какой-либо из команд
	if request.Creator.Role.Name == constant.CaptainRole {
		err = s.validateCaptain(ctx, logger, request.Creator.ID, request.Creator.Team.ID, int64(request.Team1ID), int64(request.Team2ID))
		if err != nil {
			return entities.CreateGameResponse{}, err
		}
	}

	// проверяем город мастера по турнирам на соответствие городу лиги и команд
	if request.Creator.Role.Name == constant.TournamentMaster {
		master, err := s.rdbOperations.GetTournamentMasterByUserId(logger, ctx, request.Creator.ID, &s.config.RDB)
		if err != nil {
			return entities.CreateGameResponse{}, err
		}

		if (team1resp.CityId == nil || *team1resp.CityId != master.City.ID) || (team2resp.CityId == nil || *team2resp.CityId != master.City.ID) {
			err = errors.New("город команды и город мастера по турнирам не совпадают")
			logger.Error().Err(err).Msg("master city not equal league city")
			return entities.CreateGameResponse{}, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}

		if league.CityID != master.City.ID {
			err = errors.New("город лиги и город мастера по турнирам не совпадают")
			logger.Error().Err(err).Msg("master city not equal league city")
			return entities.CreateGameResponse{}, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}

		if master.City.ID != int64(request.CityID) {
			err = errors.New("город в запросе и город мастера по турнирам не совпадают")
			logger.Error().Err(err).Msg("master city not equal request city")
			return entities.CreateGameResponse{}, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}
	}

	if league.CityID != int64(request.CityID) {
		err = errors.New("город в запросе и город лиги не совпадают")
		logger.Error().Err(err).Msg("request city not equal league city")
		return entities.CreateGameResponse{}, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	// Map for keeping player ratings while going game calculation
	rates := make(map[int]int, 0)

	var matches []entities.GamesMatch

	if len(request.Matches) > 0 {
		for _, match := range request.Matches {
			if match.ScoreTeam1 == 0 && match.ScoreTeam2 == 0 {
				continue
			}

			player1team1ID := match.Player1Team1Id
			player1team2ID := match.Player1Team2Id

			// Получаем рейтинги для всех игроков
			player1team1rate, err := s.getPlayerRating(ctx, logger, player1team1ID, leagueID, rates)
			if err != nil {
				return entities.CreateGameResponse{}, err
			}

			player1team2rate, err := s.getPlayerRating(ctx, logger, player1team2ID, leagueID, rates)
			if err != nil {
				return entities.CreateGameResponse{}, err
			}

			var player2team1rate, player2team2rate int
			var player2team1ID, player2team2ID int

			// Получаем рейтинг для player2team1 (если существует и > 0)
			if match.Player2Team1Id != nil {
				player2team1ID = *match.Player2Team1Id
				if player2team1ID > 0 {
					player2team1rate, err = s.getPlayerRating(ctx, logger, player2team1ID, leagueID, rates)
					if err != nil {
						return entities.CreateGameResponse{}, err
					}
				}
			}

			// Получаем рейтинг для player2team2 (если существует и > 0)
			if match.Player2Team2Id != nil {
				player2team2ID = *match.Player2Team2Id
				if player2team2ID > 0 {
					player2team2rate, err = s.getPlayerRating(ctx, logger, player2team2ID, leagueID, rates)
					if err != nil {
						return entities.CreateGameResponse{}, err
					}
				}
			}

			p1t1r := player1team1rate
			p1t2r := player1team2rate
			p2t1r := player2team1rate
			p2t2r := player2team2rate

			match.Player1Team1RateBefore = &p1t1r
			match.Player1Team2RateBefore = &p1t2r
			match.Player2Team1RateBefore = &p2t1r
			match.Player2Team2RateBefore = &p2t2r

			player1team1rateAfter, player1team2rateAfter, player2team1rateAfter, player2team2rateAfter, err := calculator.MatchRatingCalculation(
				ctx, match.ScoreTeam1, match.ScoreTeam2, player1team1rate, player1team2rate, player2team1rate, player2team2rate)
			if err != nil {
				return entities.CreateGameResponse{}, err
			}

			match.Player1Team1RateAfter = &player1team1rateAfter
			match.Player1Team2RateAfter = &player1team2rateAfter
			match.Player2Team1RateAfter = &player2team1rateAfter
			match.Player2Team2RateAfter = &player2team2rateAfter

			rates[match.Player1Team1Id] = player1team1rateAfter
			rates[match.Player1Team2Id] = player1team2rateAfter
			if player2team1ID > 0 {
				rates[*match.Player2Team1Id] = player2team1rateAfter
			}
			if player2team2ID > 0 {
				rates[*match.Player2Team2Id] = player2team2rateAfter
			}

			matches = append(matches, match)
		}

		request.Matches = matches
	}

	operator := "insertIgnore"

	resp, err := s.rwdbOperations.CreateGameWithRating(logger, timeout, request, rates, &operator, leagueID)
	if err != nil {
		return entities.CreateGameResponse{}, err
	}

	return resp, nil
}

func (s *Service) Delete(ctx context.Context, req *entities.DeleteGameRequest) error {
	logger := s.logger.With().Interface("service", "game.Delete").Logger()

	recalc := true
	game, err := s.rdbOperations.GetGame(logger, ctx, req.ID)
	if err != nil {
		return err
	}

	gameLeagueId := game.LeagueID
	if gameLeagueId == nil {
		err = errors.New(pkgerr.ErrGameIsNotPartOfLeague)
		logger.Error().Err(err).Msg("the game is not part of league")
		return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	// Getting league for teams and validate it
	team1resp, err := s.rdbOperations.GetTeam(logger, ctx, int64(game.Team1ID), &s.config.RDB)
	if err != nil {
		return err
	}
	team2resp, err := s.rdbOperations.GetTeam(logger, ctx, int64(game.Team2ID), &s.config.RDB)
	if err != nil {
		return err
	}

	league := &entities.League{}

	*league, err = s.rdbOperations.GetLeagueById(logger, ctx, int64(*gameLeagueId), nil)
	if err != nil {
		return err
	}

	if team1resp.Leagues == nil || team2resp.Leagues == nil ||
		(team1resp.Leagues != nil && team2resp.Leagues != nil && !bothContain(int64(*gameLeagueId), team1resp.Leagues, team2resp.Leagues)) {
		recalc = false
	}

	if req.Executor.Role.Name == constant.CaptainRole {
		err = s.validateCaptain(ctx, logger, req.Executor.ID, req.Executor.Team.ID, int64(game.Team1ID), int64(game.Team2ID))
		if err != nil {
			return err
		}
	}

	if req.Executor.Role.Name == constant.TournamentMaster {
		master, err := s.rdbOperations.GetTournamentMasterByUserId(logger, ctx, req.Executor.ID, &s.config.RDB)
		if err != nil {
			return err
		}

		if int64(game.CityID) != master.City.ID {
			err = errors.New("город игры и город мастера по турнирам не совпадают")
			logger.Error().Err(err).Msg("master city not equal game city")
			return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}

		if league.CityID != master.City.ID {
			err = errors.New("город лиги игры и город мастера по турнирам не совпадают")
			logger.Error().Err(err).Msg("master city not equal league city")
			return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}
	}

	if recalc {
		// Decreasing rating for each player of game matches
		plGmRtInc := make(map[int]int, 0) // playerGameRatingIncrease
		_matches, err := s.rdbOperations.GetMatchListByGameID(ctx, logger, int64(game.ID), &s.config.RDB)
		if err != nil {
			return err
		}
		for _, _match := range _matches {
			// player1team1
			if _match.Player1Team1RateBefore != _match.Player1Team1RateAfter {
				_, ok := plGmRtInc[_match.Player1Team1ID]
				if ok {
					plGmRtInc[_match.Player1Team1ID] += *_match.Player1Team1RateAfter - *_match.Player1Team1RateBefore
				} else {
					plGmRtInc[_match.Player1Team1ID] = *_match.Player1Team1RateAfter - *_match.Player1Team1RateBefore
				}
			}
			// player1team2
			if _match.Player1Team2RateBefore != _match.Player1Team2RateAfter {
				_, ok := plGmRtInc[_match.Player1Team2ID]
				if ok {
					plGmRtInc[_match.Player1Team2ID] += *_match.Player1Team2RateAfter - *_match.Player1Team2RateBefore
				} else {
					plGmRtInc[_match.Player1Team2ID] = *_match.Player1Team2RateAfter - *_match.Player1Team2RateBefore
				}
			}
			// player2team1
			if _match.Player2Team1ID != nil && _match.Player2Team1RateBefore != _match.Player2Team1RateAfter {
				_, ok := plGmRtInc[*_match.Player2Team1ID]
				if ok {
					plGmRtInc[*_match.Player2Team1ID] += *_match.Player2Team1RateAfter - *_match.Player2Team1RateBefore
				} else {
					plGmRtInc[*_match.Player2Team1ID] = *_match.Player2Team1RateAfter - *_match.Player2Team1RateBefore
				}
			}
			// player2team2
			if _match.Player2Team2ID != nil && _match.Player2Team2RateBefore != _match.Player2Team2RateAfter {
				_, ok := plGmRtInc[*_match.Player2Team2ID]
				if ok {
					plGmRtInc[*_match.Player2Team2ID] += *_match.Player2Team2RateAfter - *_match.Player2Team2RateBefore
				} else {
					plGmRtInc[*_match.Player2Team2ID] = *_match.Player2Team2RateAfter - *_match.Player2Team2RateBefore
				}
			}
		}
		for playerID, value := range plGmRtInc {
			rateValue, err := s.rdbOperations.GetRatingByPlayerIDAndByLeagueID(logger, ctx, int64(playerID), int64(*gameLeagueId))
			if err != nil {
				return err
			}

			rate := &entities.Rating{
				PlayerID: int64(playerID),
				LeagueID: int64(*gameLeagueId),
				Value:    rateValue - int64(value),
			}
			operator := "insertIgnore"
			err = s.rwdbOperations.CreateRating(logger, ctx, *rate, &operator)
			if err != nil {
				return err
			}
		}
	}

	err = s.rwdbOperations.DeleteGame(logger, ctx, int64(req.ID), &s.config.RWDB)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Get(ctx context.Context, gameID int) (entities.GetGameResponse, error) {
	logger := s.logger.With().Interface("service", "game.Get").Logger()
	timeout, cancel := context.WithTimeout(ctx, s.config.RDB.MaxIdleConnectionTimeout)
	defer cancel()

	resp, err := s.rdbOperations.GetGame(logger, timeout, gameID)
	if err != nil {
		return entities.GetGameResponse{}, err
	}
	return resp, nil
}

func (s *Service) Update(ctx context.Context, request entities.UpdateGameRequest) error {
	logger := s.logger.With().Interface("service", "game.Update").Logger()
	timeout, cancel := context.WithTimeout(ctx, s.config.RWDB.MaxIdleConnectionTimeout)
	defer cancel()

	league, err := s.rdbOperations.GetLeagueById(logger, ctx, int64(request.LeagueID), nil)
	if err != nil {
		return err
	}

	team1resp, err := s.rdbOperations.GetTeam(logger, ctx, int64(request.Team1ID), &s.config.RDB)
	if err != nil {
		return err
	}
	team2resp, err := s.rdbOperations.GetTeam(logger, ctx, int64(request.Team2ID), &s.config.RDB)
	if err != nil {
		return err
	}

	// проверяем, что команды из одной лиги
	leagueID := int64(request.LeagueID)
	if !bothContain(leagueID, team1resp.Leagues, team2resp.Leagues) {
		logger.Error().Stack().Err(err).Msg("failed to Update (Played Game) : leagueID is missing from one or both arrays")
		err = errors.New(pkgerr.FailedGameByTeamsLeagueMismatch)
		return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	// проверяем относится ли капитан к какой-либо из команд
	if request.Executor.Role.Name == constant.CaptainRole {
		err = s.validateCaptain(ctx, logger, request.Executor.ID, request.Executor.Team.ID, int64(request.Team1ID), int64(request.Team2ID))
		if err != nil {
			return err
		}
	}

	// проверяем город мастера по турнирам на соответствие городу лиги и команд
	if request.Executor.Role.Name == constant.TournamentMaster {
		master, err := s.rdbOperations.GetTournamentMasterByUserId(logger, ctx, request.Executor.ID, &s.config.RDB)
		if err != nil {
			return err
		}

		if (team1resp.CityId == nil || *team1resp.CityId != master.City.ID) || (team2resp.CityId == nil || *team2resp.CityId != master.City.ID) {
			err = errors.New("город команды и город мастера по турнирам не совпадают")
			logger.Error().Err(err).Msg("master city not equal league city")
			return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}

		if league.CityID != master.City.ID {
			err = errors.New("город лиги и город мастера по турнирам не совпадают")
			logger.Error().Err(err).Msg("master city not equal league city")
			return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}
	}

	rates := make(map[int]int, 0)

	// Decreasing rating for each player of game matches
	plGmRtInc := make(map[int]int, 0) // playerGameRatingIncrease
	_matches, err := s.rdbOperations.GetMatchListByGameID(ctx, logger, int64(request.ID), &s.config.RDB)
	if err != nil {
		return err
	}
	for _, _match := range _matches {
		// player1team1
		if _match.Player1Team1RateBefore != _match.Player1Team1RateAfter {
			_, ok := plGmRtInc[_match.Player1Team1ID]
			if ok {
				plGmRtInc[_match.Player1Team1ID] += *_match.Player1Team1RateAfter - *_match.Player1Team1RateBefore
			} else {
				plGmRtInc[_match.Player1Team1ID] = *_match.Player1Team1RateAfter - *_match.Player1Team1RateBefore
			}
		}
		// player1team2
		if _match.Player1Team2RateBefore != _match.Player1Team2RateAfter {
			_, ok := plGmRtInc[_match.Player1Team2ID]
			if ok {
				plGmRtInc[_match.Player1Team2ID] += *_match.Player1Team2RateAfter - *_match.Player1Team2RateBefore
			} else {
				plGmRtInc[_match.Player1Team2ID] = *_match.Player1Team2RateAfter - *_match.Player1Team2RateBefore
			}
		}
		// player2team1
		if _match.Player2Team1ID != nil && _match.Player2Team1RateBefore != _match.Player2Team1RateAfter {
			_, ok := plGmRtInc[*_match.Player2Team1ID]
			if ok {
				plGmRtInc[*_match.Player2Team1ID] += *_match.Player2Team1RateAfter - *_match.Player2Team1RateBefore
			} else {
				plGmRtInc[*_match.Player2Team1ID] = *_match.Player2Team1RateAfter - *_match.Player2Team1RateBefore
			}
		}
		// player2team2
		if _match.Player2Team2ID != nil && _match.Player2Team2RateBefore != _match.Player2Team2RateAfter {
			_, ok := plGmRtInc[*_match.Player2Team2ID]
			if ok {
				plGmRtInc[*_match.Player2Team2ID] += *_match.Player2Team2RateAfter - *_match.Player2Team2RateBefore
			} else {
				plGmRtInc[*_match.Player2Team2ID] = *_match.Player2Team2RateAfter - *_match.Player2Team2RateBefore
			}
		}
	}
	for playerID, value := range plGmRtInc {
		rateValue, err := s.rdbOperations.GetRatingByPlayerIDAndByLeagueID(logger, ctx, int64(playerID), leagueID)
		if err != nil {
			return err
		}

		rates[playerID] = int(rateValue) - value
	}

	// Map for keeping player ratings while going game calculation
	var matches []entities.NewMatch

	for _, match := range request.Matches {
		if match.ScoreTeam1 == 0 && match.ScoreTeam2 == 0 {
			continue
		}

		player1team1ID := match.Player1Team1Id
		player1team2ID := match.Player1Team2Id

		// Получаем рейтинги для всех игроков
		player1team1rate, err := s.getPlayerRating(ctx, logger, player1team1ID, leagueID, rates)
		if err != nil {
			return err
		}

		player1team2rate, err := s.getPlayerRating(ctx, logger, player1team2ID, leagueID, rates)
		if err != nil {
			return err
		}

		var player2team1rate, player2team2rate int
		var player2team1ID, player2team2ID int

		// Получаем рейтинг для player2team1 (если существует и > 0)
		if match.Player2Team1Id != nil {
			player2team1ID = *match.Player2Team1Id
			if player2team1ID > 0 {
				player2team1rate, err = s.getPlayerRating(ctx, logger, player2team1ID, leagueID, rates)
				if err != nil {
					return err
				}
			}
		}

		// Получаем рейтинг для player2team2 (если существует и > 0)
		if match.Player2Team2Id != nil {
			player2team2ID = *match.Player2Team2Id
			if player2team2ID > 0 {
				player2team2rate, err = s.getPlayerRating(ctx, logger, player2team2ID, leagueID, rates)
				if err != nil {
					return err
				}
			}
		}

		p1t1r := player1team1rate
		p1t2r := player1team2rate
		p2t1r := player2team1rate
		p2t2r := player2team2rate

		match.Player1Team1RateBefore = &p1t1r
		match.Player1Team2RateBefore = &p1t2r
		match.Player2Team1RateBefore = &p2t1r
		match.Player2Team2RateBefore = &p2t2r

		player1team1rateAfter, player1team2rateAfter, player2team1rateAfter, player2team2rateAfter, err := calculator.MatchRatingCalculation(
			ctx, match.ScoreTeam1, match.ScoreTeam2, player1team1rate, player1team2rate, player2team1rate, player2team2rate)
		if err != nil {
			return err
		}

		match.Player1Team1RateAfter = &player1team1rateAfter
		match.Player1Team2RateAfter = &player1team2rateAfter
		match.Player2Team1RateAfter = &player2team1rateAfter
		match.Player2Team2RateAfter = &player2team2rateAfter

		rates[match.Player1Team1Id] = player1team1rateAfter
		rates[match.Player1Team2Id] = player1team2rateAfter

		if player2team1ID > 0 {
			rates[*match.Player2Team1Id] = player2team1rateAfter
		}
		if player2team2ID > 0 {
			rates[*match.Player2Team2Id] = player2team2rateAfter
		}

		matches = append(matches, match)
	}

	request.Matches = matches

	newRates := make([]entities.Rating, 0, len(rates))
	for playerID, value := range rates {
		rate := &entities.Rating{
			PlayerID: int64(playerID),
			LeagueID: leagueID,
			Value:    int64(value),
		}

		newRates = append(newRates, *rate)
	}

	err = s.rwdbOperations.UpdateGame(logger, timeout, request, newRates)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Find(ctx context.Context, request entities.FindGameRequest) (entities.FindGameResponse, error) {
	logger := s.logger.With().Interface("service", "game.Find").Logger()
	timeout, cancel := context.WithTimeout(ctx, s.config.RDB.MaxIdleConnectionTimeout)
	defer cancel()

	games, err := s.rdbOperations.FindGames(logger, timeout, request)
	if err != nil {
		return entities.FindGameResponse{}, err
	}

	return entities.FindGameResponse{
		Games: games,
	}, nil
}

func (s *Service) UpdateFutureGame(ctx context.Context, request entities.UpdateFutureGameRequest) error {
	logger := s.logger.With().Str("service", "game.UpdateFutureGame").Logger()

	league, err := s.rdbOperations.GetLeagueById(logger, ctx, int64(request.LeagueID), nil)
	if err != nil {
		return err
	}

	team1, err := s.rdbOperations.GetTeamById(logger, ctx, int64(request.Team1ID), nil)
	if err != nil {
		return err
	}

	team2, err := s.rdbOperations.GetTeamById(logger, ctx, int64(request.Team2ID), nil)
	if err != nil {
		return err
	}

	// проверяем относится ли капитан к какой-либо из команд
	if request.Executor.Role.Name == constant.CaptainRole {
		err = s.validateCaptain(ctx, logger, request.Executor.ID, request.Executor.Team.ID, int64(request.Team1ID), int64(request.Team2ID))
		if err != nil {
			return err
		}
	}

	// проверяем город мастера по турнирам на соответствие городу лиги и команд
	if request.Executor.Role.Name == constant.TournamentMaster {
		master, err := s.rdbOperations.GetTournamentMasterByUserId(logger, ctx, request.Executor.ID, &s.config.RDB)
		if err != nil {
			return err
		}

		if (team1.CityId == nil || *team1.CityId != master.City.ID) || (team2.CityId == nil || *team2.CityId != master.City.ID) {
			err = errors.New("город команды и город мастера по турнирам не совпадают")
			logger.Error().Err(err).Msg("master city not equal league city")
			return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}

		if league.CityID != master.City.ID {
			err = errors.New("город лиги и город мастера по турнирам не совпадают")
			logger.Error().Err(err).Msg("master city not equal league city")
			return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}
	}

	tx, err := s.rwdbOperations.BeginTx(ctx, logger)
	if err != nil {
		return err
	}

	err = s.rwdbOperations.UpdateFutureGame(logger, ctx, request, tx)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		tx.Rollback(ctx)
		return err
	}

	return nil
}

func (s *Service) GetGamesYears(ctx context.Context) (entities.GetGamesYearsResponse, error) {
	logger := s.logger.With().Interface("service", "game.GetGamesYears").Logger()
	timeout, cancel := context.WithTimeout(ctx, s.config.RDB.MaxIdleConnectionTimeout)
	defer cancel()

	resp, err := s.rdbOperations.GetGamesYears(logger, timeout)
	if err != nil {
		return entities.GetGamesYearsResponse{}, err
	}
	return resp, nil
}

func (s *Service) GetGameList(ctx context.Context, request entities.GetGameListRequest) (entities.GetGameListResponse, error) {
	logger := s.logger.With().Str("service", "game.GetGameList").Logger()
	timeout, cancel := context.WithTimeout(ctx, s.config.RDB.MaxIdleConnectionTimeout)
	defer cancel()

	games, err := s.rdbOperations.GetGameList(logger, timeout, request)
	if err != nil {
		return entities.GetGameListResponse{}, err
	}

	return entities.GetGameListResponse{Games: games}, nil
}

func (s *Service) GetComingGames(ctx context.Context) (entities.GetComingGamesResponse, error) {
	logger := s.logger.With().Interface("service", "game.GetComingGames").Logger()
	timeout, cancel := context.WithTimeout(ctx, s.config.RDB.MaxIdleConnectionTimeout)
	defer cancel()

	games, err := s.rdbOperations.GetComingGames(logger, timeout)
	if err != nil {
		return entities.GetComingGamesResponse{}, err
	}

	return entities.GetComingGamesResponse{ComingGames: games}, nil
}

func (s *Service) GetFutureGames(ctx context.Context, cityID int) (entities.GetFutureGamesResponse, error) {
	logger := s.logger.With().Str("service", "game.GetFutureGames").Logger()
	timeout, cancel := context.WithTimeout(ctx, s.config.RDB.MaxIdleConnectionTimeout)
	defer cancel()

	games, err := s.rdbOperations.GetFutureGames(logger, timeout, cityID)
	if err != nil {
		return entities.GetFutureGamesResponse{}, err
	}

	return entities.GetFutureGamesResponse{Games: games}, nil
}

func (s *Service) CreateFutureGame(ctx context.Context, request entities.CreateFutureGameRequest) (entities.CreateFutureGameResponse, error) {
	logger := s.logger.With().Str("service", "game.CreateFutureGame").Logger()

	league, err := s.rdbOperations.GetLeagueById(logger, ctx, int64(request.LeagueID), nil)
	if err != nil {
		return entities.CreateFutureGameResponse{}, err
	}

	team1, err := s.rdbOperations.GetTeamById(logger, ctx, int64(request.Team1ID), nil)
	if err != nil {
		return entities.CreateFutureGameResponse{}, err
	}

	team2, err := s.rdbOperations.GetTeamById(logger, ctx, int64(request.Team2ID), nil)
	if err != nil {
		return entities.CreateFutureGameResponse{}, err
	}

	// проверяем относится ли капитан к какой-либо из команд
	if request.Creator.Role.Name == constant.CaptainRole {
		err = s.validateCaptain(ctx, logger, request.Creator.ID, request.Creator.Team.ID, int64(request.Team1ID), int64(request.Team2ID))
		if err != nil {
			return entities.CreateFutureGameResponse{}, err
		}
	}

	// проверяем город мастера по турнирам на соответствие городу лиги и команд
	if request.Creator.Role.Name == constant.TournamentMaster {
		master, err := s.rdbOperations.GetTournamentMasterByUserId(logger, ctx, request.Creator.ID, &s.config.RDB)
		if err != nil {
			return entities.CreateFutureGameResponse{}, err
		}

		if (team1.CityId == nil || *team1.CityId != master.City.ID) || (team2.CityId == nil || *team2.CityId != master.City.ID) {
			err = errors.New("город команды и город мастера по турнирам не совпадают")
			logger.Error().Err(err).Msg("master city not equal league city")
			return entities.CreateFutureGameResponse{}, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}

		if league.CityID != master.City.ID {
			err = errors.New("город лиги и город мастера по турнирам не совпадают")
			logger.Error().Err(err).Msg("master city not equal league city")
			return entities.CreateFutureGameResponse{}, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}

		if master.City.ID != int64(request.CityID) {
			err = errors.New("город в запросе и город мастера по турнирам не совпадают")
			logger.Error().Err(err).Msg("master city not equal request city")
			return entities.CreateFutureGameResponse{}, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}
	}

	if league.CityID != int64(request.CityID) {
		err = errors.New("город в запросе и город лиги не совпадают")
		logger.Error().Err(err).Msg("request city not equal league city")
		return entities.CreateFutureGameResponse{}, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	tx, err := s.rwdbOperations.BeginTx(ctx, logger)
	if err != nil {
		return entities.CreateFutureGameResponse{}, err
	}

	id, err := s.rwdbOperations.CreateFutureGame(logger, ctx, request, tx)
	if err != nil {
		tx.Rollback(ctx)
		return entities.CreateFutureGameResponse{}, err
	}

	if err = tx.Commit(ctx); err != nil {
		tx.Rollback(ctx)
		return entities.CreateFutureGameResponse{}, err
	}

	return entities.CreateFutureGameResponse{ID: id}, nil
}

func (s *Service) GetTeamGames(ctx context.Context, teamID int) (entities.GetTeamGamesResponse, error) {
	logger := s.logger.With().Interface("service", "game.GetTeamGames").Logger()
	timeout, cancel := context.WithTimeout(ctx, s.config.RDB.MaxIdleConnectionTimeout)
	defer cancel()

	games, err := s.rdbOperations.GetTeamGames(logger, timeout, teamID)
	if err != nil {
		return entities.GetTeamGamesResponse{}, nil
	}

	return entities.GetTeamGamesResponse{Games: games}, nil
}

func (s *Service) DeleteFutureGame(ctx context.Context, req entities.DeleteFutureGameRequest) error {
	logger := s.logger.With().Str("service", "game.DeleteFutureGame").Logger()

	game, err := s.rdbOperations.GetGameById(logger, ctx, int(req.ID), nil)
	if err != nil {
		return err
	}

	league := &entities.League{}
	if game.LeagueID != nil {
		*league, err = s.rdbOperations.GetLeagueById(logger, ctx, *game.LeagueID, nil)
		if err != nil {
			return err
		}
	}

	if req.Executor.Role.Name == constant.CaptainRole {
		err = s.validateCaptain(ctx, logger, req.Executor.ID, req.Executor.Team.ID, game.Team1ID, game.Team2ID)
		if err != nil {
			return err
		}
	}

	if req.Executor.Role.Name == constant.TournamentMaster {
		master, err := s.rdbOperations.GetTournamentMasterByUserId(logger, ctx, req.Executor.ID, &s.config.RDB)
		if err != nil {
			return err
		}

		if game.CityID != master.City.ID {
			err = errors.New("город игры и город мастера по турнирам не совпадают")
			logger.Error().Err(err).Msg("master city not equal game city")
			return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}

		if league.CityID != master.City.ID {
			err = errors.New("город лиги игры и город мастера по турнирам не совпадают")
			logger.Error().Err(err).Msg("master city not equal league city")
			return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}
	}

	if game.Date != nil && time.Now().After(*game.Date) {
		err = errors.New("сыгранная игра не может быть удалена")
		logger.Error().Err(err).Msg("game was played")
		return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	tx, err := s.rwdbOperations.BeginTx(ctx, logger)
	if err != nil {
		return err
	}

	err = s.rwdbOperations.DeleteFutureGame(logger, ctx, req, tx)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		tx.Rollback(ctx)
		return err
	}

	return nil
}

/* tournament */

func (s *Service) CreateFutureTournamentGame(ctx context.Context, request *entities.CreateFutureTournamentGameRequest) (int64, error) {
	logger := s.logger.With().Str("service", "game.CreateFutureTournamentGame").Logger()

	tournament, err := s.rdbOperations.GetTournamentById(logger, ctx, request.TournamentID, &s.config.RDB)
	if err != nil {
		return 0, err
	}

	request.CityID = tournament.CityID

	// проверяем относится ли капитан к какой-либо из команд
	if request.Creator.Role.Name == constant.CaptainRole {
		if err = s.validateCaptain(ctx, logger, request.Creator.ID, request.Creator.Team.ID, request.Team1ID, request.Team2ID); err != nil {
			return 0, err
		}
	}

	// проверяем город мастера по турнирам на соответствие городу турнира
	if request.Creator.Role.Name == constant.TournamentMaster {
		if err = s.validateTournamentMaster(ctx, logger, &tournament, request.Creator.ID); err != nil {
			return 0, err
		}
	}

	// проверяем соответствует ли столоместо городу турнира, если оно есть
	if request.PlaceID != nil {
		place, err := s.rdbOperations.GetPlaceByID(logger, ctx, *request.PlaceID, &s.config.RDB)
		if err != nil {
			return 0, err
		}
		place.Bar.City, err = s.rdbOperations.GetCityByBarId(logger, ctx, place.Bar.ID, &s.config.RDB)
		if err != nil {
			return 0, err
		}
		if place.Bar.City.ID != tournament.CityID {
			err = errors.New("город турнира и города столоместа не совпадают")
			logger.Error().Err(err).Msg("place city not equal tournament city")
			return 0, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}
	}

	// проверяем относится ли входящий id этапа к этапам турнира и не является ли завершенным
	if err = checkTournamentStage(tournament.Stages, request.StageID); err != nil {
		logger.Error().Err(err).Msg("stage not found")
		return 0, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	// проверяем все ли команды участвуют в этом турнире
	if (slices.Contains(tournament.TeamIDs, request.Team1ID) && slices.Contains(tournament.TeamIDs, request.Team2ID)) == false {
		err = errors.New("команда не участвует в турнире")
		logger.Error().Err(err).Msg("team not in tournament")
		return 0, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	// команды должны быть разные
	if request.Team1ID == request.Team2ID {
		err = errors.New("id команд не могут быть одинаковыми")
		logger.Error().Err(err).Msg("team1Id equal team2Id")
		return 0, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	if tournament.TypeID == constant.RegularTournamentTypeID || tournament.TypeID == constant.RegularOneVsOneTournamentTypeID {
		err = checkCreateRegularTournamentGame(logger, request.IsTiebreak)
		if err != nil {
			return 0, err
		}
	} else {
		// todo другие типы турниров
		return 0, errors.New("not implemented")
	}

	id, err := s.rwdbOperations.CreateFutureTournamentGame(logger, ctx, request, &s.config.RWDB)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *Service) UpdateFutureTournamentGame(ctx context.Context, request *entities.UpdateFutureTournamentGameRequest) error {
	logger := s.GetLogger().With().Str("service", "UpdateFutureTournamentGame").Logger()

	game, err := s.rdbOperations.GetTournamentGame(logger, ctx, request.GameID, &s.config.RDB)
	if err != nil {
		return err
	}

	stage, err := s.rdbOperations.GetTournamentStage(logger, ctx, game.StageID, &s.config.RDB)
	if err != nil {
		return err
	}

	tournament, err := s.rdbOperations.GetTournamentById(logger, ctx, stage.TournamentID, &s.config.RDB)
	if err != nil {
		return err
	}

	updReq := buildNewRequest(request, &game)

	// проверяем относится ли капитан к какой-либо из команд
	if request.Executor.Role.Name == constant.CaptainRole {
		if err = s.validateCaptain(ctx, logger, updReq.Executor.ID, updReq.Executor.Team.ID, *updReq.Team1ID, *updReq.Team2ID); err != nil {
			return err
		}
	}

	// проверяем город мастера по турнирам на соответствие городу турнира
	if request.Executor.Role.Name == constant.TournamentMaster {
		if err = s.validateTournamentMaster(ctx, logger, &tournament, updReq.Executor.ID); err != nil {
			return err
		}
	}

	// проверяем соответствует ли столоместо городу турнира
	if updReq.PlaceID != nil {
		place, err := s.rdbOperations.GetPlaceByID(logger, ctx, *updReq.PlaceID, &s.config.RDB)
		if err != nil {
			return err
		}
		place.Bar.City, err = s.rdbOperations.GetCityByBarId(logger, ctx, place.Bar.ID, &s.config.RDB)
		if err != nil {
			return err
		}
		if place.Bar.City.ID != tournament.CityID {
			err = errors.New("город турнира и города столоместа не совпадают")
			logger.Error().Err(err).Msg("place city not equal tournament city")
			return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}
	}

	// проверяем все ли команды участвуют в этом турнире
	if (slices.Contains(tournament.TeamIDs, *updReq.Team1ID) && slices.Contains(tournament.TeamIDs, *updReq.Team2ID)) == false {
		err = errors.New("команда не участвует в турнире")
		logger.Error().Err(err).Msg("team not in tournament")
		return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	if *updReq.Team1ID == *updReq.Team2ID {
		err = errors.New("id команд не могут быть одинаковыми")
		logger.Error().Err(err).Msg("team1Id equal team2Id")
		return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	if tournament.TypeID == constant.RegularTournamentTypeID || tournament.TypeID == constant.RegularOneVsOneTournamentTypeID {
		err = checkUpdateRegularTournamentGame(logger, *updReq.Team1ID, *updReq.Team2ID, &game)
		if err != nil {
			return err
		}
	} else {
		// todo другие типы турниров
		return errors.New("not implemented")
	}

	err = s.rwdbOperations.UpdateFutureTournamentGame(logger, ctx, updReq, &s.config.RWDB)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) DeleteFutureTournamentGame(ctx context.Context, request *entities.DeleteFutureTournamentGameRequest) error {
	logger := s.GetLogger().With().Str("service", "DeleteFutureTournamentGame").Logger()

	game, err := s.rdbOperations.GetTournamentGame(logger, ctx, request.GameID, &s.config.RDB)
	if err != nil {
		return err
	}

	stage, err := s.rdbOperations.GetTournamentStage(logger, ctx, game.StageID, &s.config.RDB)
	if err != nil {
		return err
	}

	tournament, err := s.rdbOperations.GetTournamentById(logger, ctx, stage.TournamentID, &s.config.RDB)
	if err != nil {
		return err
	}

	// проверяем относится ли капитан к какой-либо из команд
	if request.Executor.Role.Name == constant.CaptainRole {
		if err = s.validateCaptain(ctx, logger, request.Executor.ID, request.Executor.Team.ID, game.Team1ID, game.Team2ID); err != nil {
			return err
		}
	}

	// проверяем город мастера по турнирам на соответствие городу турнира
	if request.Executor.Role.Name == constant.TournamentMaster {
		if err = s.validateTournamentMaster(ctx, logger, &tournament, request.Executor.ID); err != nil {
			return err
		}
	}

	// не позволяем удалять игры у завершенных этапов
	if stage.IsFinished == true {
		err = errors.New("у завершенного этапа нельзя удалять игры")
		logger.Error().Err(err).Msg("stage is finished")
		return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	matches, err := s.rdbOperations.GetMatchListByGameID(ctx, logger, request.GameID, &s.config.RDB)
	if err != nil {
		return err
	}
	// нельзя удалить игру с прошедшими матчами
	if len(matches) > 0 {
		for _, match := range matches {
			if time.Now().After(match.Date) {
				err = errors.New("игру с начатым матчем нельзя удалять")
				logger.Error().Err(err).Msg("match is started")
				return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
			}
		}
	}

	if tournament.TypeID == constant.RegularTournamentTypeID || tournament.TypeID == constant.RegularOneVsOneTournamentTypeID {
		err = checkDeleteRegularTournamentGame(logger, &game)
		if err != nil {
			return err
		}
	} else {
		// todo другие типы турниров
		return errors.New("not implemented")
	}

	err = s.rwdbOperations.DeleteGame(logger, ctx, game.ID, &s.config.RWDB)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) CreatePlayedTournamentGame(ctx context.Context, request *entities.CreatePlayedTournamentGameRequest) (int64, error) {
	logger := s.logger.With().Str("service", "game.CreatePlayedTournamentGame").Logger()

	tournament, err := s.rdbOperations.GetTournamentById(logger, ctx, request.TournamentID, &s.config.RDB)
	if err != nil {
		return 0, err
	}

	request.CityID = tournament.CityID

	// проверяем относится ли капитан к какой-либо из команд
	if request.Creator.Role.Name == constant.CaptainRole {
		if err = s.validateCaptain(ctx, logger, request.Creator.ID, request.Creator.Team.ID, request.Team1ID, request.Team2ID); err != nil {
			return 0, err
		}
	}

	// проверяем город мастера по турнирам на соответствие городу турнира
	if request.Creator.Role.Name == constant.TournamentMaster {
		if err = s.validateTournamentMaster(ctx, logger, &tournament, request.Creator.ID); err != nil {
			return 0, err
		}
	}

	// проверяем соответствует ли столоместо городу турнира
	place, err := s.rdbOperations.GetPlaceByID(logger, ctx, request.PlaceID, &s.config.RDB)
	if err != nil {
		return 0, err
	}
	place.Bar.City, err = s.rdbOperations.GetCityByBarId(logger, ctx, place.Bar.ID, &s.config.RDB)
	if err != nil {
		return 0, err
	}
	if place.Bar.City.ID != tournament.CityID {
		err = errors.New("город турнира и города столоместа не совпадают")
		logger.Error().Err(err).Msg("place city not equal tournament city")
		return 0, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	// проверяем относится ли входящий id этапа к этапам турнира и не является ли завершенным
	if err = checkTournamentStage(tournament.Stages, request.StageID); err != nil {
		logger.Error().Err(err).Msg("stage not found")
		return 0, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	// проверяем все ли команды участвуют в этом турнире
	if (slices.Contains(tournament.TeamIDs, request.Team1ID) && slices.Contains(tournament.TeamIDs, request.Team2ID)) == false {
		err = errors.New("команда не участвует в турнире")
		logger.Error().Err(err).Msg("team not in tournament")
		return 0, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	// команды должны быть разные
	if request.Team1ID == request.Team2ID {
		err = errors.New("id команд не могут быть одинаковыми")
		logger.Error().Err(err).Msg("team1Id equal team2Id")
		return 0, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	if request.TechLooseTeamID != nil && (*request.TechLooseTeamID != request.Team1ID && *request.TechLooseTeamID != request.Team2ID) {
		err = errors.New("нельзя присудить техническое поражение команде не участвующей в игре")
		logger.Error().Err(err).Msg("techLooseTeamID not equal team1Id/team2Id")
		return 0, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	for _, m := range request.Matches {
		if request.Team1ID != int64(m.Team1ID) || request.Team2ID != int64(m.Team2ID) {
			err = errors.New("команда матча не учствует в игре")
			logger.Error().Err(err).Msg("match team not equal game team")
			return 0, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}

		if m.Date.Year() != request.Date.Year() || m.Date.Month() != request.Date.Month() || m.Date.Day() != request.Date.Day() {
			err = errors.New("даты игры и матча не совпадают")
			logger.Error().Err(err).Msg("dates not equal")
			return 0, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}
	}

	if tournament.TypeID == constant.RegularTournamentTypeID {
		err = checkCreateRegularTournamentGame(logger, request.IsTiebreak)
		if err != nil {
			return 0, err
		}

	} else if tournament.TypeID == constant.RegularOneVsOneTournamentTypeID {
		err = checkCreateRegularTournamentGame(logger, request.IsTiebreak)
		if err != nil {
			return 0, err
		}

		for _, m := range request.Matches {
			if m.Player2Team1Id != nil || m.Player2Team2Id != nil {
				err = errors.New("в турнире типа Regular.OneVsOne у команды не может быть второго игрока")
				logger.Error().Err(err).Msg("two players in OneVsOne")
				return 0, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
			}
		}

	} else {
		// todo другие типы турниров
		return 0, errors.New("not implemented")
	}

	// Map for keeping player ratings while going game calculation
	rates := make(map[int]int)

	var matches []entities.GamesMatch

	if len(request.Matches) > 0 {
		for _, match := range request.Matches {
			// не позволяем дату матча ставить раньше даты игры. Возможно стоит проверить на равенство по дню?
			if request.Date.After(match.Date) {
				err = errors.New("дата матча не может быть раньше даты игры")
				logger.Error().Err(err).Msg("game date after match date")
				return 0, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
			}

			// проверка соответствия команд
			if request.Team1ID != int64(match.Team1ID) || request.Team2ID != int64(match.Team2ID) {
				err = errors.New("команды матчей и игр должны совпадать")
				logger.Error().Err(err).Msg("game teams not equal match teams")
				return 0, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
			}

			if match.ScoreTeam1 == 0 && match.ScoreTeam2 == 0 {
				continue
			}

			player1team1ID := match.Player1Team1Id
			player1team2ID := match.Player1Team2Id

			// Получаем рейтинги для всех игроков
			player1team1rate, err := s.getPlayerRating(ctx, logger, player1team1ID, tournament.ID, rates)
			if err != nil {
				return 0, err
			}

			player1team2rate, err := s.getPlayerRating(ctx, logger, player1team2ID, tournament.ID, rates)
			if err != nil {
				return 0, err
			}

			var player2team1rate, player2team2rate int
			var player2team1ID, player2team2ID int

			// Получаем рейтинг для player2team1 (если существует и > 0)
			if match.Player2Team1Id != nil {
				player2team1ID = *match.Player2Team1Id
				if player2team1ID > 0 {
					player2team1rate, err = s.getPlayerRating(ctx, logger, player2team1ID, tournament.ID, rates)
					if err != nil {
						return 0, err
					}
				}
			}

			// Получаем рейтинг для player2team2 (если существует и > 0)
			if match.Player2Team2Id != nil {
				player2team2ID = *match.Player2Team2Id
				if player2team2ID > 0 {
					player2team2rate, err = s.getPlayerRating(ctx, logger, player2team2ID, tournament.ID, rates)
					if err != nil {
						return 0, err
					}
				}
			}

			p1t1r := player1team1rate
			p1t2r := player1team2rate
			p2t1r := player2team1rate
			p2t2r := player2team2rate

			match.Player1Team1RateBefore = &p1t1r
			match.Player1Team2RateBefore = &p1t2r
			match.Player2Team1RateBefore = &p2t1r
			match.Player2Team2RateBefore = &p2t2r

			player1team1rateAfter, player1team2rateAfter, player2team1rateAfter, player2team2rateAfter, err := calculator.MatchRatingCalculation(
				ctx, match.ScoreTeam1, match.ScoreTeam2, player1team1rate, player1team2rate, player2team1rate, player2team2rate)
			if err != nil {
				return 0, err
			}

			match.Player1Team1RateAfter = &player1team1rateAfter
			match.Player1Team2RateAfter = &player1team2rateAfter
			match.Player2Team1RateAfter = &player2team1rateAfter
			match.Player2Team2RateAfter = &player2team2rateAfter

			rates[match.Player1Team1Id] = player1team1rateAfter
			rates[match.Player1Team2Id] = player1team2rateAfter
			if player2team1ID > 0 {
				rates[*match.Player2Team1Id] = player2team1rateAfter
			}
			if player2team2ID > 0 {
				rates[*match.Player2Team2Id] = player2team2rateAfter
			}

			matches = append(matches, match)
		}

		request.Matches = matches
	}

	id, err := s.rwdbOperations.CreatePlayedTournamentGame(logger, ctx, request, &s.config.RWDB)
	if err != nil {
		return 0, err
	}

	for _, m := range request.Matches {
		_, err = s.rwdbOperations.CreateGameMatch(logger, ctx, m, id, &s.config.RWDB)
		if err != nil {
			return 0, err
		}
	}

	for playerID, value := range rates {
		err = s.rwdbOperations.CreateTournamentRating(logger, ctx, int64(playerID), int64(value), request.TournamentID, &s.config.RWDB)
		if err != nil {
			return 0, err
		}
	}

	return id, nil
}

func (s *Service) UpdatePlayedTournamentGame(ctx context.Context, request *entities.UpdatePlayedTournamentGameRequest) error {
	logger := s.GetLogger().With().Str("service", "UpdatePlayedTournamentGame").Logger()

	game, err := s.rdbOperations.GetTournamentGame(logger, ctx, request.GameID, &s.config.RDB)
	if err != nil {
		return err
	}

	stage, err := s.rdbOperations.GetTournamentStage(logger, ctx, game.StageID, &s.config.RDB)
	if err != nil {
		return err
	}

	request.StageID = stage.ID

	tournament, err := s.rdbOperations.GetTournamentById(logger, ctx, stage.TournamentID, &s.config.RDB)
	if err != nil {
		return err
	}

	// проверяем относится ли капитан к какой-либо из команд
	if request.Executor.Role.Name == constant.CaptainRole {
		if err = s.validateCaptain(ctx, logger, request.Executor.ID, request.Executor.Team.ID, request.Team1ID, request.Team2ID); err != nil {
			return err
		}
	}

	// проверяем город мастера по турнирам на соответствие городу турнира
	if request.Executor.Role.Name == constant.TournamentMaster {
		if err = s.validateTournamentMaster(ctx, logger, &tournament, request.Executor.ID); err != nil {
			return err
		}
	}

	// проверяем соответствует ли столоместо городу турнира
	place, err := s.rdbOperations.GetPlaceByID(logger, ctx, request.PlaceID, &s.config.RDB)
	if err != nil {
		return err
	}
	place.Bar.City, err = s.rdbOperations.GetCityByBarId(logger, ctx, place.Bar.ID, &s.config.RDB)
	if err != nil {
		return err
	}
	if place.Bar.City.ID != tournament.CityID {
		err = errors.New("город турнира и города столоместа не совпадают")
		logger.Error().Err(err).Msg("place city not equal tournament city")
		return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	// проверяем все ли команды участвуют в этом турнире
	if (slices.Contains(tournament.TeamIDs, request.Team1ID) && slices.Contains(tournament.TeamIDs, request.Team2ID)) == false {
		err = errors.New("команда не участвует в турнире")
		logger.Error().Err(err).Msg("team not in tournament")
		return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	// проверяем что команды разные
	if request.Team1ID == request.Team2ID {
		err = errors.New("id команд не могут быть одинаковыми")
		logger.Error().Err(err).Msg("team1Id equal team2Id")
		return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	if tournament.TypeID == constant.RegularTournamentTypeID {
		err = checkUpdateRegularTournamentGame(logger, request.Team1ID, request.Team2ID, &game)
		if err != nil {
			return err
		}
	} else if tournament.TypeID == constant.RegularOneVsOneTournamentTypeID {
		err = checkUpdateRegularTournamentGame(logger, request.Team1ID, request.Team2ID, &game)
		if err != nil {
			return err
		}

		for _, m := range request.Matches {
			if m.Player2Team1Id != nil || m.Player2Team2Id != nil {
				err = errors.New("в турнире типа Regular.OneVsOne у команды не может быть второго игрока")
				logger.Error().Err(err).Msg("two players in OneVsOne")
				return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
			}
		}
	} else {
		// todo другие типы турниров
		return errors.New("not implemented")
	}

	newRates := make(map[int]int)
	oldRates := make(map[int]int)

	// прошедшие матчи
	matchesV2, err := s.rdbOperations.GetMatchListByGameID(ctx, logger, request.GameID, &s.config.RDB)
	if err != nil {
		return err
	}

	// заполнение старого рейтинга
	for _, matchV2 := range matchesV2 {
		// player1team1
		if matchV2.Player1Team1RateBefore != matchV2.Player1Team1RateAfter {
			_, ok := oldRates[matchV2.Player1Team1ID]
			if ok {
				oldRates[matchV2.Player1Team1ID] += *matchV2.Player1Team1RateAfter - *matchV2.Player1Team1RateBefore
			} else {
				oldRates[matchV2.Player1Team1ID] = *matchV2.Player1Team1RateAfter - *matchV2.Player1Team1RateBefore
			}
		}
		// player1team2
		if matchV2.Player1Team2RateBefore != matchV2.Player1Team2RateAfter {
			_, ok := oldRates[matchV2.Player1Team2ID]
			if ok {
				oldRates[matchV2.Player1Team2ID] += *matchV2.Player1Team2RateAfter - *matchV2.Player1Team2RateBefore
			} else {
				oldRates[matchV2.Player1Team2ID] = *matchV2.Player1Team2RateAfter - *matchV2.Player1Team2RateBefore
			}
		}
		// player2team1
		if matchV2.Player2Team1ID != nil && matchV2.Player2Team1RateBefore != matchV2.Player2Team1RateAfter {
			_, ok := oldRates[*matchV2.Player2Team1ID]
			if ok {
				oldRates[*matchV2.Player2Team1ID] += *matchV2.Player2Team1RateAfter - *matchV2.Player2Team1RateBefore
			} else {
				oldRates[*matchV2.Player2Team1ID] = *matchV2.Player2Team1RateAfter - *matchV2.Player2Team1RateBefore
			}
		}
		// player2team2
		if matchV2.Player2Team2ID != nil && matchV2.Player2Team2RateBefore != matchV2.Player2Team2RateAfter {
			_, ok := oldRates[*matchV2.Player2Team2ID]
			if ok {
				oldRates[(*matchV2.Player2Team2ID)] += *matchV2.Player2Team2RateAfter - *matchV2.Player2Team2RateBefore
			} else {
				oldRates[*matchV2.Player2Team2ID] = *matchV2.Player2Team2RateAfter - *matchV2.Player2Team2RateBefore
			}
		}
	}

	for playerID, value := range oldRates {
		rateValue, err := s.rdbOperations.GetPlayerRatingByTournamentId(logger, ctx, int64(playerID), tournament.ID)
		if err != nil {
			return err
		}

		newRates[playerID] = int(rateValue) - value
	}

	var matches []entities.NewMatch

	// пересчет рейтинга
	for _, match := range request.Matches {
		if match.ScoreTeam1 == 0 && match.ScoreTeam2 == 0 {
			continue
		}

		player1team1ID := match.Player1Team1Id
		player1team2ID := match.Player1Team2Id

		// Получаем рейтинги для всех игроков
		player1team1rate, err := s.getPlayerRating(ctx, logger, player1team1ID, tournament.ID, newRates)
		if err != nil {
			return err
		}

		player1team2rate, err := s.getPlayerRating(ctx, logger, player1team2ID, tournament.ID, newRates)
		if err != nil {
			return err
		}

		var player2team1rate, player2team2rate int
		var player2team1ID, player2team2ID int

		// Получаем рейтинг для player2team1 (если существует и > 0)
		if match.Player2Team1Id != nil {
			player2team1ID = *match.Player2Team1Id
			if player2team1ID > 0 {
				player2team1rate, err = s.getPlayerRating(ctx, logger, player2team1ID, tournament.ID, newRates)
				if err != nil {
					return err
				}
			}
		}

		// Получаем рейтинг для player2team2 (если существует и > 0)
		if match.Player2Team2Id != nil {
			player2team2ID = *match.Player2Team2Id
			if player2team2ID > 0 {
				player2team2rate, err = s.getPlayerRating(ctx, logger, player2team2ID, tournament.ID, newRates)
				if err != nil {
					return err
				}
			}
		}

		p1t1r := player1team1rate
		p1t2r := player1team2rate
		p2t1r := player2team1rate
		p2t2r := player2team2rate

		match.Player1Team1RateBefore = &p1t1r
		match.Player1Team2RateBefore = &p1t2r
		match.Player2Team1RateBefore = &p2t1r
		match.Player2Team2RateBefore = &p2t2r

		player1team1rateAfter, player1team2rateAfter, player2team1rateAfter, player2team2rateAfter, err := calculator.MatchRatingCalculation(
			ctx, match.ScoreTeam1, match.ScoreTeam2, player1team1rate, player1team2rate, player2team1rate, player2team2rate)
		if err != nil {
			return err
		}

		match.Player1Team1RateAfter = &player1team1rateAfter
		match.Player1Team2RateAfter = &player1team2rateAfter
		match.Player2Team1RateAfter = &player2team1rateAfter
		match.Player2Team2RateAfter = &player2team2rateAfter

		newRates[match.Player1Team1Id] = player1team1rateAfter
		newRates[match.Player1Team2Id] = player1team2rateAfter

		if player2team1ID > 0 {
			newRates[*match.Player2Team1Id] = player2team1rateAfter
		}
		if player2team2ID > 0 {
			newRates[*match.Player2Team2Id] = player2team2rateAfter
		}

		matches = append(matches, match)
	}

	request.Matches = matches

	// список id-шников матчей НЕ подлежащих удалению
	var matchIds = make([]int, 0)
	for _, match := range request.Matches {
		if match.ID != nil {
			matchIds = append(matchIds, *match.ID)
		}
	}

	// удаление всех матчей игры кроме входящих
	err = s.rwdbOperations.DeleteOldGameMatches(logger, ctx, request.GameID, matchIds, &s.config.RWDB)
	if err != nil {
		return err
	}

	// обновление существующих матчей и создание новых
	for _, match := range request.Matches {
		if match.Player2Team1Id != nil && *match.Player2Team1Id == 0 {
			match.Player2Team1Id = nil
		}
		if match.Player2Team2Id != nil && *match.Player2Team2Id == 0 {
			match.Player2Team2Id = nil
		}

		if match.ID == nil {
			err = s.rwdbOperations.CreateNewMatch(logger, ctx, request.GameID, &match, &s.config.RWDB)
			if err != nil {
				return err
			}
		} else {
			err = s.rwdbOperations.UpdateOldMatch(logger, ctx, request.GameID, &match, &s.config.RWDB)
			if err != nil {
				return err
			}
		}
	}

	// обновление рейтингов
	for playerID, value := range newRates {
		err = s.rwdbOperations.CreateTournamentRating(logger, ctx, int64(playerID), int64(value), tournament.ID, &s.config.RWDB)
		if err != nil {
			return err
		}
	}

	// обновление игры
	err = s.rwdbOperations.UpdatePlayedTournamentGame(logger, ctx, request, &s.config.RWDB)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) DeletePlayedTournamentGame(ctx context.Context, request *entities.DeletePlayedTournamentGameRequest) error {
	logger := s.GetLogger().With().Str("service", "DeletePlayedTournamentGame").Logger()

	game, err := s.rdbOperations.GetTournamentGame(logger, ctx, request.GameID, &s.config.RDB)
	if err != nil {
		return err
	}

	stage, err := s.rdbOperations.GetTournamentStage(logger, ctx, game.StageID, &s.config.RDB)
	if err != nil {
		return err
	}

	tournament, err := s.rdbOperations.GetTournamentById(logger, ctx, stage.TournamentID, &s.config.RDB)
	if err != nil {
		return err
	}

	// проверяем относится ли капитан к какой-либо из команд
	if request.Executor.Role.Name == constant.CaptainRole {
		if err = s.validateCaptain(ctx, logger, request.Executor.ID, request.Executor.Team.ID, game.Team1ID, game.Team2ID); err != nil {
			return err
		}
	}

	// проверяем город мастера по турнирам на соответствие городу турнира
	if request.Executor.Role.Name == constant.TournamentMaster {
		if err = s.validateTournamentMaster(ctx, logger, &tournament, request.Executor.ID); err != nil {
			return err
		}
	}

	// не позволяем удалять игры у завршенных этапов
	if stage.IsFinished == true {
		err = errors.New("у завершенного этапа нельзя удалять игры")
		logger.Error().Err(err).Msg("stage is finished")
		return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	if tournament.TypeID == constant.RegularTournamentTypeID {
		err = checkDeleteRegularTournamentGame(logger, &game)
		if err != nil {
			return err
		}
	} else {
		// todo другие типы турниров
		return errors.New("not implemented")
	}

	matches, err := s.rdbOperations.GetMatchListByGameID(ctx, logger, request.GameID, &s.config.RDB)
	if err != nil {
		return err
	}
	// нельзя удалить игру с прошедшими матчами
	if len(matches) > 0 {
		for _, match := range matches {
			if time.Now().After(match.Date) {
				err = errors.New("игру с начатым матчем нельзя удалять")
				logger.Error().Err(err).Msg("match is started")
				return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
			}
		}
	}

	oldRate := make(map[int]int)

	for _, match := range matches {
		// player1team1
		if match.Player1Team1RateBefore != match.Player1Team1RateAfter {
			_, ok := oldRate[match.Player1Team1ID]
			if ok {
				oldRate[match.Player1Team1ID] += *match.Player1Team1RateAfter - *match.Player1Team1RateBefore
			} else {
				oldRate[match.Player1Team1ID] = *match.Player1Team1RateAfter - *match.Player1Team1RateBefore
			}
		}
		// player1team2
		if match.Player1Team2RateBefore != match.Player1Team2RateAfter {
			_, ok := oldRate[match.Player1Team2ID]
			if ok {
				oldRate[match.Player1Team2ID] += *match.Player1Team2RateAfter - *match.Player1Team2RateBefore
			} else {
				oldRate[match.Player1Team2ID] = *match.Player1Team2RateAfter - *match.Player1Team2RateBefore
			}
		}
		// player2team1
		if match.Player2Team1ID != nil && match.Player2Team1RateBefore != match.Player2Team1RateAfter {
			_, ok := oldRate[*match.Player2Team1ID]
			if ok {
				oldRate[*match.Player2Team1ID] += *match.Player2Team1RateAfter - *match.Player2Team1RateBefore
			} else {
				oldRate[*match.Player2Team1ID] = *match.Player2Team1RateAfter - *match.Player2Team1RateBefore
			}
		}
		// player2team2
		if match.Player2Team2ID != nil && match.Player2Team2RateBefore != match.Player2Team2RateAfter {
			_, ok := oldRate[*match.Player2Team2ID]
			if ok {
				oldRate[*match.Player2Team2ID] += *match.Player2Team2RateAfter - *match.Player2Team2RateBefore
			} else {
				oldRate[*match.Player2Team2ID] = *match.Player2Team2RateAfter - *match.Player2Team2RateBefore
			}
		}
	}

	for playerID, value := range oldRate {
		rateValue, err := s.rdbOperations.GetPlayerRatingByTournamentId(logger, ctx, int64(playerID), tournament.ID)
		if err != nil {
			return err
		}

		updValue := rateValue - int64(value)

		err = s.rwdbOperations.CreateTournamentRating(logger, ctx, int64(playerID), updValue, tournament.ID, &s.config.RWDB)
		if err != nil {
			return err
		}
	}

	err = s.rwdbOperations.DeleteGame(logger, ctx, request.GameID, &s.config.RWDB)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) GetTournamentGameList(ctx context.Context, tournamentId int64) ([]entities.FullTournamentGame, error) {
	logger := s.GetLogger().With().Str("service", "GetTournamentGameList").Logger()

	fullGames := make([]entities.FullTournamentGame, 0)

	games, err := s.rdbOperations.GetTournamentGameList(logger, ctx, &entities.GetTournamentGameList{TournamentId: &tournamentId})
	if err != nil {
		return nil, err
	}

	for _, game := range games {
		fullTeam1, fullTeam2 := entities.FullTournamentTeam{}, entities.FullTournamentTeam{}

		playersTeam1, playersTeam2 := make([]entities.Player, 0), make([]entities.Player, 0)

		// первая команда
		team1, err := s.rdbOperations.GetTeamById(logger, ctx, game.Team1ID, nil)
		if err != nil {
			return nil, err
		}

		for _, pId := range team1.PlayersIds {
			// ее игроки
			player, err := s.rdbOperations.GetPlayerByID(logger, ctx, int(pId), nil)
			if err != nil {
				return nil, err
			}
			playersTeam1 = append(playersTeam1, player)
		}
		fullTeam1.ID = team1.Id
		fullTeam1.Name = team1.Name
		fullTeam1.ShortName = team1.ShortName
		fullTeam1.CityID = pointer.GetValue(team1.CityId)
		fullTeam1.Players = playersTeam1

		// вторая команда
		team2, err := s.rdbOperations.GetTeamById(logger, ctx, game.Team2ID, nil)
		if err != nil {
			return nil, err
		}
		for _, pId := range team2.PlayersIds {
			// ее игроки
			player, err := s.rdbOperations.GetPlayerByID(logger, ctx, int(pId), nil)
			if err != nil {
				return nil, err
			}
			playersTeam2 = append(playersTeam2, player)
		}
		fullTeam2.ID = team2.Id
		fullTeam2.Name = team2.Name
		fullTeam2.ShortName = team2.ShortName
		fullTeam2.CityID = pointer.GetValue(team2.CityId)
		fullTeam2.Players = playersTeam2

		fullGames = append(fullGames, entities.FullTournamentGame{
			Game:  game,
			Team1: &fullTeam1,
			Team2: &fullTeam2,
		})
	}

	return fullGames, nil
}

func (s *Service) GetFutureTournamentGameList(ctx context.Context, req *entities.GetTournamentGameList) ([]entities.TournamentGame, error) {
	logger := s.GetLogger().With().Str("service", "GetFutureTournamentGameList").Logger()

	futureGames := make([]entities.TournamentGame, 0)

	games, err := s.rdbOperations.GetTournamentGameList(logger, ctx, req)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	// "Таймстемп проведения игры > now, либо таймстемп отсутствут - будущая игра"
	for _, game := range games {

		if game.Date == nil || now.Before(*game.Date) {
			futureGames = append(futureGames, game)
		}
	}

	return futureGames, nil
}

func (s *Service) GetPlayedTournamentGameList(ctx context.Context, req *entities.GetTournamentGameList) ([]entities.FullTournamentGame, error) {
	logger := s.GetLogger().With().Str("service", "GetFutureTournamentGameList").Logger()

	playedGames := make([]entities.FullTournamentGame, 0)

	games, err := s.rdbOperations.GetTournamentGameList(logger, ctx, req)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	for _, game := range games {
		if game.Date != nil {

			//  "Таймстемп проведения игры <= now - прошлая"
			if now.After(*game.Date) || now.Equal(*game.Date) {

				//матчи как необязательный, но безвредный элемент, оставляем без обработки ошибки
				matches, _ := s.rdbOperations.FetchMatches(logger, ctx, game.ID)
				playedGames = append(playedGames, entities.FullTournamentGame{
					Game:    game,
					Matches: matches,
				})
			}
		}
	}

	return playedGames, nil
}

/*local methods*/

func (s *Service) validateCaptain(ctx context.Context, logger zerolog.Logger, capId, capTeamId, team1Id, team2Id int64) error {
	captain, err := s.rdbOperations.GetUser(logger, ctx, &capId, nil, &s.config.RDB)
	if err != nil {
		return err
	}

	if captain == nil || captain.Team == nil {
		err = errors.New("капитан не найден или у него нет команды")
		logger.Error().Err(err).Msg("captain is nil or captain.Team is nil")
		return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	if capTeamId != captain.Team.ID {
		err = errors.New("teamId капитана из токена не совпадает с его фактическим teamId")
		logger.Error().Err(err).Msg("token teamId not equal db teamId")
		return error_templates.New(err.Error(), err, codes.Unauthenticated, http.StatusForbidden)
	}

	if captain.Team.ID != team1Id && captain.Team.ID != team2Id {
		err = errors.New("капитан не является членом ни одной из команд")
		logger.Error().Err(err).Msg("captain not in team")
		return error_templates.New(err.Error(), err, codes.Unauthenticated, http.StatusForbidden)
	}

	return nil
}

func (s *Service) validateTournamentMaster(ctx context.Context, logger zerolog.Logger, tournament *entities.Tournament, masterId int64) error {
	master, err := s.rdbOperations.GetTournamentMasterByUserId(logger, ctx, masterId, &s.config.RDB)
	if err != nil {
		return err
	}

	if tournament.CityID != master.City.ID {
		err = errors.New("город турнира и город мастера по турнирам не совпадают")
		logger.Error().Err(err).Msg("master city not equal tournament city")
		return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	return nil
}

/*local functions*/

// содержится ли leagueID в массиве.
func contains(arr []entities.LeagueShort, leagueID int64) bool {
	for _, v := range arr {
		if v.ID == leagueID {
			return true
		}
	}
	return false
}

// содержится ли leagueID в обоих массивах.
func bothContain(leagueID int64, leaguesTeam1, leaguesTeam2 []entities.LeagueShort) bool {
	return contains(leaguesTeam1, leagueID) && contains(leaguesTeam2, leagueID)
}

// checkTournamentStage ищет есть ли искомый этап в этапах турнира и не является ли он завершенным
func checkTournamentStage(stages []entities.TournamentStage, targetStageId int64) error {
	isContains := false

	for _, stage := range stages {
		if stage.IsFinished == true {
			return errors.New("этап завершен")
		}
		if stage.ID == targetStageId {
			isContains = true
		}
	}

	if isContains == false {
		return errors.New("этап не найден")
	}

	return nil
}

// buildNewRequest собирает новый запрос и старые данные в одну структуру с приоритетом на новый запрос
func buildNewRequest(request *entities.UpdateFutureTournamentGameRequest, game *entities.TournamentGame) *entities.UpdateFutureTournamentGameRequest {
	nr := &entities.UpdateFutureTournamentGameRequest{}

	nr.GameID = game.ID

	if request.PlaceID != nil {
		nr.PlaceID = request.PlaceID
	} else {
		nr.PlaceID = game.PlaceID
	}

	if request.Date != nil {
		nr.Date = request.Date
	} else {
		nr.Date = game.Date
	}

	if request.Team1ID != nil {
		nr.Team1ID = request.Team1ID
	} else {
		nr.Team1ID = &game.Team1ID
	}

	if request.Team2ID != nil {
		nr.Team2ID = request.Team2ID
	} else {
		nr.Team2ID = &game.Team2ID
	}

	nr.Executor = request.Executor

	return nr
}

// checkUpdateRegularTournamentGame проверяет условия при которых обновления для игры турнира Regular невозможны
func checkUpdateRegularTournamentGame(logger zerolog.Logger, team1ID, team2ID int64, game *entities.TournamentGame) error {
	if (team1ID != game.Team1ID && team1ID != game.Team2ID) || (team2ID != game.Team1ID && team2ID != game.Team2ID) {
		err := errors.New("в игре турнира типа Regular нельзя менять состав команд")
		logger.Error().Err(err).Msg("change game teams pair in update Regular")
		return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	return nil
}

// checkDeleteRegularTournamentGame проверяет условия при которых удаление игры турнира Regular.BestOfOne невозможны
func checkDeleteRegularTournamentGame(logger zerolog.Logger, game *entities.TournamentGame) error {
	if game.IsTiebreak == false {
		err := errors.New("у регулярного турнира можно удалить только tiebreak игру")
		logger.Error().Err(err).Msg("not tiebreak in create Regular")
		return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	return nil
}

// checkCreateRegularTournamentGame проверяет условия при которых создание игры турнира Regular невозможны
func checkCreateRegularTournamentGame(logger zerolog.Logger, isTiebreak bool) error {
	// для этого типа турнира можно создать только Tiebreak, потому что при создании турнира вся турнирная сетка создается сразу
	// TODO: уточнить корректность условия для создания игры регулярного турнира
	if isTiebreak != true {
		err := errors.New("для регулярного турнира c одной игрой можно создать дополнительно только tiebreak игру")
		logger.Error().Err(err).Msg("not tiebreak in Regular")
		return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	return nil
}

func (s *Service) getPlayerRating(ctx context.Context, logger zerolog.Logger, playerID int, leagueID int64, rates map[int]int) (int, error) {
	if value, ok := rates[playerID]; ok {
		return value, nil
	}
	resp, err := s.rdbOperations.GetRatingByPlayerIDAndByLeagueID(logger, ctx, int64(playerID), leagueID)
	if err != nil {
		if outputError, ok := (err).(*error_templates.OutputError); ok {
			code, _ := outputError.GetHTTP()
			if code == http.StatusNotFound {
				return 1000, nil
			}
			return 0, err
		}
		return 0, err
	}
	return int(resp), nil
}
