package game

import (
	"context"
	"errors"
	"fmt"
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
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/helpers"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/helpers/pointer"
)

func (s *Service) Create(ctx context.Context, request entities.CreateGameRequest) (entities.CreateGameResponse, error) {
	logger := s.logger.With().Str("service", "game.Create").Logger()

	league, err := s.rdbOperations.GetLeagueById(logger, ctx, int64(request.LeagueID), nil)
	if err != nil {
		return entities.CreateGameResponse{}, err
	}

	team1resp, err := s.rdbOperations.GetTeam(logger, ctx, int64(request.Team1ID))
	if err != nil {
		return entities.CreateGameResponse{}, err
	}
	team2resp, err := s.rdbOperations.GetTeam(logger, ctx, int64(request.Team2ID))
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
		master, err := s.rdbOperations.GetTournamentMasterByUserId(logger, ctx, request.Creator.ID)
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

	tx, err := s.rwdbOperations.BeginTx(ctx, logger)
	if err != nil {
		return entities.CreateGameResponse{}, err
	}

	gameId, err := s.rwdbOperations.CreateGame(logger, ctx, request, tx)
	if err != nil {
		tx.Rollback(ctx)
		return entities.CreateGameResponse{}, err
	}

	matchIds := make([]int64, 0, len(matches))

	for _, match := range request.Matches {
		id, err := s.rwdbOperations.CreateGameMatch(logger, ctx, match, gameId, tx)
		if err != nil {
			tx.Rollback(ctx)
			return entities.CreateGameResponse{}, err
		}
		matchIds = append(matchIds, id)
	}

	for playerID, value := range rates {
		if playerID == 0 {
			continue
		}

		err = s.rwdbOperations.CreateRatingUpdateOnConflict(logger, ctx, entities.Rating{
			PlayerID: int64(playerID),
			LeagueID: leagueID,
			Value:    int64(value),
		}, tx)
		if err != nil {
			tx.Rollback(ctx)
			return entities.CreateGameResponse{}, err
		}
	}

	if err = tx.Commit(ctx); err != nil {
		tx.Rollback(ctx)
		return entities.CreateGameResponse{}, err
	}

	return entities.CreateGameResponse{
		GameID:   gameId,
		MatchIDs: matchIds,
	}, nil
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
	team1resp, err := s.rdbOperations.GetTeam(logger, ctx, int64(game.Team1ID))
	if err != nil {
		return err
	}
	team2resp, err := s.rdbOperations.GetTeam(logger, ctx, int64(game.Team2ID))
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
		master, err := s.rdbOperations.GetTournamentMasterByUserId(logger, ctx, req.Executor.ID)
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

	tx, err := s.rwdbOperations.BeginTx(ctx, logger)
	if err != nil {
		return err
	}

	if recalc {
		// Decreasing rating for each player of game matches
		plGmRtInc := make(map[int]int, 0) // playerGameRatingIncrease
		_matches, err := s.rdbOperations.GetMatchListByGameID(ctx, logger, int64(game.ID))
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

			err = s.rwdbOperations.CreateRatingUpdateOnConflict(logger, ctx, *rate, tx)
			if err != nil {
				tx.Rollback(ctx)
				return err
			}
		}
	}

	err = s.rwdbOperations.DeleteGame(ctx, logger, int64(req.ID), tx)
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
	logger := s.logger.With().Str("service", "game.Update").Logger()

	league, err := s.rdbOperations.GetLeagueById(logger, ctx, int64(request.LeagueID), nil)
	if err != nil {
		return err
	}

	if league.ID != int64(request.LeagueID) {
		err = errors.New("лига игры и лига в запросе не совпадают")
		logger.Error().Msg("LeagueID mismatch")
		return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	team1resp, err := s.rdbOperations.GetTeam(logger, ctx, int64(request.Team1ID))
	if err != nil {
		return err
	}
	team2resp, err := s.rdbOperations.GetTeam(logger, ctx, int64(request.Team2ID))
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

	for _, m := range request.Matches {
		if m.Team1ID != request.Team1ID || m.Team2ID != request.Team2ID {
			logger.Error().Err(err).Msg("match team not equal request team")
			return error_templates.New(pkgerr.ErrDifferentTeams, errors.New(pkgerr.ErrDifferentTeams), codes.InvalidArgument, http.StatusBadRequest)
		}
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
		master, err := s.rdbOperations.GetTournamentMasterByUserId(logger, ctx, request.Executor.ID)
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
	_matches, err := s.rdbOperations.GetMatchListByGameID(ctx, logger, int64(request.ID))
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

	// список id-шников матчей НЕ подлежащих удалению
	var matchIds = make([]int, 0)
	for _, match := range request.Matches {
		if match.ID != nil {
			matchIds = append(matchIds, *match.ID)
		}
	}

	tx, err := s.rwdbOperations.BeginTx(ctx, logger)
	if err != nil {
		return err
	}

	// удаление всех матчей игры кроме входящих
	err = s.rwdbOperations.DeleteOldGameMatches(ctx, logger, int64(request.ID), matchIds, tx)
	if err != nil {
		tx.Rollback(ctx)
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
			err = s.rwdbOperations.CreateNewMatch(ctx, logger, int64(request.ID), &match, tx)
			if err != nil {
				tx.Rollback(ctx)
				return err
			}
		} else {
			err = s.rwdbOperations.UpdateOldMatch(ctx, logger, int64(request.ID), &match, tx)
			if err != nil {
				tx.Rollback(ctx)
				return err
			}
		}
	}

	// обновление рейтингов
	for _, rating := range newRates {
		err = s.rwdbOperations.CreateRatingUpdateOnConflict(logger, ctx, rating, tx)
		if err != nil {
			tx.Rollback(ctx)
			return err
		}
	}

	// обновление игры
	err = s.rwdbOperations.UpdateGame(logger, ctx, request, tx)
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
		master, err := s.rdbOperations.GetTournamentMasterByUserId(logger, ctx, request.Executor.ID)
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
		master, err := s.rdbOperations.GetTournamentMasterByUserId(logger, ctx, request.Creator.ID)
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
		master, err := s.rdbOperations.GetTournamentMasterByUserId(logger, ctx, req.Executor.ID)
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

	tournament, err := s.rdbOperations.GetTournamentById(ctx, logger, request.TournamentID)
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
	_, err = checkTournamentStage(tournament.Stages, request.StageID)
	if err != nil {
		logger.Error().Err(err).Msg("failed checkTournamentStage")
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

	if tournament.TypeID == constant.RegularTournamentTypeID ||
		tournament.TypeID == constant.RegularOneVsOneTournamentTypeID ||
		tournament.TypeID == constant.PlayoffTournamentTypeID {
		if err = checkTiebreakGame(logger, request.IsTiebreak); err != nil {
			return 0, err
		}
	} else if tournament.TypeID == constant.RegularPlayoffTournamentTypeID {
		if err = checkTiebreakGame(logger, request.IsTiebreak); err != nil {
			return 0, err
		}
		if err = s.checkTeamsByStageRegularPlayoff(ctx, logger, request.Team1ID, request.Team2ID, request.StageID); err != nil {
			return 0, err
		}
	} else {
		// todo другие типы турниров
		return 0, errors.New("not implemented")
	}

	tx, err := s.rwdbOperations.BeginTx(ctx, logger)
	if err != nil {
		return 0, err
	}

	id, err := s.rwdbOperations.CreateFutureTournamentGame(ctx, logger, request, tx)
	if err != nil {
		tx.Rollback(ctx)
		return 0, err
	}

	if err = tx.Commit(ctx); err != nil {
		tx.Rollback(ctx)
		return 0, err
	}

	return id, nil
}

func (s *Service) UpdateFutureTournamentGame(ctx context.Context, request *entities.UpdateFutureTournamentGameRequest) error {
	logger := s.GetLogger().With().Str("service", "UpdateFutureTournamentGame").Logger()

	game, err := s.rdbOperations.GetTournamentGame(ctx, logger, request.GameID, &s.config.RDB)
	if err != nil {
		return err
	}

	stage, err := s.rdbOperations.GetTournamentStage(ctx, logger, game.StageID)
	if err != nil {
		return err
	}

	tournament, err := s.rdbOperations.GetTournamentById(ctx, logger, stage.TournamentID)
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

	if tournament.TypeID == constant.RegularTournamentTypeID ||
		tournament.TypeID == constant.RegularOneVsOneTournamentTypeID {
		err = checkTournamentGamePairs(logger, *updReq.Team1ID, *updReq.Team2ID, &game)
		if err != nil {
			return err
		}
	} else if tournament.TypeID == constant.RegularPlayoffTournamentTypeID {
		if err = checkTournamentGamePairs(logger, *updReq.Team1ID, *updReq.Team2ID, &game); err != nil {
			return err
		}
		if err = s.checkTeamsByStageRegularPlayoff(ctx, logger, *updReq.Team1ID, *updReq.Team2ID, stage.ID); err != nil {
			return err
		}
	} else {
		// todo другие типы турниров
		return errors.New("not implemented")
	}

	if err = s.checkGameForMatches(ctx, logger, game.ID); err != nil {
		return err
	}

	tx, err := s.rwdbOperations.BeginTx(ctx, logger)
	if err != nil {
		return err
	}

	err = s.rwdbOperations.UpdateFutureTournamentGame(logger, ctx, updReq, tx)
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

func (s *Service) DeleteFutureTournamentGame(ctx context.Context, request *entities.DeleteFutureTournamentGameRequest) error {
	logger := s.GetLogger().With().Str("service", "DeleteFutureTournamentGame").Logger()

	game, err := s.rdbOperations.GetTournamentGame(ctx, logger, request.GameID, &s.config.RDB)
	if err != nil {
		return err
	}

	stage, err := s.rdbOperations.GetTournamentStage(ctx, logger, game.StageID)
	if err != nil {
		return err
	}

	tournament, err := s.rdbOperations.GetTournamentById(ctx, logger, stage.TournamentID)
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

	if err = s.checkGameForMatches(ctx, logger, game.ID); err != nil {
		return err
	}

	if tournament.TypeID == constant.RegularTournamentTypeID ||
		tournament.TypeID == constant.RegularOneVsOneTournamentTypeID ||
		tournament.TypeID == constant.RegularPlayoffTournamentTypeID {
		err = checkDeleteTournamentGame(logger, game)
		if err != nil {
			return err
		}
	} else {
		// todo другие типы турниров
		return errors.New("not implemented")
	}

	tx, err := s.rwdbOperations.BeginTx(ctx, logger)
	if err != nil {
		return err
	}

	err = s.rwdbOperations.DeleteGame(ctx, logger, game.ID, tx)
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

func (s *Service) CreatePlayedTournamentGame(ctx context.Context, request *entities.CreatePlayedTournamentGameRequest) (entities.CreatePlayedTournamentGameResponse, error) {
	logger := s.logger.With().Str("service", "game.CreatePlayedTournamentGame").Logger()

	tournament, err := s.rdbOperations.GetTournamentById(ctx, logger, request.TournamentID)
	if err != nil {
		return entities.CreatePlayedTournamentGameResponse{}, err
	}

	request.CityID = tournament.CityID

	if err = s.checkRole(ctx, logger, tournament, request.Creator, request.Team1ID, request.Team2ID); err != nil {
		return entities.CreatePlayedTournamentGameResponse{}, err
	}

	// проверяем соответствует ли столоместо городу турнира
	if err = s.checkBarPlaceByTournamentCity(ctx, logger, request.PlaceID, tournament.CityID); err != nil {
		return entities.CreatePlayedTournamentGameResponse{}, err
	}

	// проверяем относится ли входящий id этапа к этапам турнира и не является ли завершенным
	currentStage, err := checkTournamentStage(tournament.Stages, request.StageID)
	if err != nil {
		logger.Error().Err(err).Msg("failed checkTournamentStage")
		return entities.CreatePlayedTournamentGameResponse{}, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	// парсим порядковый номер этапа и общее кол-во этапов из поля number
	currentStageNumber, stageQty, err := helpers.ParseTournamentStageNumber(currentStage.Number, constant.SeparatorStageNumber)
	if err != nil {
		logger.Error().Err(err).Msg("failed helpers.ParseTournamentStageNumber")
		return entities.CreatePlayedTournamentGameResponse{}, error_templates.New(err.Error(), err, codes.Internal, http.StatusInternalServerError)
	}

	if err = checkTeams(tournament, request.Team1ID, request.Team2ID); err != nil {
		logger.Error().Err(err).Msg(err.Error())
		return entities.CreatePlayedTournamentGameResponse{}, err
	}

	// проверяем что id технически проигравшей команды принадлежит одной из игравших
	if request.TechLooseTeamID != nil && (*request.TechLooseTeamID != request.Team1ID && *request.TechLooseTeamID != request.Team2ID) {
		err = errors.New("нельзя присудить техническое поражение команде не участвующей в игре")
		logger.Error().Err(err).Msg("techLooseTeamID not equal team1Id/team2Id")
		return entities.CreatePlayedTournamentGameResponse{}, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	for _, m := range request.Matches {
		if request.Team1ID != int64(m.Team1ID) || request.Team2ID != int64(m.Team2ID) {
			err = errors.New("команда матча не участвует в игре")
			logger.Error().Err(err).Msg("match team not equal game team")
			return entities.CreatePlayedTournamentGameResponse{}, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}

		if m.Date.Year() != request.Date.Year() || m.Date.Month() != request.Date.Month() || m.Date.Day() != request.Date.Day() {
			err = errors.New("даты игры и матча не совпадают")
			logger.Error().Err(err).Msg("dates not equal")
			return entities.CreatePlayedTournamentGameResponse{}, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}
	}

	// пока вся сетка имеющихся типов турниров формируется автоматически,
	// дополнительно можно создать только tiebreak игру
	if request.IsTiebreak != true {
		err = errors.New("можно создать дополнительно только tiebreak игру")
		logger.Error().Err(err).Msg("tiebreak only for create played game")
		return entities.CreatePlayedTournamentGameResponse{}, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	if tournament.TypeID == constant.RegularOneVsOneTournamentTypeID {
		for _, m := range request.Matches {
			if m.Player2Team1Id != nil || m.Player2Team2Id != nil {
				err = errors.New("в турнире типа Regular.OneVsOne у команды не может быть второго игрока")
				logger.Error().Err(err).Msg("two players in OneVsOne")
				return entities.CreatePlayedTournamentGameResponse{}, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
			}
		}
	} else if tournament.TypeID == constant.RegularPlayoffTournamentTypeID {
		if err = s.checkTeamsByStageRegularPlayoff(ctx, logger, request.Team1ID, request.Team2ID, currentStage.ID); err != nil {
			return entities.CreatePlayedTournamentGameResponse{}, err
		}
	} else if tournament.TypeID == constant.RegularTournamentTypeID {
		// @ToDo: проверить, нужно ли доделать функционал
	} else if tournament.TypeID == constant.PlayoffTournamentTypeID {
		// @ToDo: проверить, нужно ли доделать функционал
	} else {
		return entities.CreatePlayedTournamentGameResponse{}, errors.New("not implemented")
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
				return entities.CreatePlayedTournamentGameResponse{}, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
			}

			// проверка соответствия команд
			if request.Team1ID != int64(match.Team1ID) || request.Team2ID != int64(match.Team2ID) {
				err = errors.New("команды матчей и игр должны совпадать")
				logger.Error().Err(err).Msg("game teams not equal match teams")
				return entities.CreatePlayedTournamentGameResponse{}, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
			}

			if match.ScoreTeam1 == 0 && match.ScoreTeam2 == 0 {
				continue
			}

			player1team1ID := match.Player1Team1Id
			player1team2ID := match.Player1Team2Id

			// Получаем рейтинги для всех игроков
			player1team1rate, err := s.getPlayerRating(ctx, logger, player1team1ID, tournament.ID, rates)
			if err != nil {
				return entities.CreatePlayedTournamentGameResponse{}, err
			}

			player1team2rate, err := s.getPlayerRating(ctx, logger, player1team2ID, tournament.ID, rates)
			if err != nil {
				return entities.CreatePlayedTournamentGameResponse{}, err
			}

			var player2team1rate, player2team2rate int
			var player2team1ID, player2team2ID int

			// Получаем рейтинг для player2team1 (если существует и > 0)
			if match.Player2Team1Id != nil {
				player2team1ID = *match.Player2Team1Id
				if player2team1ID > 0 {
					player2team1rate, err = s.getPlayerRating(ctx, logger, player2team1ID, tournament.ID, rates)
					if err != nil {
						return entities.CreatePlayedTournamentGameResponse{}, err
					}
				}
			}

			// Получаем рейтинг для player2team2 (если существует и > 0)
			if match.Player2Team2Id != nil {
				player2team2ID = *match.Player2Team2Id
				if player2team2ID > 0 {
					player2team2rate, err = s.getPlayerRating(ctx, logger, player2team2ID, tournament.ID, rates)
					if err != nil {
						return entities.CreatePlayedTournamentGameResponse{}, err
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
				return entities.CreatePlayedTournamentGameResponse{}, err
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

	tx, err := s.rwdbOperations.BeginTx(ctx, logger)
	if err != nil {
		return entities.CreatePlayedTournamentGameResponse{}, err
	}

	id, err := s.rwdbOperations.CreatePlayedTournamentGame(ctx, logger, request, tx)
	if err != nil {
		tx.Rollback(ctx)
		return entities.CreatePlayedTournamentGameResponse{}, err
	}

	for _, m := range request.Matches {
		_, err = s.rwdbOperations.CreateGameMatch(logger, ctx, m, id, tx)
		if err != nil {
			tx.Rollback(ctx)
			return entities.CreatePlayedTournamentGameResponse{}, err
		}
	}

	for playerID, value := range rates {
		err = s.rwdbOperations.CreateTournamentRating(ctx, logger, int64(playerID), int64(value), request.TournamentID, tx)
		if err != nil {
			tx.Rollback(ctx)
			return entities.CreatePlayedTournamentGameResponse{}, err
		}
	}

	if err = tx.Commit(ctx); err != nil {
		tx.Rollback(ctx)
		return entities.CreatePlayedTournamentGameResponse{}, err
	}

	var stateStageMsg string

	if tournament.TypeID == constant.RegularTournamentTypeID || tournament.TypeID == constant.RegularOneVsOneTournamentTypeID {
		stateStageMsg = "не завершён: этапы турниров Regular и 1vs1 завершаются вручную"
	} else if tournament.TypeID == constant.PlayoffTournamentTypeID {
		stateStageMsg = s.checkPlayoffStageForFinish(ctx, logger, currentStage, tournament, currentStageNumber, stageQty)
	} else if tournament.TypeID == constant.RegularPlayoffTournamentTypeID {
		stateStageMsg = s.checkRegularPlayoffStageForFinish(ctx, logger, currentStage, tournament, currentStageNumber, stageQty)
	} else {
		return entities.CreatePlayedTournamentGameResponse{}, errors.New("not implemented")
	}

	return entities.CreatePlayedTournamentGameResponse{
		GameId:     id,
		StageId:    currentStage.ID,
		StageState: stateStageMsg,
	}, nil
}

func (s *Service) UpdatePlayedTournamentGame(ctx context.Context, request *entities.UpdatePlayedTournamentGameRequest) (entities.UpdatePlayedTournamentGameResponse, error) {
	logger := s.GetLogger().With().Str("service", "UpdatePlayedTournamentGame").Logger()

	game, err := s.rdbOperations.GetTournamentGame(ctx, logger, request.GameID, &s.config.RDB)
	if err != nil {
		return entities.UpdatePlayedTournamentGameResponse{}, err
	}

	currentStage, err := s.rdbOperations.GetTournamentStage(ctx, logger, game.StageID)
	if err != nil {
		return entities.UpdatePlayedTournamentGameResponse{}, err
	}
	if currentStage.IsFinished == true {
		err = errors.New("нельзя редактировать игры завершенного этапа")
		return entities.UpdatePlayedTournamentGameResponse{}, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}
	request.StageID = currentStage.ID

	// парсим порядковый номер этапа и общее кол-во этапов из поля number
	currentStageNumber, stageQty, err := helpers.ParseTournamentStageNumber(currentStage.Number, constant.SeparatorStageNumber)
	if err != nil {
		logger.Error().Err(err).Msg("failed helpers.ParseTournamentStageNumber")
		return entities.UpdatePlayedTournamentGameResponse{}, error_templates.New(err.Error(), err, codes.Internal, http.StatusInternalServerError)
	}

	tournament, err := s.rdbOperations.GetTournamentById(ctx, logger, currentStage.TournamentID)
	if err != nil {
		return entities.UpdatePlayedTournamentGameResponse{}, err
	}

	if err = s.checkRole(ctx, logger, tournament, request.Executor, request.Team1ID, request.Team2ID); err != nil {
		return entities.UpdatePlayedTournamentGameResponse{}, err
	}

	// проверяем соответствует ли столоместо городу турнира
	if err = s.checkBarPlaceByTournamentCity(ctx, logger, request.PlaceID, tournament.CityID); err != nil {
		return entities.UpdatePlayedTournamentGameResponse{}, err
	}

	if err = checkTeams(tournament, request.Team1ID, request.Team2ID); err != nil {
		logger.Error().Err(err).Msg(err.Error())
		return entities.UpdatePlayedTournamentGameResponse{}, err
	}

	// проверяем что не пришли посторонние команды
	err = checkTournamentGamePairs(logger, request.Team1ID, request.Team2ID, &game)
	if err != nil {
		return entities.UpdatePlayedTournamentGameResponse{}, err
	}

	// проверки специфичные для PlayOff турниров
	if tournament.TypeID == constant.PlayoffTournamentTypeID {
		_, err = checkTournamentStage(tournament.Stages, request.StageID)
		if err != nil {
			logger.Error().Err(err).Msg("failed to checkTournamentStage")
			return entities.UpdatePlayedTournamentGameResponse{}, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}
	} else if tournament.TypeID == constant.RegularPlayoffTournamentTypeID {
		if _, err = checkTournamentStage(tournament.Stages, request.StageID); err != nil {
			logger.Error().Err(err).Msg("failed to checkTournamentStage")
			return entities.UpdatePlayedTournamentGameResponse{}, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}
		if err = s.checkTeamsByStageRegularPlayoff(ctx, logger, request.Team1ID, request.Team2ID, currentStage.ID); err != nil {
			logger.Error().Err(err).Msg("failed to checkTeamsByStageRegularPlayoff")
			return entities.UpdatePlayedTournamentGameResponse{}, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}
	} else if tournament.TypeID == constant.RegularTournamentTypeID {
		// @ToDo: проверить, нужно ли доделать функционал
	} else if tournament.TypeID == constant.RegularOneVsOneTournamentTypeID {
		// @ToDo: проверить, нужно ли доделать функционал
	} else {
		return entities.UpdatePlayedTournamentGameResponse{}, errors.New("not implemented")
	}

	newRates := make(map[int]int)
	oldRates := make(map[int]int)

	// прошедшие матчи
	matchesV2, err := s.rdbOperations.GetMatchListByGameID(ctx, logger, request.GameID)
	if err != nil {
		return entities.UpdatePlayedTournamentGameResponse{}, err
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
		rateValue, err := s.rdbOperations.GetPlayerRatingByTournamentId(ctx, logger, int64(playerID), tournament.ID)
		if err != nil {
			return entities.UpdatePlayedTournamentGameResponse{}, err
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
			return entities.UpdatePlayedTournamentGameResponse{}, err
		}

		player1team2rate, err := s.getPlayerRating(ctx, logger, player1team2ID, tournament.ID, newRates)
		if err != nil {
			return entities.UpdatePlayedTournamentGameResponse{}, err
		}

		var player2team1rate, player2team2rate int
		var player2team1ID, player2team2ID int

		// Получаем рейтинг для player2team1 (если существует и > 0)
		if match.Player2Team1Id != nil {
			player2team1ID = *match.Player2Team1Id
			if player2team1ID > 0 {
				player2team1rate, err = s.getPlayerRating(ctx, logger, player2team1ID, tournament.ID, newRates)
				if err != nil {
					return entities.UpdatePlayedTournamentGameResponse{}, err
				}
			}
		}

		// Получаем рейтинг для player2team2 (если существует и > 0)
		if match.Player2Team2Id != nil {
			player2team2ID = *match.Player2Team2Id
			if player2team2ID > 0 {
				player2team2rate, err = s.getPlayerRating(ctx, logger, player2team2ID, tournament.ID, newRates)
				if err != nil {
					return entities.UpdatePlayedTournamentGameResponse{}, err
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
			return entities.UpdatePlayedTournamentGameResponse{}, err
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

	tx, err := s.rwdbOperations.BeginTx(ctx, logger)
	if err != nil {
		return entities.UpdatePlayedTournamentGameResponse{}, err
	}

	// удаление всех матчей игры кроме входящих
	err = s.rwdbOperations.DeleteOldGameMatches(ctx, logger, request.GameID, matchIds, tx)
	if err != nil {
		tx.Rollback(ctx)
		return entities.UpdatePlayedTournamentGameResponse{}, err
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
			err = s.rwdbOperations.CreateNewMatch(ctx, logger, request.GameID, &match, tx)
			if err != nil {
				tx.Rollback(ctx)
				return entities.UpdatePlayedTournamentGameResponse{}, err
			}
		} else {
			err = s.rwdbOperations.UpdateOldMatch(ctx, logger, request.GameID, &match, tx)
			if err != nil {
				tx.Rollback(ctx)
				return entities.UpdatePlayedTournamentGameResponse{}, err
			}
		}
	}

	// обновление рейтингов
	for playerID, value := range newRates {
		err = s.rwdbOperations.CreateTournamentRating(ctx, logger, int64(playerID), int64(value), tournament.ID, tx)
		if err != nil {
			tx.Rollback(ctx)
			return entities.UpdatePlayedTournamentGameResponse{}, err
		}
	}

	// обновление игры
	err = s.rwdbOperations.UpdatePlayedTournamentGame(ctx, logger, request, tx)
	if err != nil {
		tx.Rollback(ctx)
		return entities.UpdatePlayedTournamentGameResponse{}, err
	}

	if err = tx.Commit(ctx); err != nil {
		tx.Rollback(ctx)
		return entities.UpdatePlayedTournamentGameResponse{}, err
	}

	var stateStageMsg string
	if tournament.TypeID == constant.RegularTournamentTypeID || tournament.TypeID == constant.RegularOneVsOneTournamentTypeID {
		stateStageMsg = "не завершён: этапы турниров Regular и 1vs1 завершаются вручную"
	} else if tournament.TypeID == constant.PlayoffTournamentTypeID {
		stateStageMsg = s.checkPlayoffStageForFinish(ctx, logger, currentStage, tournament, currentStageNumber, stageQty)
	} else if tournament.TypeID == constant.RegularPlayoffTournamentTypeID {
		stateStageMsg = s.checkRegularPlayoffStageForFinish(ctx, logger, currentStage, tournament, currentStageNumber, stageQty)
	} else {
		return entities.UpdatePlayedTournamentGameResponse{}, errors.New("not implemented")
	}

	return entities.UpdatePlayedTournamentGameResponse{
		GameID:     request.GameID,
		StageID:    request.StageID,
		StageState: stateStageMsg,
	}, nil
}

func (s *Service) DeletePlayedTournamentGame(ctx context.Context, request *entities.DeletePlayedTournamentGameRequest) error {
	logger := s.GetLogger().With().Str("service", "DeletePlayedTournamentGame").Logger()

	game, err := s.rdbOperations.GetTournamentGame(ctx, logger, request.GameID, &s.config.RDB)
	if err != nil {
		return err
	}

	stage, err := s.rdbOperations.GetTournamentStage(ctx, logger, game.StageID)
	if err != nil {
		return err
	}

	tournament, err := s.rdbOperations.GetTournamentById(ctx, logger, stage.TournamentID)
	if err != nil {
		return err
	}

	if err = s.checkRole(ctx, logger, tournament, request.Executor, game.Team1ID, game.Team2ID); err != nil {
		return err
	}

	// не позволяем удалять игры у завршенных этапов
	if stage.IsFinished == true {
		err = errors.New("у завершенного этапа нельзя удалять игры")
		logger.Error().Err(err).Msg("stage is finished")
		return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	if tournament.TypeID == constant.RegularTournamentTypeID ||
		tournament.TypeID == constant.RegularOneVsOneTournamentTypeID ||
		tournament.TypeID == constant.PlayoffTournamentTypeID ||
		tournament.TypeID == constant.RegularPlayoffTournamentTypeID {
		if err = checkDeleteTournamentGame(logger, game); err != nil {
			return err
		}
	} else {
		return errors.New("not implemented")
	}

	matches, err := s.rdbOperations.GetMatchListByGameID(ctx, logger, request.GameID)
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

	tx, err := s.rwdbOperations.BeginTx(ctx, logger)
	if err != nil {
		return err
	}

	for playerID, value := range oldRate {
		rateValue, err := s.rdbOperations.GetPlayerRatingByTournamentId(ctx, logger, int64(playerID), tournament.ID)
		if err != nil {
			return err
		}

		updValue := rateValue - int64(value)

		err = s.rwdbOperations.CreateTournamentRating(ctx, logger, int64(playerID), updValue, tournament.ID, tx)
		if err != nil {
			tx.Rollback(ctx)
			return err
		}
	}

	err = s.rwdbOperations.DeleteGame(ctx, logger, request.GameID, tx)
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
	master, err := s.rdbOperations.GetTournamentMasterByUserId(logger, ctx, masterId)
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

func (s *Service) checkPlayoffStageForFinish(
	ctx context.Context,
	logger zerolog.Logger,
	stage entities.TournamentStage,
	tournament entities.Tournament,
	currenStageNumber int64,
	stageQty int64,
) string {
	games, err := s.rdbOperations.GetTournamentStageGames(ctx, logger, stage.ID)
	if err != nil {
		logger.Error().Err(err).Msg("failed GetTournamentStageGames")
		return "не может быть завершён: " + err.Error()
	}

	teams, err := s.rdbOperations.GetTournamentTeamList(ctx, logger, tournament.ID)
	if err != nil {
		return "не может быть завершён: " + err.Error()
	}

	teamsExtraPoints := make(map[int64]int64, len(teams))

	for _, team := range teams {
		extraPoints, err := s.rdbOperations.GetTournamentTeamExtraPointsCountV2(ctx, logger, team.ID, tournament.ID)
		if err != nil {
			return "не может быть завершён: " + err.Error()
		}

		teamsExtraPoints[team.ID] = extraPoints
	}

	if len(games) == 0 {
		logger.Error().Msg("games not found")
		return "не может быть завершён: нет игр"
	}

	gameStats := make([]entities.GameStat, 0, len(games))

	for _, game := range games {
		var gs entities.GameStat
		gs.Game = game

		if game.Date == nil || (*game.Date).After(time.Now()) {
			logger.Error().Msg("date in future or nil")
			return "не может быть завершён: есть несыгранная игра"
		}

		matches, err := s.rdbOperations.FetchMatches(logger, ctx, game.ID)
		if err != nil {
			logger.Error().Err(err).Msg("failed FetchMatches")
			return "не может быть завершён: " + err.Error()
		}

		gs.Matches = matches

		if len(matches) == 0 && game.TechLooseTeamID == nil {
			logger.Error().Msg("matches not found, tech loose is nil")
			return "не может быть завершён: игра не содержит матчи и не указано техническое поражение"
		}

		// суммируем все очки матчей для каждой команды
		var totalScoreTeam1 int64
		var totalScoreTeam2 int64
		for _, m := range matches {
			totalScoreTeam1 += m.ScoreTeam1
			totalScoreTeam2 += m.ScoreTeam2

			if extraPointsTeam1, ok := teamsExtraPoints[m.Team1ID]; ok && extraPointsTeam1 > 0 {
				totalScoreTeam1 += extraPointsTeam1
			} else if extraPointsTeam2, ok2 := teamsExtraPoints[m.Team2ID]; ok2 && extraPointsTeam2 > 0 {
				totalScoreTeam2 += extraPointsTeam2
			}
		}

		// если у кого-то есть преимущество по очкам - то он победитель, иначе - ничего не делаем
		if totalScoreTeam1 > totalScoreTeam2 {
			gs.WinnerId = game.Team1ID
		} else if totalScoreTeam2 > totalScoreTeam1 {
			gs.WinnerId = game.Team2ID
		}

		gameStats = append(gameStats, gs)
	}

	winners := helpers.ExtractWinners(gameStats)

	err = helpers.CheckQtyWinnersInCurrentPlayoffStage(len(tournament.TeamIDs), len(winners), int(currenStageNumber), int(stageQty))
	if err != nil {
		return "не может быть завершён: " + err.Error()
	}

	return "этап может быть завершён"
}

func (s *Service) checkRegularPlayoffStageForFinish(
	ctx context.Context,
	logger zerolog.Logger,
	stage entities.TournamentStage,
	tournament entities.Tournament,
	currenStageNumber int64,
	stageQty int64,
) string {
	games, err := s.rdbOperations.GetTournamentStageGames(ctx, logger, stage.ID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to GetTournamentStageGames")
		return "не может быть завершён: " + err.Error()
	}

	if len(games) == 0 {
		logger.Error().Msg("games not found")
		return "не может быть завершён: нет игр"
	}

	teams, err := s.rdbOperations.GetTournamentTeamList(ctx, logger, tournament.ID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to GetTournamentTeamList")
		return "не может быть завершён: " + err.Error()
	}

	teamsExtraPoints := make(map[int64]int64, len(teams))

	for _, team := range teams {
		extraPoints, err := s.rdbOperations.GetTournamentTeamExtraPointsCountV2(ctx, logger, team.ID, tournament.ID)
		if err != nil {
			logger.Error().Err(err).Msg("failed to GetTournamentTeamExtraPointsCountV2")
			return "не может быть завершён: " + err.Error()
		}

		teamsExtraPoints[team.ID] = extraPoints
	}

	gameStats := make([]entities.GameStat, 0, len(games))

	for _, game := range games {
		var gs entities.GameStat
		gs.Game = game

		if game.Date == nil || (*game.Date).After(time.Now()) {
			logger.Error().Msg("date in future or nil")
			return "не может быть завершён: есть несыгранная игра"
		}

		matches, err := s.rdbOperations.FetchMatches(logger, ctx, game.ID)
		if err != nil {
			logger.Error().Err(err).Msg("failed FetchMatches")
			return "не может быть завершён: " + err.Error()
		}

		gs.Matches = matches

		if len(matches) == 0 && game.TechLooseTeamID == nil {
			logger.Error().Msg("matches not found, tech loose is nil")
			return "не может быть завершён: игра не содержит матчи и не указано техническое поражение"
		}

		// суммируем все очки матчей для каждой команды
		var totalScoreTeam1 int64
		var totalScoreTeam2 int64
		for _, m := range matches {
			totalScoreTeam1 += m.ScoreTeam1
			totalScoreTeam2 += m.ScoreTeam2

			if extraPointsTeam1, ok := teamsExtraPoints[m.Team1ID]; ok && extraPointsTeam1 > 0 {
				totalScoreTeam1 += extraPointsTeam1
			} else if extraPointsTeam2, ok2 := teamsExtraPoints[m.Team2ID]; ok2 && extraPointsTeam2 > 0 {
				totalScoreTeam2 += extraPointsTeam2
			}
		}

		// если у кого-то есть преимущество по очкам - то он победитель, иначе - ничего не делаем
		if totalScoreTeam1 > totalScoreTeam2 {
			gs.WinnerId = game.Team1ID
		} else if totalScoreTeam2 > totalScoreTeam1 {
			gs.WinnerId = game.Team2ID
		}

		gameStats = append(gameStats, gs)
	}

	var winners []int64
	if currenStageNumber == constant.FirstStage {
		// для оконченного regular-этапа надо понять, какие топ N команд пройдут в первый playoff-этап (т.е. второй этап)
		winners, err = helpers.ExtractWinnersRegularPlayoff(gameStats, teamsExtraPoints)
		if err != nil {
			return "не может быть завершён: " + err.Error()
		}
	} else {
		// для всех остальных playoff этапов
		winners = helpers.ExtractWinners(gameStats)
	}

	if int(currenStageNumber)+1 <= int(stageQty) {
		nextStageNumber := currenStageNumber + 1

		err = helpers.CheckQtyTeamsInRegularPlayoffStage(int(tournament.Rules.RegularPlayoff.PlayoffTeamsCountOnStart), len(winners), int(nextStageNumber), int(stageQty))
		if err != nil {
			return "не может быть завершён: " + err.Error()
		}
	}

	return "этап может быть завершён"
}

func (s *Service) checkGameForMatches(ctx context.Context, logger zerolog.Logger, gameId int64) error {
	matches, err := s.rdbOperations.GetMatchListByGameID(ctx, logger, gameId)
	if err != nil {
		return err
	}

	for _, match := range matches {
		if time.Now().After(match.Date) {
			msg := "у игры есть начатые матчи"
			err := errors.New(msg)
			logger.Error().Err(err).Msg(err.Error())

			return error_templates.New(err.Error(), err, codes.FailedPrecondition, http.StatusBadRequest)
		}
	}

	return nil
}

func (s *Service) checkBarPlaceByTournamentCity(ctx context.Context, logger zerolog.Logger, placeId, tournamentCityId int64) error {
	// проверяем соответствует ли столоместо городу турнира
	place, err := s.rdbOperations.GetPlaceByID(logger, ctx, placeId, &s.config.RDB)
	if err != nil {
		return err
	}

	place.Bar.City, err = s.rdbOperations.GetCityByBarId(logger, ctx, place.Bar.ID, &s.config.RDB)
	if err != nil {
		return err
	}

	if place.Bar.City.ID != tournamentCityId {
		err = errors.New("город турнира и города столоместа не совпадают")
		logger.Error().Err(err).Msg("place city not equal tournament city")
		return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	return nil
}

// checkTeamsByStageRegularPlayoff проверяет, что команды из реквеста есть в этапе regular+playoff турнира,
// так как в этапе могут участвовать не все команды, а только N лучших из предыдущего этапа
func (s *Service) checkTeamsByStageRegularPlayoff(ctx context.Context, logger zerolog.Logger, reqTeam1Id, reqTeam2Id, stageId int64) error {
	teams, err := s.rdbOperations.GetUniqueTeamIdsByStage(ctx, logger, stageId)
	if err != nil {
		return err
	}

	msgErr := "команда id=%d или id=%d не участвует в этапе id=%d"

	if !slices.Contains(teams, reqTeam1Id) || !slices.Contains(teams, reqTeam2Id) {
		msg := fmt.Sprintf(msgErr, reqTeam1Id, reqTeam2Id, stageId)
		err := errors.New(msg)
		logger.Error().Err(err).Msg(err.Error())

		return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	stageGames, err := s.rdbOperations.GetTournamentStageGames(ctx, logger, stageId)
	if err != nil {
		logger.Error().Err(err).Msg(err.Error())
		return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	var found bool
	for _, stGame := range stageGames {
		if stGame.Team1ID == reqTeam1Id && stGame.Team2ID == reqTeam2Id {
			found = true
		}
	}

	if !found {
		msg := "эти команды не играют друг с другом в текущем этапе"
		err := errors.New(msg)
		logger.Error().Err(err).Msg(msg)

		return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	return nil
}

func (s *Service) checkRole(
	ctx context.Context,
	logger zerolog.Logger,
	tournament entities.Tournament,
	creator entities.User,
	team1Id,
	team2Id int64,
) error {
	// проверяем относится ли капитан к какой-либо из команд
	if creator.Role.Name == constant.CaptainRole {
		if err := s.validateCaptain(ctx, logger, creator.ID, creator.Team.ID, team1Id, team2Id); err != nil {
			return err
		}
	}

	// проверяем город мастера по турнирам на соответствие городу турнира
	if creator.Role.Name == constant.TournamentMaster {
		if err := s.validateTournamentMaster(ctx, logger, &tournament, creator.ID); err != nil {
			return err
		}
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
func checkTournamentStage(stages []entities.TournamentStage, targetStageId int64) (entities.TournamentStage, error) {
	isContains := false
	var currentStage entities.TournamentStage
	var previousStage *entities.TournamentStage

	for idx, stage := range stages {
		if idx > 0 {
			previousStage = &stages[idx-1]
		}

		if stage.ID == targetStageId {
			if stage.IsFinished == true {
				return entities.TournamentStage{}, errors.New("этап завершен")
			}

			isContains = true
			currentStage = stage
			break
		}
	}

	if isContains == false {
		return entities.TournamentStage{}, errors.New("этап не найден")
	}

	if previousStage != nil && !previousStage.IsFinished {
		return entities.TournamentStage{}, errors.New("предыдущий этап не завершен")
	}

	return currentStage, nil
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

// checkTournamentGamePairs проверяет, изменились ли составы команды
func checkTournamentGamePairs(logger zerolog.Logger, team1ID, team2ID int64, game *entities.TournamentGame) error {
	if (team1ID != game.Team1ID && team1ID != game.Team2ID) || (team2ID != game.Team1ID && team2ID != game.Team2ID) {
		err := errors.New("в игре турнира нельзя менять состав команд")
		logger.Error().Err(err).Msg("error on change game teams pair")
		return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	return nil
}

func checkDeleteTournamentGame(logger zerolog.Logger, game entities.TournamentGame) error {
	if !game.IsTiebreak {
		err := errors.New("у турнира можно удалить только tiebreak игру")
		logger.Error().Err(err).Msg("not tiebreak in delete tournament game")
		return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	return nil
}

func checkTiebreakGame(logger zerolog.Logger, reqIsTiebreak bool) error {
	if reqIsTiebreak != true {
		err := errors.New("можно создать дополнительно только tiebreak игру")
		logger.Error().Err(err).Msg("not tiebreak")
		return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	return nil
}

func checkTeams(tournament entities.Tournament, team1Id, team2Id int64) error {
	// проверяем все ли команды участвуют в этом турнире
	if (slices.Contains(tournament.TeamIDs, team1Id) && slices.Contains(tournament.TeamIDs, team2Id)) == false {
		err := errors.New("команда не участвует в турнире")
		return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	// проверяем что команды разные
	if team1Id == team2Id {
		err := errors.New("id команд не могут быть одинаковыми")
		return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	return nil
}
