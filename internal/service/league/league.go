package league

import (
	"context"
	"fmt"
	"slices"

	stderr "errors"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/constant"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/calculator"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/errors"
)

func (s *Service) GetList(ctx context.Context, cityID int64) ([]entities.League, error) {
	logger := s.logger.With().Interface("service", "GetList").Logger()

	leagues, err := s.rdbOperations.GetLeagueList(logger, ctx, cityID)
	if err != nil {
		return nil, err
	}

	return leagues, nil
}

func (s *Service) Create(ctx context.Context, league entities.League, teams []int64) (*int64, error) {
	logger := s.logger.With().Interface("service", "Create").Logger()

	id, err := s.rwdbOperations.CreateLeague(logger, ctx, league, teams)
	if err != nil {
		return nil, err
	}

	return &id, nil
}

func (s *Service) Update(ctx context.Context, league entities.League, teams []int64) error {
	logger := s.logger.With().Interface("service", "Update").Logger()

	oldTeams, err := s.rdbOperations.GetTeamsByLeague(logger, ctx, league.ID)
	if err != nil {
		return err
	}

	newTeams := teams
	var teamsToDelete []int64

	for _, oldTeam := range oldTeams {
		if !slices.Contains(newTeams, oldTeam.Id) {
			teamsToDelete = append(teamsToDelete, oldTeam.Id)
		}
	}

	for _, teamId := range teamsToDelete {
		games, err := s.rdbOperations.GetPastGamesByTeamAndLeague(logger, ctx, int(teamId), int(league.ID))
		if err != nil {
			return err
		}

		if len(games) > 0 {
			err = stderr.New(fmt.Sprintf(errors.ErrDeleteTeamFromLeague, teamId, league.ID))
			return err
		}
	}

	err = s.rwdbOperations.UpdateLeague(logger, ctx, league, teams)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	logger := s.logger.With().Interface("service", "Delete").Logger()

	err := s.rwdbOperations.DeleteLeague(logger, ctx, id)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Recalc(ctx context.Context, id int64) error {
	logger := s.logger.With().Interface("service", "Recalc").Logger()

	// Get all players from league
	playerIDs, err := s.rdbOperations.GetPlayerIDsByLeagueID(logger, ctx, id)
	if err != nil {
		return err
	}

	// Reset ratings for all palyers from league
	ratings := make(map[int64]entities.Rating, len(playerIDs))
	for _, playerID := range playerIDs {
		rating := &entities.Rating{
			PlayerID: playerID,
			LeagueID: id,
			Value:    int64(constant.DefaultRating),
		}
		ratings[playerID] = *rating
	}

	// Get all matches from league
	matches, err := s.rdbOperations.GetMatchListByLeagueID(logger, ctx, id)
	if err != nil {
		return err
	}

	// Recalc and update all matches
	for i, match := range matches {
		var _rating11, _rating12, _rating21, _rating22 int64

		_, ok := ratings[int64(*match.Player1Team1ID)]
		if ok {
			_rating11 = ratings[int64(*match.Player1Team1ID)].Value
		} else {
			rating := &entities.Rating{
				PlayerID: int64(*match.Player1Team1ID),
				LeagueID: id,
				Value:    int64(constant.DefaultRating),
			}
			ratings[int64(*match.Player1Team1ID)] = *rating
			_rating11 = rating.Value
		}

		_, ok = ratings[int64(*match.Player1Team2ID)]
		if ok {
			_rating12 = ratings[int64(*match.Player1Team2ID)].Value
		} else {
			rating := &entities.Rating{
				PlayerID: int64(*match.Player1Team2ID),
				LeagueID: id,
				Value:    int64(constant.DefaultRating),
			}
			ratings[int64(*match.Player1Team2ID)] = *rating
			_rating12 = rating.Value
		}

		_rating21 = 0
		if match.Player2Team1ID != nil && *match.Player2Team1ID > 0 {
			_, ok = ratings[int64(*match.Player2Team1ID)]
			if ok {
				_rating21 = ratings[int64(*match.Player2Team1ID)].Value
			} else {
				rating := &entities.Rating{
					PlayerID: int64(*match.Player2Team1ID),
					LeagueID: id,
					Value:    int64(constant.DefaultRating),
				}
				ratings[int64(*match.Player2Team1ID)] = *rating
				_rating21 = rating.Value
			}
		}

		_rating22 = 0
		if match.Player2Team2ID != nil && *match.Player2Team2ID > 0 {
			_, ok = ratings[int64(*match.Player2Team2ID)]
			if ok {
				_rating22 = ratings[int64(*match.Player2Team2ID)].Value
			} else {
				rating := &entities.Rating{
					PlayerID: int64(*match.Player2Team2ID),
					LeagueID: id,
					Value:    int64(constant.DefaultRating),
				}
				ratings[int64(*match.Player2Team2ID)] = *rating
				_rating22 = rating.Value
			}
		}

		matches[i].Player1Team1RateBefore = &_rating11
		matches[i].Player1Team2RateBefore = &_rating12
		matches[i].Player2Team1RateBefore = &_rating21
		matches[i].Player2Team2RateBefore = &_rating22

		rating11, rating12, rating21, rating22, err := calculator.MatchRaitingCalculation(ctx, int(*match.ScoreTeam1), int(*match.ScoreTeam2), int(_rating11), int(_rating12), int(_rating21), int(_rating22))
		if err != nil {
			return err
		}

		var zero int64
		matches[i].Player2Team1RateAfter = &zero
		matches[i].Player2Team2RateAfter = &zero

		r11 := int64(rating11)
		matches[i].Player1Team1RateAfter = &r11
		r12 := int64(rating12)
		matches[i].Player1Team2RateAfter = &r12

		ratings[int64(*match.Player1Team1ID)] = entities.Rating{
			PlayerID: int64(*match.Player1Team1ID),
			LeagueID: id,
			Value:    int64(rating11),
		}
		ratings[int64(*match.Player1Team2ID)] = entities.Rating{
			PlayerID: int64(*match.Player1Team2ID),
			LeagueID: id,
			Value:    int64(rating12),
		}
		if match.Player2Team1ID != nil && *match.Player2Team1ID > 0 {
			r21 := int64(rating21)
			matches[i].Player2Team1RateAfter = &r21

			ratings[int64(*match.Player2Team1ID)] = entities.Rating{
				PlayerID: int64(*match.Player2Team1ID),
				LeagueID: id,
				Value:    int64(rating21),
			}
		}
		if match.Player2Team2ID != nil && *match.Player2Team2ID > 0 {
			r22 := int64(rating22)
			matches[i].Player2Team2RateAfter = &r22

			ratings[int64(*match.Player2Team2ID)] = entities.Rating{
				PlayerID: int64(*match.Player2Team2ID),
				LeagueID: id,
				Value:    int64(rating22),
			}
		}
	}

	err = s.rwdbOperations.RewriteMatchesAndPlayerRatings(logger, ctx, matches, ratings)
	if err != nil {
		return err
	}

	return nil
}

///////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////

func (s *Service) CreateExtraPoints(ctx context.Context, req *entities.CreateExtraPointsRequest) (int64, error) {
	logger := s.logger.With().Str("service", "CreateExtraPoints").Logger()

	id, err := s.rwdbOperations.CreateExtraPoints(logger, ctx, req)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *Service) UpdateExtraPoints(ctx context.Context, req *entities.UpdateExtraPointsRequest) (bool, error) {
	logger := s.logger.With().Str("service", "UpdateExtraPoints").Logger()

	res, err := s.rwdbOperations.UpdateExtraPoints(logger, ctx, req)
	if err != nil {
		return res, err
	}

	return res, err
}

func (s *Service) DeleteExtraPoints(ctx context.Context, extraPointsId int64) (bool, error) {
	logger := s.logger.With().Str("service", "DeleteExtraPoints").Logger()

	res, err := s.rwdbOperations.DeleteExtraPoints(logger, ctx, extraPointsId)
	if err != nil {
		return res, err
	}

	return res, err
}

func (s *Service) GetExtraPointsListByTeamAndLeagueId(ctx context.Context, teamId, leagueId int64) ([]entities.ExtraPoints, error) {
	logger := s.logger.With().Str("service", "GetExtraPointsListByTeamAndLeagueId").Logger()

	res, err := s.rdbOperations.GetExtraPointsListByTeamAndLeagueId(logger, ctx, teamId, leagueId)
	if err != nil {
		return res, err
	}

	return res, nil
}

func (s *Service) GetExtraPointsById(ctx context.Context, extraPointsId int64) (entities.ExtraPoints, error) {
	logger := s.logger.With().Str("service", "GetExtraPointsById").Logger()

	res, err := s.rdbOperations.GetExtraPointsById(logger, ctx, extraPointsId)
	if err != nil {
		return res, err
	}

	return res, nil
}
