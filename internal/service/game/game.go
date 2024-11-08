package game

import (
	"context"
	stderr "errors"
	"net/http"

	"google.golang.org/grpc/codes"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/entity"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/calculator"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/error_templates"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/errors"
)

func (s *Service) Create(ctx context.Context, request entities.CreateGameRequest) (entities.CreateGameResponse, error) {
	logger := s.logger.With().Interface("service", "game.Create").Logger()
	timeout, cancel := context.WithTimeout(ctx, s.config.RWDB.MaxIdleConnectionTimeout)
	defer cancel()

	// Getting league for teams and validate it
	team1resp, err := s.rdbOperations.GetTeam(logger, ctx, int64(request.Team1ID))
	if err != nil {
		return entities.CreateGameResponse{}, err
	}
	team2resp, err := s.rdbOperations.GetTeam(logger, ctx, int64(request.Team2ID))
	if err != nil {
		return entities.CreateGameResponse{}, err
	}
	if int64(*team1resp.LeagueId) != int64(*team2resp.LeagueId) {
		logger.Error().Stack().Err(err).Msg("failed to postgresql.GetTeam")
		err = stderr.New(errors.FailedGameByTeamsLeagueMismatch)
		return entities.CreateGameResponse{}, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	// Map for keeping player ratings while going game calculation
	rates := make(map[int]int, 0)

	leagueID := int64(*team1resp.LeagueId)
	var matches []entities.GamesMatch

	for _, match := range request.Matches {
		if match.ScoreTeam1 == 0 && match.ScoreTeam2 == 0 {
			continue
		}

		player1team1rate := 0
		player1team1ID := match.Player1Team1Id
		value, ok := rates[player1team1ID]
		if ok {
			player1team1rate = value
		} else {
			resp, err := s.rdbOperations.GetRatingByPlayerIDAndByLeagueID(logger, ctx, int64(player1team1ID), leagueID)
			if err != nil {
				if outputError, ok := (err).(*error_templates.OutputError); ok {
					code, _ := outputError.GetHTTP()
					if code == http.StatusNotFound {
						resp = 1000
					} else {
						return entities.CreateGameResponse{}, err
					}
				} else {
					return entities.CreateGameResponse{}, err
				}
			}
			player1team1rate = int(resp)
		}

		player1team2rate := 0
		player1team2ID := match.Player1Team2Id
		value, ok = rates[player1team2ID]
		if ok {
			player1team2rate = value
		} else {
			resp, err := s.rdbOperations.GetRatingByPlayerIDAndByLeagueID(logger, ctx, int64(player1team2ID), leagueID)
			if err != nil {
				if outputError, ok := (err).(*error_templates.OutputError); ok {
					code, _ := outputError.GetHTTP()
					if code == http.StatusNotFound {
						resp = 1000
					} else {
						return entities.CreateGameResponse{}, err
					}
				} else {
					return entities.CreateGameResponse{}, err
				}
			}
			player1team2rate = int(resp)
		}

		player2team1rate := 0
		player2team1ID := *match.Player2Team1Id
		if player2team1ID > 0 {
			value, ok = rates[player2team1ID]
			if ok {
				player2team1rate = value
			} else {
				resp, err := s.rdbOperations.GetRatingByPlayerIDAndByLeagueID(logger, ctx, int64(player2team1ID), leagueID)
				if err != nil {
					if outputError, ok := (err).(*error_templates.OutputError); ok {
						code, _ := outputError.GetHTTP()
						if code == http.StatusNotFound {
							resp = 1000
						} else {
							return entities.CreateGameResponse{}, err
						}
					} else {
						return entities.CreateGameResponse{}, err
					}
				}
				player2team1rate = int(resp)
			}
		}

		player2team2rate := 0
		player2team2ID := *match.Player2Team2Id
		if player2team2ID > 0 {
			value, ok = rates[player2team2ID]
			if ok {
				player2team2rate = value
			} else {
				resp, err := s.rdbOperations.GetRatingByPlayerIDAndByLeagueID(logger, ctx, int64(player2team2ID), leagueID)
				if err != nil {
					if outputError, ok := (err).(*error_templates.OutputError); ok {
						code, _ := outputError.GetHTTP()
						if code == http.StatusNotFound {
							resp = 1000
						} else {
							return entities.CreateGameResponse{}, err
						}
					} else {
						return entities.CreateGameResponse{}, err
					}
				}
				player2team2rate = int(resp)
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

		player1team1rateAfter, player1team2rateAfter, player2team1rateAfter, player2team2rateAfter, err := calculator.MatchRaitingCalculation(
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

	operator := "insertIgnore"

	resp, err := s.rwdbOperations.CreateGameWithRating(logger, timeout, request, rates, &operator, leagueID)
	if err != nil {
		return entities.CreateGameResponse{}, err
	}

	return resp, nil
}

func (s *Service) Delete(ctx context.Context, gameID int) error {
	logger := s.logger.With().Interface("service", "game.Delete").Logger()
	timeout, cancel := context.WithTimeout(ctx, s.config.RWDB.MaxIdleConnectionTimeout)
	defer cancel()

	recalc := true
	game, err := s.rdbOperations.GetGame(logger, ctx, gameID)
	if err != nil {
		return err
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
	if team1resp.LeagueId == nil || team2resp.LeagueId == nil ||
		(team1resp.LeagueId != nil && team2resp.LeagueId != nil && int64(*team1resp.LeagueId) != int64(*team2resp.LeagueId)) {
		recalc = false
	}

	if recalc {
		leagueID := int64(*team1resp.LeagueId)

		// Decreasing rating for each player of game matches
		plGmRtInc := make(map[int]int, 0) // playerGameRatingIncrease
		_matches, err := s.rdbOperations.GetMatchListByGameID(ctx, logger, game.ID)
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

			rate := &entity.Rating{
				PlayerID: int64(playerID),
				LeagueID: leagueID,
				Value:    rateValue - int64(value),
			}
			operator := "insertIgnore"
			err = s.rwdbOperations.CreateRating(logger, ctx, *rate, &operator)
			if err != nil {
				return err
			}
		}
	}

	err = s.rwdbOperations.DeleteGame(logger, timeout, gameID)
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

	// Getting league for teams and validate it
	team1resp, err := s.rdbOperations.GetTeam(logger, ctx, int64(request.Team1ID))
	if err != nil {
		return err
	}
	team2resp, err := s.rdbOperations.GetTeam(logger, ctx, int64(request.Team2ID))
	if err != nil {
		return err
	}
	if int64(*team1resp.LeagueId) != int64(*team2resp.LeagueId) {
		logger.Error().Stack().Err(err).Msg("failed to postgresql.GetTeam")
		err = stderr.New(errors.FailedGameByTeamsLeagueMismatch)
		return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	leagueID := int64(*team1resp.LeagueId)
	rates := make(map[int]int, 0)

	// Decreasing rating for each player of game matches
	plGmRtInc := make(map[int]int, 0) // playerGameRatingIncrease
	_matches, err := s.rdbOperations.GetMatchListByGameID(ctx, logger, request.ID)
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

		player1team1rate := 0
		player1team1ID := match.Player1Team1Id
		val, ok := rates[player1team1ID]
		if ok {
			player1team1rate = val
		} else {
			resp, err := s.rdbOperations.GetRatingByPlayerIDAndByLeagueID(logger, ctx, int64(player1team1ID), leagueID)
			if err != nil {
				if outputError, ok := (err).(*error_templates.OutputError); ok {
					code, _ := outputError.GetHTTP()
					if code == http.StatusNotFound {
						resp = 1000
					} else {
						return err
					}
				} else {
					return err
				}
			}
			player1team1rate = int(resp)
		}

		player1team2rate := 0
		player1team2ID := match.Player1Team2Id
		val, ok = rates[player1team2ID]
		if ok {
			player1team2rate = val
		} else {
			resp, err := s.rdbOperations.GetRatingByPlayerIDAndByLeagueID(logger, ctx, int64(player1team2ID), leagueID)
			if err != nil {
				if outputError, ok := (err).(*error_templates.OutputError); ok {
					code, _ := outputError.GetHTTP()
					if code == http.StatusNotFound {
						resp = 1000
					} else {
						return err
					}
				} else {
					return err
				}
			}
			player1team2rate = int(resp)
		}

		player2team1rate := 0
		player2team1ID := *match.Player2Team1Id
		if player2team1ID > 0 {
			val, ok = rates[player2team1ID]
			if ok {
				player2team1rate = val
			} else {
				resp, err := s.rdbOperations.GetRatingByPlayerIDAndByLeagueID(logger, ctx, int64(player2team1ID), leagueID)
				if err != nil {
					if outputError, ok := (err).(*error_templates.OutputError); ok {
						code, _ := outputError.GetHTTP()
						if code == http.StatusNotFound {
							resp = 1000
						} else {
							return err
						}
					} else {
						return err
					}
				}
				player2team1rate = int(resp)
			}
		}

		player2team2rate := 0
		player2team2ID := *match.Player2Team2Id
		if player2team2ID > 0 {
			val, ok = rates[player2team2ID]
			if ok {
				player2team2rate = val
			} else {
				resp, err := s.rdbOperations.GetRatingByPlayerIDAndByLeagueID(logger, ctx, int64(player2team2ID), leagueID)
				if err != nil {
					if outputError, ok := (err).(*error_templates.OutputError); ok {
						code, _ := outputError.GetHTTP()
						if code == http.StatusNotFound {
							resp = 1000
						} else {
							return err
						}
					} else {
						return err
					}
				}
				player2team2rate = int(resp)
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

		player1team1rateAfter, player1team2rateAfter, player2team1rateAfter, player2team2rateAfter, err := calculator.MatchRaitingCalculation(
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

	newRates := make([]entity.Rating, 0, len(rates))
	for playerID, value := range rates {
		rate := &entity.Rating{
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

func (s *Service) UpdateFutureGame(ctx context.Context, request entity.UpdateFutureGameRequest) error {
	logger := s.logger.With().Interface("service", "game.UpdateFutureGame").Logger()
	timeout, cancel := context.WithTimeout(ctx, s.config.RWDB.MaxIdleConnectionTimeout)
	defer cancel()

	err := s.rwdbOperations.UpdateFutureGame(logger, timeout, request)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) GetGamesYears(ctx context.Context) (entity.GetGamesYearsResponse, error) {
	logger := s.logger.With().Interface("service", "game.GetGamesYears").Logger()
	timeout, cancel := context.WithTimeout(ctx, s.config.RDB.MaxIdleConnectionTimeout)
	defer cancel()

	resp, err := s.rdbOperations.GetGamesYears(logger, timeout)
	if err != nil {
		return entity.GetGamesYearsResponse{}, err
	}
	return resp, nil
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

func (s *Service) GetFutureGames(ctx context.Context, cityID int) (entity.GetFutureGamesResponse, error) {
	logger := s.logger.With().Interface("service", "game.GetFutureGames").Logger()
	timeout, cancel := context.WithTimeout(ctx, s.config.RDB.MaxIdleConnectionTimeout)
	defer cancel()

	games, err := s.rdbOperations.GetFutureGames(logger, timeout, cityID)
	if err != nil {
		return entity.GetFutureGamesResponse{}, err
	}

	return entity.GetFutureGamesResponse{Games: games}, nil
}

func (s *Service) CreateFutureGame(ctx context.Context, request entity.CreateFutureGameRequest) (entity.CreateFutureGameResponse, error) {
	logger := s.logger.With().Interface("service", "game.CreateFutureGame").Logger()
	timeout, cancel := context.WithTimeout(ctx, s.config.RWDB.MaxIdleConnectionTimeout)
	defer cancel()

	id, err := s.rwdbOperations.CreateFutureGame(logger, timeout, request)
	if err != nil {
		return entity.CreateFutureGameResponse{}, err
	}

	return entity.CreateFutureGameResponse{ID: id}, nil
}

func (s *Service) GetTeamGames(ctx context.Context, teamID int) (entity.GetTeamGamesResponse, error) {
	logger := s.logger.With().Interface("service", "game.GetTeamGames").Logger()
	timeout, cancel := context.WithTimeout(ctx, s.config.RDB.MaxIdleConnectionTimeout)
	defer cancel()

	games, err := s.rdbOperations.GetTeamGames(logger, timeout, teamID)
	if err != nil {
		return entity.GetTeamGamesResponse{}, nil
	}

	return entity.GetTeamGamesResponse{Games: games}, nil
}
