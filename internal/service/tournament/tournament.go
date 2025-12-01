package tournament

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"slices"
	"strconv"
	"strings"
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

const (
	indexZero int    = 0
	oneStage  string = "1/1"
)

func (s *Service) GetTournamentTypeList(ctx context.Context, withDeleted bool) ([]entities.TournamentType, error) {
	logger := s.logger.With().Str("service", "GetTournamentTypeList").Logger()

	list, err := s.rdbOperations.GetTournamentTypeList(logger, ctx, withDeleted, &s.config.RDB)
	if err != nil {
		return nil, err
	}

	return list, nil
}

func (s *Service) Create(ctx context.Context, request *entities.CreateTournamentRequest) (int64, error) {
	logger := s.logger.With().Str("service", "Create").Logger()

	var err error

	// валидируем соответствие типа турнира входящим параметрам
	if err = validateRulesTypesIds(request.Rules, request.TournamentTypeID, request.TeamsIDs, request.PlayersIDs); err != nil {
		logger.Error().Err(err).Msg("failed to set tournament type")
		return 0, err
	}

	// если запрос от мастера по турнирам, то тщательно проверяем соответствие города мастера
	// городу из запроса, которого ожидается, что не будет указано,но на всякий случай
	if request.Creator.Role.Name == constant.TournamentMaster {
		master, err := s.rdbOperations.GetTournamentMasterByUserId(logger, ctx, request.Creator.ID)
		if err != nil {
			return 0, err
		}

		if request.CityID != nil && *request.CityID != master.City.ID {
			err = errors.New(pkgerr.ErrCityIdNotEqualMasterCityId)
			logger.Error().Err(err).Msg("cityIdParam not equal masterCityId in tournament.Create")
			return 0, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}

		if request.CityID == nil {
			request.CityID = &master.City.ID
		}
	} else {
		if request.CityID == nil {
			err = errors.New(pkgerr.EmptyParameterError + ": cityId")
			logger.Error().Err(err).Msg("cityId is nil")
			return 0, err
		}
		if request.Creator.Role.Name != constant.SuperUserRole && request.Creator.Role.Name != constant.AdminRole {
			err = errors.New(pkgerr.WrongUserRole)
			logger.Error().Err(err).Msgf("forbidden for this role: %s", request.Creator.Role.Name)
			return 0, error_templates.New(err.Error(), err, codes.Unauthenticated, http.StatusForbidden)
		}
	}

	// не позволяем в турнире участвовать команде из другого города
	for _, tId := range request.TeamsIDs {
		team, err := s.rdbOperations.GetTeamById(logger, ctx, tId, nil)
		if err != nil {
			return 0, err
		}

		if pointer.GetValue(team.CityId) != pointer.GetValue(request.CityID) {
			err = fmt.Errorf("id города команды %s(%d) не совпадает с id города турнира(%d)", team.Name, pointer.GetValue(team.CityId), request.CityID)
			return 0, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}
	}

	// создание турнира типа Regular каждый с каждым
	if request.TournamentTypeID == constant.RegularTournamentTypeID {
		tournamentID, err := s.createRegular(ctx, logger, *request)
		if err != nil {
			return 0, err
		}

		return tournamentID, nil

		// создание турнира типа Regular.OneVsOne каждый с каждым по одному игроку в команде
	} else if request.TournamentTypeID == constant.RegularOneVsOneTournamentTypeID {
		tournamentID, err := s.createRegularOneVsOne(ctx, logger, *request)
		if err != nil {
			return 0, err
		}

		return tournamentID, nil

		// создание турнира типа Playoff
		// автоматом создаются все этапы турнира и все игры первого этапа турнира
	} else if request.TournamentTypeID == constant.PlayoffTournamentTypeID {
		tournamentID, err := s.createPlayoff(ctx, logger, *request)
		if err != nil {
			return 0, err
		}

		return tournamentID, nil

	} else if request.TournamentTypeID == constant.RegularPlayoffTournamentTypeID {
		tournamentID, err := s.createRegularPlayoff(ctx, logger, *request)
		if err != nil {
			return 0, err
		}

		return tournamentID, nil

	} else { //временная ошибка, создание других турниров реализуется позже
		return 0, errors.New("not implemented")
	}
}

func (s *Service) Update(ctx context.Context, request *entities.UpdateTournamentRequest) error {
	logger := s.logger.With().Str("service", "Update").Logger()

	tournament, err := s.rdbOperations.GetTournamentById(ctx, logger, request.ID)
	if err != nil {
		return err
	}

	request.TournamentTypeID = tournament.TypeID

	games, err := s.rdbOperations.GetTournamentGameList(logger, ctx, &entities.GetTournamentGameList{TournamentId: &tournament.ID})
	if err != nil {
		return err
	}

	haveMatches := false
	for _, game := range games {
		matches, err := s.rdbOperations.FetchMatches(logger, ctx, game.ID)
		if err != nil {
			return err
		}

		if len(matches) > 0 {
			haveMatches = true
			break
		}
	}

	// обновление мастером возможно только в неначатых турнирах
	if request.Executor.Role.Name == constant.TournamentMaster {
		err = s.checkTournamentMasterCredentials(ctx, logger, request, tournament, games)
		if err != nil {
			return err
		}
	} else {
		// если это админы и турнир начат, то можно обновить минимальную инфу (имя турнира, сезон) без перетасовки команд и изменения типа и сетки турнира
		if haveFinishedStage(tournament.Stages) == true || haveStartedGames(games) == true || haveMatches == true {
			if request.CityID != nil || request.Rules != nil || request.TeamsIDs != nil || request.PlayersIDs != nil {
				err = fmt.Errorf("турнир с завершенными или начатыми играми не могут обновляться поля: cityId, rules, teamIds, playersIds")
				logger.Error().Err(err).Msg("Failed tournament.Update: finished stage or started games")
				return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
			}

			err = s.rwdbOperations.UpdateTournament(logger, ctx, *request, nil)
			if err != nil {
				return err
			}

			return nil
		}
	}

	if request.TournamentTypeID == constant.RegularTournamentTypeID {
		err = s.updateRegular(ctx, logger, *request, tournament)
		if err != nil {
			return err
		}

		return nil
	} else if request.TournamentTypeID == constant.RegularOneVsOneTournamentTypeID {
		err = s.updateRegularOneVsOne(ctx, logger, *request, tournament)
		if err != nil {
			return err
		}

		return nil
	} else if request.TournamentTypeID == constant.PlayoffTournamentTypeID {
		err = s.updatePlayoff(ctx, logger, *request, tournament)
		if err != nil {
			return err
		}

		return nil
	} else if request.TournamentTypeID == constant.RegularPlayoffTournamentTypeID {
		err = s.updateRegularPlayoff(ctx, logger, *request, tournament)
		if err != nil {
			return err
		}

		return nil
	}

	return errors.New("not implemented")
}

func (s *Service) Delete(ctx context.Context, request *entities.DeleteTournamentRequest) error {
	logger := s.logger.With().Str("service", "Delete").Logger()

	var err error
	tournament, err := s.rdbOperations.GetTournamentById(ctx, logger, request.ID)
	if err != nil {
		return err
	}

	if request.Executor.Role.Name != constant.SuperUserRole && request.Executor.Role.Name != constant.AdminRole && request.Executor.Role.Name != constant.TournamentMaster {
		err = errors.New(pkgerr.WrongUserRole)
		logger.Error().Err(err).Msgf("forbidden for this role: %s", request.Executor.Role.Name)
		return error_templates.New(err.Error(), err, codes.Unauthenticated, http.StatusForbidden)
	}

	if request.Executor.Role.Name == constant.TournamentMaster {
		err = s.checkTournamentMasterCredentialsOnDelete(ctx, logger, request, tournament)
		if err != nil {
			return err
		}
	}

	tx, err := s.rwdbOperations.BeginTx(ctx, logger)
	if err != nil {
		return err
	}

	// удаляем рейтинг
	if err = s.rwdbOperations.DeleteTournamentRating(logger, ctx, tournament.ID, tx); err != nil {
		tx.Rollback(ctx)
		return err
	}

	// отвязываем команды
	if err = s.rwdbOperations.UnlinkTeamsFromTournament(logger, ctx, tournament.ID, tx); err != nil {
		tx.Rollback(ctx)
		return err
	}

	// удалить матчи
	if err = s.rwdbOperations.DeleteTournamentMatches(logger, ctx, tournament.ID, tx); err != nil {
		tx.Rollback(ctx)
		return err
	}

	// удаляем игры
	if err = s.rwdbOperations.DeleteTournamentGames(logger, ctx, tournament.ID, tx); err != nil {
		tx.Rollback(ctx)
		return err
	}

	// удаляем этапы
	if err = s.rwdbOperations.DeleteTournamentStages(logger, ctx, tournament.ID, tx); err != nil {
		tx.Rollback(ctx)
		return err
	}

	// удаляем доп.очки
	if err = s.rwdbOperations.DeleteExtraPointsByTournamentId(ctx, logger, tournament.ID, tx); err != nil {
		tx.Rollback(ctx)
		return err
	}

	// удаляем турнир
	if err = s.rwdbOperations.DeleteTournament(logger, ctx, tournament.ID, tx); err != nil {
		tx.Rollback(ctx)
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		tx.Rollback(ctx)
		return err
	}

	return nil
}

func (s *Service) FinishStage(ctx context.Context, request *entities.FinishStageRequest) error {
	logger := s.logger.With().Str("service", "FinishStage").Logger()

	stage, err := s.rdbOperations.GetTournamentStage(ctx, logger, request.ID)
	if err != nil {
		return err
	}

	if stage.IsFinished == true {
		err = errors.New("этап уже завершён")
		logger.Error().Err(err).Msg("stage already finished")
		return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	tournament, err := s.rdbOperations.GetTournamentById(ctx, logger, stage.TournamentID)
	if err != nil {
		return err
	}

	games, err := s.rdbOperations.GetTournamentStageGames(ctx, logger, request.ID)
	if err != nil {
		return err
	}

	// проверка соответствия города у мастера
	if request.Finisher.Role.Name == constant.TournamentMaster {
		master, err := s.rdbOperations.GetTournamentMasterByUserId(logger, ctx, request.Finisher.ID)
		if err != nil {
			return err
		}

		if master.City.ID != tournament.CityID {
			err = errors.New("город турнира и город мастера по турнирам не совпадают")
			logger.Error().Err(err).Msg("master city not equal tournament city")
			return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}
	}

	if tournament.TypeID == constant.RegularTournamentTypeID || tournament.TypeID == constant.RegularOneVsOneTournamentTypeID {
		err = s.finishStageRegular(ctx, logger, request, tournament, games)
		if err != nil {
			return err
		}

	} else if tournament.TypeID == constant.PlayoffTournamentTypeID {
		err = s.finishStagePlayoff(ctx, logger, tournament, games, stage)
		if err != nil {
			return err
		}

	} else if tournament.TypeID == constant.RegularPlayoffTournamentTypeID {
		err = s.finishStageRegularPlayoff(ctx, logger, tournament, games, stage)
		if err != nil {
			return err
		}
	} else {
		return errors.New("tournament type 4 not implemented")
	}

	tx, err := s.rwdbOperations.BeginTx(ctx, logger)
	if err != nil {
		return err
	}

	err = s.rwdbOperations.UpdateTournamentStage(ctx, logger, entities.NullableStage{ID: stage.ID, IsFinished: pointer.GetPointer(true)}, tx)
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

func (s *Service) GetTournamentStageList(ctx context.Context, id int64) ([]entities.TournamentStageItem, error) {
	logger := s.logger.With().Str("service", "GetTournamentStageList").Logger()

	stages, err := s.rdbOperations.GetTournamentStageList(ctx, logger, id)
	if err != nil {
		return nil, err
	}

	var stageItems []entities.TournamentStageItem

	for _, stage := range stages {
		games, err := s.rdbOperations.GetTournamentStageGames(ctx, logger, stage.ID)
		if err != nil {
			return nil, err
		}

		stageItems = append(stageItems, entities.TournamentStageItem{
			Stage: stage,
			Games: games,
		})
	}

	return stageItems, nil
}

func (s *Service) GetTournamentList(ctx context.Context, request *entities.GetTournamentListRequest) ([]entities.TournamentShort, int64, error) {
	logger := s.logger.With().Str("service", "GetTournamentList").Logger()

	list, count, err := s.rdbOperations.GetTournamentList(logger, ctx, *request)
	if err != nil {
		return nil, 0, err
	}

	return list, count, nil
}

func (s *Service) Recalc(ctx context.Context, id int64) error {
	logger := s.logger.With().Str("service", "Recalc").Logger()

	// Get all players from tournament
	playerIDs, err := s.rdbOperations.GetPlayerIdsByTournamentId(logger, ctx, id)
	if err != nil {
		return err
	}

	// Reset ratings for all players from tournament
	ratings := make(map[int64]entities.TournamentRating, len(playerIDs))
	for _, playerID := range playerIDs {
		rating := &entities.TournamentRating{
			PlayerID:     playerID,
			TournamentID: id,
			Value:        int64(constant.DefaultRating),
		}
		ratings[playerID] = *rating
	}

	// Get all matches from tournament
	matches, err := s.rdbOperations.GetMatchListByTournamentId(logger, ctx, id)
	if err != nil {
		return err
	}

	// Recalc and update all matches
	for i, match := range matches {
		if match.Player1Team1ID == nil || match.Player1Team2ID == nil || match.Player2Team1ID == nil || match.Player2Team2ID == nil {
			msg := fmt.Sprintf("nil player id from match id: %d", match.ID)
			return errors.New(msg)
		}

		var _rating11, _rating12, _rating21, _rating22 int64

		_, ok := ratings[int64(*match.Player1Team1ID)]
		if ok {
			_rating11 = ratings[int64(*match.Player1Team1ID)].Value
		} else {
			rating := &entities.TournamentRating{
				PlayerID:     int64(*match.Player1Team1ID),
				TournamentID: id,
				Value:        int64(constant.DefaultRating),
			}
			ratings[int64(*match.Player1Team1ID)] = *rating
			_rating11 = rating.Value
		}

		_, ok = ratings[int64(*match.Player1Team2ID)]
		if ok {
			_rating12 = ratings[int64(*match.Player1Team2ID)].Value
		} else {
			rating := &entities.TournamentRating{
				PlayerID:     int64(*match.Player1Team2ID),
				TournamentID: id,
				Value:        int64(constant.DefaultRating),
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
				rating := &entities.TournamentRating{
					PlayerID:     int64(*match.Player2Team1ID),
					TournamentID: id,
					Value:        int64(constant.DefaultRating),
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
				rating := &entities.TournamentRating{
					PlayerID:     int64(*match.Player2Team2ID),
					TournamentID: id,
					Value:        int64(constant.DefaultRating),
				}
				ratings[int64(*match.Player2Team2ID)] = *rating
				_rating22 = rating.Value
			}
		}

		matches[i].Player1Team1RateBefore = &_rating11
		matches[i].Player1Team2RateBefore = &_rating12
		matches[i].Player2Team1RateBefore = &_rating21
		matches[i].Player2Team2RateBefore = &_rating22

		rating11, rating12, rating21, rating22, err := calculator.MatchRatingCalculation(ctx, int(*match.ScoreTeam1), int(*match.ScoreTeam2), int(_rating11), int(_rating12), int(_rating21), int(_rating22))
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

		ratings[int64(*match.Player1Team1ID)] = entities.TournamentRating{
			PlayerID:     int64(*match.Player1Team1ID),
			TournamentID: id,
			Value:        int64(rating11),
		}
		ratings[int64(*match.Player1Team2ID)] = entities.TournamentRating{
			PlayerID:     int64(*match.Player1Team2ID),
			TournamentID: id,
			Value:        int64(rating12),
		}
		if match.Player2Team1ID != nil && *match.Player2Team1ID > 0 {
			r21 := int64(rating21)
			matches[i].Player2Team1RateAfter = &r21

			ratings[int64(*match.Player2Team1ID)] = entities.TournamentRating{
				PlayerID:     int64(*match.Player2Team1ID),
				TournamentID: id,
				Value:        int64(rating21),
			}
		}
		if match.Player2Team2ID != nil && *match.Player2Team2ID > 0 {
			r22 := int64(rating22)
			matches[i].Player2Team2RateAfter = &r22

			ratings[int64(*match.Player2Team2ID)] = entities.TournamentRating{
				PlayerID:     int64(*match.Player2Team2ID),
				TournamentID: id,
				Value:        int64(rating22),
			}
		}
	}

	tx, err := s.rwdbOperations.BeginTx(ctx, logger)
	if err != nil {
		return err
	}

	err = s.rwdbOperations.RewriteTournamentMatchesAndPlayerRatings(logger, ctx, matches, ratings, tx)
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

func (s *Service) StartNextStage(ctx context.Context, req *entities.StartNextStageRequest) (int64, error) {
	logger := s.logger.With().Str("service", "StartNextStage").Logger()

	tournament, err := s.rdbOperations.GetTournamentById(ctx, logger, req.TournamentID)
	if err != nil {
		return 0, err
	}

	// если запрос от мастера по турнирам, то тщательно проверяем соответствие города мастера
	// городу из запроса, которого ожидается, что не будет указано,но на всякий случай
	if req.Executor.Role.Name == constant.TournamentMaster {
		master, err := s.rdbOperations.GetTournamentMasterByUserId(logger, ctx, req.Executor.ID)
		if err != nil {
			return 0, err
		}

		if tournament.CityID != master.City.ID {
			err = errors.New(pkgerr.ErrCityIdNotEqualMasterCityId)
			logger.Error().Err(err).Msg("cityIdParam not equal masterCityId in tournament.Create")
			return 0, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}
	}

	// этот метода вернет ошибку для всех одноэтапных турниров
	if tournament.TypeID == constant.RegularTournamentTypeID || tournament.TypeID == constant.RegularOneVsOneTournamentTypeID {
		err = errors.New("у турниров типа \"Regular\" и \"RegularOneVsOne\" может быть только один этап")
		return 0, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	if tournament.TypeID == constant.PlayoffTournamentTypeID {
		stageId, err := s.startNextStagePlayoff(ctx, logger, tournament)
		if err != nil {
			return 0, err
		}

		return stageId, nil
	} else if tournament.TypeID == constant.RegularPlayoffTournamentTypeID {
		stageId, err := s.startNextStageRegularPlayoff(ctx, logger, tournament)
		if err != nil {
			return 0, err
		}

		return stageId, nil
	}

	return 0, errors.New("not implemented")
}

func (s *Service) CreateExtraPoints(ctx context.Context, req *entities.CreateExtraPointsTournamentRequest) (int64, error) {
	logger := s.logger.With().Str("service", "CreateExtraPointsTournament").Logger()

	err := s.checkMasterAndTournamentCities(ctx, req.TournamentId, req.UserId, req.Role)
	if err != nil {
		return 0, err
	}

	tournament, err := s.rdbOperations.GetTournamentById(ctx, logger, req.TournamentId)
	if err != nil {
		return 0, err
	}

	if tournament.Rules.PlayOff != nil {
		err := error_templates.BadRequestError(errors.New("для playoff турнира нельзя создать доп.очки"))

		logger.Error().Err(err).Msg(err.Error())
		return 0, err
	}

	tx, err := s.rwdbOperations.BeginTx(ctx, logger)
	if err != nil {
		return 0, err
	}

	id, err := s.rwdbOperations.CreateExtraPointsTournament(ctx, logger, req, tx)
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

func (s *Service) UpdateExtraPoints(ctx context.Context, req *entities.UpdateExtraPointsTournamentRequest) (bool, error) {
	logger := s.logger.With().Str("service", "UpdateExtraPointsTournament").Logger()

	extraPointsInfo, err := s.GetExtraPointsById(ctx, req.Id)
	if err != nil {
		return false, err
	}

	err = s.checkMasterAndTournamentCities(ctx, extraPointsInfo.TournamentId, req.UserId, req.Role)
	if err != nil {
		return false, err
	}

	tx, err := s.rwdbOperations.BeginTx(ctx, logger)
	if err != nil {
		return false, err
	}

	res, err := s.rwdbOperations.UpdateExtraPointsTournament(ctx, logger, req, tx)
	if err != nil {
		tx.Rollback(ctx)
		return false, err
	}

	if err = tx.Commit(ctx); err != nil {
		tx.Rollback(ctx)
		return false, err
	}

	return res, nil
}

func (s *Service) DeleteExtraPoints(ctx context.Context, req *entities.DeleteExtraPointsRequest) (bool, error) {
	logger := s.logger.With().Str("service", "DeleteExtraPointsTournament").Logger()

	extraPointsInfo, err := s.GetExtraPointsById(ctx, req.Id)
	if err != nil {
		return false, err
	}

	err = s.checkMasterAndTournamentCities(ctx, extraPointsInfo.TournamentId, req.UserId, req.Role)
	if err != nil {
		return false, err
	}

	tx, err := s.rwdbOperations.BeginTx(ctx, logger)
	if err != nil {
		return false, err
	}

	res, err := s.rwdbOperations.DeleteExtraPointsTournament(ctx, logger, req.Id, tx)
	if err != nil {
		tx.Rollback(ctx)
		return false, err
	}

	if err = tx.Commit(ctx); err != nil {
		tx.Rollback(ctx)
		return false, err
	}

	return res, nil
}

func (s *Service) GetExtraPointsById(ctx context.Context, extraPointsId int64) (entities.ExtraPointsTournament, error) {
	logger := s.logger.With().Str("service", "GetExtraPointsById").Logger()

	res, err := s.rdbOperations.GetExtraPointsTournamentById(ctx, logger, extraPointsId)
	if err != nil {
		return entities.ExtraPointsTournament{}, err
	}

	return res, nil
}

func (s *Service) GetExtraPointsListByTeamAndTournamentId(ctx context.Context, teamId, tournamentId int64) ([]entities.ExtraPointsTournament, error) {
	logger := s.logger.With().Str("service", "GetExtraPointsListByTeamAndTournamentId").Logger()

	res, err := s.rdbOperations.GetExtraPointsListByTeamAndTournamentId(ctx, logger, teamId, tournamentId)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *Service) MigrateLeaguesToTournamentsUp(ctx context.Context) error {
	logger := s.logger.With().Str("service", "MigrateLeaguesToTournamentsUp").Logger()

	migratedLeagueIds, err := s.rdbOperations.GetMigratedLeagueIds(ctx, logger)
	if err != nil {
		return err
	}

	leagues, err := s.rdbOperations.GetLeagueListToMigrate(ctx, logger)
	if err != nil {
		return err
	}

	tx, err := s.rwdbOperations.BeginTx(ctx, logger)
	if err != nil {
		return err
	}

	for _, leagueToMigrate := range leagues {
		if slices.Contains(migratedLeagueIds, leagueToMigrate.ID) {
			logger.Info().Msgf("лига id=%d уже мигрировала, пропускаем", leagueToMigrate.ID)
			continue
		}

		createTournamentReq := &entities.CreateTournamentRequest{}

		leagueTeamList, err := s.rdbOperations.GetTeamsByLeague(logger, ctx, leagueToMigrate.ID)
		if err != nil {
			tx.Rollback(ctx)
			return err
		}

		createTournamentReq.CityID = &leagueToMigrate.CityID
		createTournamentReq.Name = leagueToMigrate.Name
		createTournamentReq.Rules = entities.TournamentRule{Regular: &entities.Regular{BestOf: 2}}
		createTournamentReq.TournamentTypeID = constant.RegularTournamentTypeID
		createTournamentReq.Creator.Role = &entities.Role{Name: constant.SuperUserRole}

		var teamIds []int64
		for _, leagueTeam := range leagueTeamList {
			teamIds = append(teamIds, leagueTeam.Id)
		}

		createTournamentReq.TeamsIDs = teamIds

		if leagueToMigrate.SeasonID != nil {
			createTournamentReq.SeasonID = *leagueToMigrate.SeasonID
		}

		createdTournamentId, err := s.Create(ctx, createTournamentReq)
		if err != nil {
			tx.Rollback(ctx)
			return err
		}

		err = s.rwdbOperations.MarkLeagueAsMigrated(ctx, logger, leagueToMigrate.ID, createdTournamentId, tx)
		if err != nil {
			tx.Rollback(ctx)
			return err
		}

		leagueGames, err := s.rdbOperations.GetLeagueGamesToMigrate(ctx, logger, leagueToMigrate.ID)
		if err != nil {
			tx.Rollback(ctx)
			return err
		}

		for _, leagueGame := range leagueGames {
			if leagueGame.IsTiebreak {
				tournament, err := s.rdbOperations.GetTournamentById(ctx, logger, createdTournamentId)
				if err != nil {
					tx.Rollback(ctx)
					return err
				}

				var gameDate time.Time
				if leagueGame.Date != nil {
					gameDate = *leagueGame.Date
				}

				var placeId int64
				if leagueGame.PlaceID != nil {
					placeId = *leagueGame.PlaceID
				}

				_, err = s.rwdbOperations.CreatePlayedTournamentGame(ctx, logger, &entities.CreatePlayedTournamentGameRequest{
					TournamentID:    createdTournamentId,
					StageID:         tournament.Stages[0].ID, // у турнира по лиге только 1 стейдж, поэтому индекс 0
					PlaceID:         placeId,
					Date:            gameDate,
					Team1ID:         leagueGame.Team1ID,
					Team2ID:         leagueGame.Team2ID,
					IsTiebreak:      leagueGame.IsTiebreak,
					TechLooseTeamID: leagueGame.TechLooseTeamID,
					CityID:          leagueGame.CityID,
					Creator:         entities.User{Role: &entities.Role{Name: constant.SuperUserRole}},
				}, tx)
				if err != nil {
					tx.Rollback(ctx)
					return err
				}

				continue
			}

			id, err := s.rwdbOperations.UpdateMigratedTournamentGame(ctx, logger, leagueGame, createdTournamentId, tx)
			if err != nil {
				tx.Rollback(ctx)
				return err
			}

			gameMatches, err := s.rdbOperations.GetMatchListByGameID(ctx, logger, leagueGame.ID)
			if err != nil {
				tx.Rollback(ctx)
				return err
			}

			for _, match := range gameMatches {
				var player2Team1ID, player2Team2ID *int64
				var player1Team1RateBefore, player2Team1RateBefore, player1Team2RateBefore, player2Team2RateBefore *int64
				var player1Team1RateAfter, player2Team1RateAfter, player1Team2RateAfter, player2Team2RateAfter *int64

				if match.Player2Team1ID != nil {
					player2Team1ID = pointer.GetPointer(int64(*match.Player2Team1ID))
				}
				if match.Player2Team2ID != nil {
					player2Team2ID = pointer.GetPointer(int64(*match.Player2Team2ID))
				}
				if match.Player1Team1RateBefore != nil {
					player1Team1RateBefore = pointer.GetPointer(int64(*match.Player1Team1RateBefore))
				}
				if match.Player2Team1RateBefore != nil {
					player2Team1RateBefore = pointer.GetPointer(int64(*match.Player2Team1RateBefore))
				}
				if match.Player1Team2RateBefore != nil {
					player1Team2RateBefore = pointer.GetPointer(int64(*match.Player1Team2RateBefore))
				}
				if match.Player2Team2RateBefore != nil {
					player2Team2RateBefore = pointer.GetPointer(int64(*match.Player2Team2RateBefore))
				}
				if match.Player1Team1RateAfter != nil {
					player1Team1RateAfter = pointer.GetPointer(int64(*match.Player1Team1RateAfter))
				}
				if match.Player2Team1RateAfter != nil {
					player2Team1RateAfter = pointer.GetPointer(int64(*match.Player2Team1RateAfter))
				}
				if match.Player1Team2RateAfter != nil {
					player1Team2RateAfter = pointer.GetPointer(int64(*match.Player1Team2RateAfter))
				}
				if match.Player2Team2RateAfter != nil {
					player2Team2RateAfter = pointer.GetPointer(int64(*match.Player2Team2RateAfter))
				}

				_, err = s.rwdbOperations.CreateMatch(logger, ctx, entities.Match{
					Date:                   match.Date,
					GameID:                 &id,
					Team1ID:                pointer.GetPointer(int64(match.Team1ID)),
					Team2ID:                pointer.GetPointer(int64(match.Team2ID)),
					Player1Team1ID:         pointer.GetPointer(int64(match.Player1Team1ID)),
					Player2Team1ID:         player2Team1ID,
					Player1Team2ID:         pointer.GetPointer(int64(match.Player1Team2ID)),
					Player2Team2ID:         player2Team2ID,
					ScoreTeam1:             pointer.GetPointer(int64(match.ScoreTeam1)),
					ScoreTeam2:             pointer.GetPointer(int64(match.ScoreTeam2)),
					Player1Team1RateBefore: player1Team1RateBefore,
					Player2Team1RateBefore: player2Team1RateBefore,
					Player1Team2RateBefore: player1Team2RateBefore,
					Player2Team2RateBefore: player2Team2RateBefore,
					Player1Team1RateAfter:  player1Team1RateAfter,
					Player2Team1RateAfter:  player2Team1RateAfter,
					Player1Team2RateAfter:  player1Team2RateAfter,
					Player2Team2RateAfter:  player2Team2RateAfter,
					UpdatedAt:              match.UpdatedAt,
					Sort:                   match.Sort,
				})
				if err != nil {
					tx.Rollback(ctx)
					return err
				}
			}
		}

		ratings, err := s.rdbOperations.GetRatingsByLeagueIdToMigrate(ctx, logger, leagueToMigrate.ID)
		if err != nil {
			tx.Rollback(ctx)
			return err
		}

		for _, rating := range ratings {
			err = s.rwdbOperations.CreateTournamentRating(ctx, logger, rating.PlayerID, rating.Value, createdTournamentId, tx)
			if err != nil {
				tx.Rollback(ctx)
				return err
			}
		}

		extraPoints, err := s.rdbOperations.GetExtraPointsByLeagueIdToMigrate(ctx, logger, leagueToMigrate.ID)
		if err != nil {
			tx.Rollback(ctx)
			return err
		}

		for _, extraPoint := range extraPoints {
			_, err = s.rwdbOperations.CreateExtraPointsTournament(ctx, logger, &entities.CreateExtraPointsTournamentRequest{
				TeamId:       extraPoint.TeamId,
				TournamentId: createdTournamentId,
				Reason:       extraPoint.Reason,
				Points:       extraPoint.Points,
			}, tx)
			if err != nil {
				tx.Rollback(ctx)
				return err
			}
		}
	}

	if err = tx.Commit(ctx); err != nil {
		tx.Rollback(ctx)
		return err
	}

	return nil
}

func (s *Service) MigrateLeaguesToTournamentsDown(ctx context.Context) error {
	logger := s.logger.With().Str("service", "MigrateLeaguesToTournamentsDown").Logger()

	migratedTournamentIds, err := s.rdbOperations.GetMigratedTournamentIds(ctx, logger)
	if err != nil {
		return err
	}

	tx, err := s.rwdbOperations.BeginTx(ctx, logger)
	if err != nil {
		return err
	}

	for _, tournamentId := range migratedTournamentIds {
		err = s.rwdbOperations.UnmarkMigratedLeague(ctx, logger, tournamentId, tx)
		if err != nil {
			tx.Rollback(ctx)
			return err
		}

		err = s.Delete(ctx, &entities.DeleteTournamentRequest{
			ID:       tournamentId,
			Executor: entities.User{Role: &entities.Role{Name: constant.SuperUserRole}},
		})
		if err != nil {
			tx.Rollback(ctx)
			return err
		}
	}

	if err = tx.Commit(ctx); err != nil {
		tx.Rollback(ctx)
		return err
	}

	return nil
}

/*local methods*/

func (s *Service) createRegular(ctx context.Context, logger zerolog.Logger, request entities.CreateTournamentRequest) (int64, error) {
	tx, err := s.rwdbOperations.BeginTx(ctx, logger)
	if err != nil {
		return 0, err
	}

	id, err := s.rwdbOperations.CreateTournament(logger, ctx, request, tx)
	if err != nil {
		tx.Rollback(ctx)
		return 0, err
	}

	stageId, err := s.rwdbOperations.CreateTournamentStage(logger, ctx, id, oneStage, tx)
	if err != nil {
		tx.Rollback(ctx)
		return 0, err
	}

	team1IDs, team2IDs := helpers.GeneratePairs(request.TeamsIDs, int(request.Rules.Regular.BestOf))

	err = s.rwdbOperations.CreateFutureTournamentStageGames(ctx, logger, stageId, *request.CityID, team1IDs, team2IDs, tx)
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

func (s *Service) createRegularOneVsOne(ctx context.Context, logger zerolog.Logger, request entities.CreateTournamentRequest) (int64, error) {
	if len(request.PlayersIDs) == 0 {
		err := errors.New("игроки обязательны")
		return 0, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	teamIds := make([]int64, 0, len(request.TeamsIDs))

	players := make([]entities.Player, 0, len(request.PlayersIDs))

	tx, err := s.rwdbOperations.BeginTx(ctx, logger)
	if err != nil {
		return 0, err
	}

	for _, pId := range request.PlayersIDs {

		player, err := s.rdbOperations.GetPlayerByID(logger, ctx, int(pId), tx)
		if err != nil {
			tx.Rollback(ctx)
			return 0, err
		}

		if int64(pointer.GetValue(player.CityID)) != pointer.GetValue(request.CityID) {
			err = fmt.Errorf("id города игрока %s(%d) не совпадает с id города турнира(%d)", pointer.GetValue(player.Name), pointer.GetValue(player.CityID), pointer.GetValue(request.CityID))
			tx.Rollback(ctx)
			return 0, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}

		players = append(players, player)
	}

	for _, p := range players {
		teamAlreadyExist, existingTeam := false, entities.TeamItem{}

		// получаем все команды игрока, если команд нет, ошибки быть не должно
		teams, err := s.rdbOperations.GetTeamsByPlayerID(logger, ctx, p.ID, tx)
		if err != nil {
			tx.Rollback(ctx)
			return 0, err
		}

		// если среди всех команд есть с таким же названием как у игрока,
		// то проверяем сколько игроков в ней, если 1 - то считаем что она подходит и используем ее не создавая новую
		for _, team := range teams {
			name, _ := helpers.BuildTeamName(p.Name, p.SecondName, p.LastName, int64(p.ID))

			if team.Name == name {
				playersLoc, err := s.rdbOperations.GetPlayersByTeamID(logger, ctx, team.ID, tx)
				if err != nil {
					tx.Rollback(ctx)
					return 0, err
				}
				if len(playersLoc) == 1 {
					teamAlreadyExist = true
					existingTeam = team
					break
				}
			}
		}

		if teamAlreadyExist == true {
			// добавим существующую команду
			teamIds = append(teamIds, int64(existingTeam.ID))
		} else {
			// если такой команды не нашлось, создаем новую по имени
			name, shortName := helpers.BuildTeamName(p.Name, p.SecondName, p.LastName, int64(p.ID))

			teamId, err := s.rwdbOperations.CreateTeam(logger, ctx, entities.CreateTeamRequest{
				Name:      name,
				ShortName: shortName,
				CityId:    *request.CityID,
			}, tx)
			if err != nil {
				tx.Rollback(ctx)
				return 0, err
			}

			_, err = s.rwdbOperations.AddPlayerIntoTeam(logger, ctx, int64(p.ID), teamId, tx)
			if err != nil {
				tx.Rollback(ctx)
				return 0, err
			}

			teamIds = append(teamIds, teamId)
		}
	}

	request.TeamsIDs = teamIds

	id, err := s.rwdbOperations.CreateTournament(logger, ctx, request, tx)
	if err != nil {
		tx.Rollback(ctx)
		return 0, err
	}

	stageId, err := s.rwdbOperations.CreateTournamentStage(logger, ctx, id, oneStage, tx)
	if err != nil {
		tx.Rollback(ctx)
		return 0, err
	}

	team1IDs, team2IDs := helpers.GeneratePairs(request.TeamsIDs, int(request.Rules.Regular.BestOf))

	err = s.rwdbOperations.CreateFutureTournamentStageGames(ctx, logger, stageId, *request.CityID, team1IDs, team2IDs, tx)
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

func (s *Service) createPlayoff(ctx context.Context, logger zerolog.Logger, request entities.CreateTournamentRequest) (int64, error) {
	err := helpers.ValidatePlayoffTournamentRulesOnCreate(&request)
	if err != nil {
		logger.Error().Err(err).Msg("helpers.ValidatePlayoffTournamentRulesOnCreate")
		return 0, err
	}

	stageNums, stageBestOfList, err := helpers.SortKeysValues(request.Rules.PlayOff.Stages)
	if err != nil {
		logger.Error().Err(err).Msg("failed helpers.SortKeysValues")
		err = errors.New("нарушена нумерация этапов")
		return 0, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	tx, err := s.rwdbOperations.BeginTx(ctx, logger)
	if err != nil {
		return 0, err
	}

	tournamentId, err := s.rwdbOperations.CreateTournament(logger, ctx, request, tx)
	if err != nil {
		tx.Rollback(ctx)
		return 0, err
	}

	// id первого этапа, чтобы потом игры в него сохранить
	var firstStageId int64

	for i, stageNum := range stageNums {

		// создаем строку вида <порядковый_номер_этапа>/<кол-во_этапов>, образец: 1/4, 2/4, 3/4, 4/4 - финал
		stageSlashNumberOfStages := strconv.FormatInt(stageNum, 10) + "/" + strconv.Itoa(len(request.Rules.PlayOff.Stages))

		// создаем этап
		stageId, err := s.rwdbOperations.CreateTournamentStage(logger, ctx, tournamentId, stageSlashNumberOfStages, tx)
		if err != nil {
			tx.Rollback(ctx)
			return 0, err
		}

		// сохраним id первого этапа, чтобы потом туда положить готовые игры
		if i == indexZero {
			firstStageId = stageId
		}
	}

	// расставляем пары для первого этапа
	// BestOf берётся из первого элемента мапы реквеста - stages
	teams1Ids, teams2Ids := helpers.GeneratePlayoffPairs(request.TeamsIDs, int(stageBestOfList[indexZero].Bo))
	// создаем игры первого этапа
	if err = s.rwdbOperations.CreateFutureTournamentStageGames(ctx, logger, firstStageId, *request.CityID, teams1Ids, teams2Ids, tx); err != nil {
		tx.Rollback(ctx)
		return 0, err
	}

	if err = tx.Commit(ctx); err != nil {
		tx.Rollback(ctx)
		return 0, err
	}

	return tournamentId, nil
}

func (s *Service) createRegularPlayoff(ctx context.Context, logger zerolog.Logger, req entities.CreateTournamentRequest) (int64, error) {
	tx, err := s.rwdbOperations.BeginTx(ctx, logger)
	if err != nil {
		return 0, err
	}

	err = helpers.ValidateRegularPlayoffTournamentRulesOnCreate(&req)
	if err != nil {
		logger.Error().Err(err).Msg("failed helpers.ValidateRegularPlayoffTournamentRulesOnCreate")
		return 0, err
	}

	// здесь важно просто создать этапы для будущих игр, но на данный момент мы не знаем, кто будет играть в них,
	// т.к. ещё нет результатов regular-этапа
	playoffStageNums, _, err := helpers.SortKeysValues(req.Rules.RegularPlayoff.PlayOff.Stages)
	if err != nil {
		logger.Error().Err(err).Msg("failed helpers.SortKeysValues")
		err = errors.New("нарушена нумерация этапов")
		return 0, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	tournamentId, err := s.rwdbOperations.CreateTournament(logger, ctx, req, tx)
	if err != nil {
		tx.Rollback(ctx)
		return 0, err
	}

	// создаем regular-этап 1/N
	regularStage := strconv.FormatInt(constant.FirstStage, 10) + constant.SeparatorStageNumber + strconv.Itoa(constant.FirstStage+len(req.Rules.RegularPlayoff.PlayOff.Stages))

	firstStageId, err := s.rwdbOperations.CreateTournamentStage(logger, ctx, tournamentId, regularStage, tx)
	if err != nil {
		tx.Rollback(ctx)
		return 0, err
	}

	team1IDs, team2IDs := helpers.GeneratePairs(req.TeamsIDs, int(req.Rules.RegularPlayoff.Regular.BestOf))

	// игры regular-этапа
	err = s.rwdbOperations.CreateFutureTournamentStageGames(ctx, logger, firstStageId, *req.CityID, team1IDs, team2IDs, tx)
	if err != nil {
		tx.Rollback(ctx)
		return 0, err
	}

	for i := 0; i < len(playoffStageNums); i++ {
		// создаём строку вида <порядковый_номер_этапа>/<кол-во_этапов>, образец: 2/4, 3/4, 4/4 - финал.
		// 1/N пропускаем, так как он уже создан в рамках regular
		stageSlashNumberOfStages := strconv.FormatInt(int64(i+2), 10) + constant.SeparatorStageNumber + strconv.Itoa(constant.FirstStage+len(req.Rules.RegularPlayoff.PlayOff.Stages))

		_, err = s.rwdbOperations.CreateTournamentStage(logger, ctx, tournamentId, stageSlashNumberOfStages, tx)
		if err != nil {
			tx.Rollback(ctx)
			return 0, err
		}
	}

	if err = tx.Commit(ctx); err != nil {
		tx.Rollback(ctx)
		return 0, err
	}

	return tournamentId, nil
}

func (s *Service) updateRegular(ctx context.Context, logger zerolog.Logger, req entities.UpdateTournamentRequest, tournament entities.Tournament) error {
	if len(req.PlayersIDs) > 0 {
		err := errors.New("для обновления тура Regular playersIDs не должны быть заполнены")
		return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	newReq := entities.UpdateTournamentRequest{
		ID:               req.ID,
		Executor:         req.Executor,
		TournamentTypeID: req.TournamentTypeID,
		PlayersIDs:       nil,
	}

	if req.CityID != nil {
		newReq.CityID = req.CityID
	} else {
		newReq.CityID = &tournament.CityID
	}

	if req.SeasonID != nil {
		newReq.SeasonID = req.SeasonID
	} else {
		newReq.SeasonID = &tournament.SeasonID
	}

	if req.Name != nil {
		newReq.Name = req.Name
	} else {
		newReq.Name = &tournament.Name
	}

	if req.Rules != nil {
		newReq.Rules = req.Rules
	} else {
		newReq.Rules = &tournament.Rules
	}

	if len(req.TeamsIDs) > 0 {
		err := s.checkTeamsCityIds(ctx, logger, req.TeamsIDs, tournament, newReq.CityID)
		if err != nil {
			return err
		}

		// если ошибок нет, то ставим команды на место
		newReq.TeamsIDs = req.TeamsIDs

	} else {
		newReq.TeamsIDs = tournament.TeamIDs
	}

	err := validateRulesTypesIds(*newReq.Rules, newReq.TournamentTypeID, newReq.TeamsIDs, nil)
	if err != nil {
		return err
	}

	tx, err := s.rwdbOperations.BeginTx(ctx, logger)
	if err != nil {
		return err
	}

	err = s.rwdbOperations.DeleteTournamentGamesTeamLinks(logger, ctx, newReq.ID, tx)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	err = s.rwdbOperations.UpdateTournament(logger, ctx, newReq, tx)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	team1IDs, team2IDs := helpers.GeneratePairs(newReq.TeamsIDs, int(newReq.Rules.Regular.BestOf))

	err = s.rwdbOperations.CreateFutureTournamentStageGames(ctx, logger, tournament.Stages[indexZero].ID, *newReq.CityID, team1IDs, team2IDs, tx)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	err = s.rwdbOperations.UpdateTournamentTeamsLinks(logger, ctx, newReq.TeamsIDs, newReq.ID, tx)
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

func (s *Service) updateRegularOneVsOne(ctx context.Context, logger zerolog.Logger, req entities.UpdateTournamentRequest, tournament entities.Tournament) error {
	if len(req.PlayersIDs) > 0 && len(req.TeamsIDs) > 0 {
		err := errors.New("для обновления тура Regular.OneVsOne в запросе должны быть либо только teamsIds либо playersIds либо ничего")
		return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	newReq := entities.UpdateTournamentRequest{
		ID:               req.ID,
		Executor:         req.Executor,
		TournamentTypeID: req.TournamentTypeID,
	}

	if req.CityID != nil {
		newReq.CityID = req.CityID
	} else {
		newReq.CityID = &tournament.CityID
	}

	if req.SeasonID != nil {
		newReq.SeasonID = req.SeasonID
	} else {
		newReq.SeasonID = &tournament.SeasonID
	}

	if req.Name != nil {
		newReq.Name = req.Name
	} else {
		newReq.Name = &tournament.Name
	}

	if req.Rules != nil {
		newReq.Rules = req.Rules
	} else {
		newReq.Rules = &tournament.Rules
	}

	if len(req.TeamsIDs) > 0 {
		err := s.checkTeamsCityIds(ctx, logger, req.TeamsIDs, tournament, newReq.CityID)
		if err != nil {
			return err
		}

		newReq.TeamsIDs = req.TeamsIDs
	} else {
		if len(req.PlayersIDs) == 0 {
			newReq.TeamsIDs = tournament.TeamIDs
		}
	}

	if len(req.PlayersIDs) > 0 {
		err := s.checkPlayersCityIds(ctx, logger, req.PlayersIDs, newReq.CityID)
		if err != nil {
			return err
		}

		newReq.PlayersIDs = req.PlayersIDs
	}

	if err := validateRulesTypesIds(*newReq.Rules, newReq.TournamentTypeID, newReq.TeamsIDs, newReq.PlayersIDs); err != nil {
		return err
	}

	tx, err := s.rwdbOperations.BeginTx(ctx, logger)
	if err != nil {
		return err
	}

	if err = s.rwdbOperations.UpdateTournament(logger, ctx, newReq, tx); err != nil {
		tx.Rollback(ctx)
		return err
	}

	// если пришли новые игроки,то делаем из них новые команды, перезаписываем всё что с ними связано и выходим
	if len(req.PlayersIDs) > 0 {
		teamIds := make([]int64, 0, len(newReq.TeamsIDs))
		players := make([]entities.Player, 0, len(newReq.PlayersIDs))

		for _, pId := range req.PlayersIDs {

			player, err := s.rdbOperations.GetPlayerByID(logger, ctx, int(pId), tx)
			if err != nil {
				tx.Rollback(ctx)
				return err
			}

			players = append(players, player)
		}

		for _, p := range players {
			teamAlreadyExist, existingTeam := false, entities.TeamItem{}

			// получаем все команды игрока, если они есть
			teams, err := s.rdbOperations.GetTeamsByPlayerID(logger, ctx, p.ID, tx)
			if err != nil {
				tx.Rollback(ctx)
				return err
			}

			// если среди всех команд есть с таким же названием как у игрока,
			// то проверяем сколько игроков в ней, если 1 - то считаем что она подходит и используем ее не создавая новую
			for _, team := range teams {
				name, _ := helpers.BuildTeamName(p.Name, p.SecondName, p.LastName, int64(p.ID))

				if strings.Contains(team.Name, name) {
					playersLoc, err := s.rdbOperations.GetPlayersByTeamID(logger, ctx, team.ID, tx)
					if err != nil {
						tx.Rollback(ctx)
						return err
					}
					if len(playersLoc) == 1 {
						teamAlreadyExist = true
						existingTeam = team
						break
					}
				}
			}

			// добавим существующую команду
			if teamAlreadyExist == true {

				teamIds = append(teamIds, int64(existingTeam.ID))

				// если такой команды не нашлось, создаем новую по имени
			} else {

				name, shortName := helpers.BuildTeamName(p.Name, p.SecondName, p.LastName, int64(p.ID))

				teamId, err := s.rwdbOperations.CreateTeam(logger, ctx, entities.CreateTeamRequest{
					Name:      name,
					ShortName: shortName,
					CityId:    *newReq.CityID,
				}, tx)
				if err != nil {
					tx.Rollback(ctx)
					return err
				}

				_, err = s.rwdbOperations.AddPlayerIntoTeam(logger, ctx, int64(p.ID), teamId, tx)
				if err != nil {
					tx.Rollback(ctx)
					return err
				}

				teamIds = append(teamIds, teamId)
			}
		}

		// удаляем связи всех старых команд с турниром и удаляем игры турнира
		if err = s.rwdbOperations.DeleteTournamentGamesTeamLinks(logger, ctx, newReq.ID, tx); err != nil {
			tx.Rollback(ctx)
			return err
		}

		// делаем новые пары
		team1IDs, team2IDs := helpers.GeneratePairs(teamIds, int(newReq.Rules.Regular.BestOf))

		// создаем новые игры
		err = s.rwdbOperations.CreateFutureTournamentStageGames(ctx, logger, tournament.Stages[indexZero].ID, *newReq.CityID, team1IDs, team2IDs, tx)
		if err != nil {
			tx.Rollback(ctx)
			return err
		}

		// обновляем связи турнира и новых команд
		if err = s.rwdbOperations.UpdateTournamentTeamsLinks(logger, ctx, teamIds, newReq.ID, tx); err != nil {
			tx.Rollback(ctx)
			return err
		}

		if err = tx.Commit(ctx); err != nil {
			tx.Rollback(ctx)
			return err
		}

		return nil
	}

	// если в запросе пришли команды то тоже перезаписываем все данные и выходим
	if len(req.TeamsIDs) > 0 {
		teamIds := make([]int64, 0, len(newReq.TeamsIDs))

		for _, tId := range newReq.TeamsIDs {
			team, err := s.rdbOperations.GetTeamById(logger, ctx, tId, tx)
			if err != nil {
				tx.Rollback(ctx)
				return err
			}

			players, err := s.rdbOperations.GetPlayersByTeamID(logger, ctx, int(tId), tx)
			if err != nil {
				tx.Rollback(ctx)
				return err
			}
			if len(players) > 1 {
				err = errors.New("в команде более одного игрока")
				logger.Error().Err(err).Msg("more than 1 player")
				tx.Rollback(ctx)
				return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
			}

			teamIds = append(teamIds, team.Id)
		}

		// удаляем связи всех старых команд с турниром и удаляем игры турнира
		if err = s.rwdbOperations.DeleteTournamentGamesTeamLinks(logger, ctx, newReq.ID, tx); err != nil {
			tx.Rollback(ctx)
			return err
		}

		// делаем новые пары
		team1IDs, team2IDs := helpers.GeneratePairs(teamIds, int(newReq.Rules.Regular.BestOf))

		// создаем новые игры
		err = s.rwdbOperations.CreateFutureTournamentStageGames(ctx, logger, tournament.Stages[indexZero].ID, *newReq.CityID, team1IDs, team2IDs, tx)
		if err != nil {
			tx.Rollback(ctx)
			return err
		}

		// обновляем связи турнира и новых команд
		if err = s.rwdbOperations.UpdateTournamentTeamsLinks(logger, ctx, teamIds, newReq.ID, tx); err != nil {
			tx.Rollback(ctx)
			return err
		}

		if err = tx.Commit(ctx); err != nil {
			tx.Rollback(ctx)
			return err
		}

		return nil
	}

	// если команды не поменялись, но поменялись правила формирования турнира или город, то перезаписываем всё и выходим
	if len(req.TeamsIDs) == 0 && (req.Rules != nil || req.CityID != nil) {

		// удаляем связь турнира с командами и игры турнира
		if err = s.rwdbOperations.DeleteTournamentGamesTeamLinks(logger, ctx, newReq.ID, tx); err != nil {
			tx.Rollback(ctx)
			return err
		}

		// делаем новые пары
		team1IDs, team2IDs := helpers.GeneratePairs(newReq.TeamsIDs, int(newReq.Rules.Regular.BestOf))

		// создаем новые игры
		if err = s.rwdbOperations.CreateFutureTournamentStageGames(ctx, logger, tournament.Stages[indexZero].ID, *newReq.CityID, team1IDs, team2IDs, tx); err != nil {
			tx.Rollback(ctx)
			return err
		}

		// обновляем связи турнира и новых команд
		if err = s.rwdbOperations.UpdateTournamentTeamsLinks(logger, ctx, newReq.TeamsIDs, newReq.ID, tx); err != nil {
			tx.Rollback(ctx)
			return err
		}

		if err = tx.Commit(ctx); err != nil {
			tx.Rollback(ctx)
			return err
		}

		return nil
	}

	if err = tx.Commit(ctx); err != nil {
		tx.Rollback(ctx)
		return err
	}

	// команды не менялись, правила турнира и город тоже
	return nil
}

func (s *Service) updatePlayoff(ctx context.Context, logger zerolog.Logger, req entities.UpdateTournamentRequest, tournament entities.Tournament) error {
	// вызываем первый раз для валидации только входящих данных, второй параметр поэтому равен nil
	err := helpers.ValidatePlayoffTournamentRulesOnUpdate(req, nil)
	if err != nil {
		logger.Error().Err(err).Msg("helpers.ValidatePlayoffTournamentRulesOnUpdate")
		return err
	}

	// собираем запрос с учетом старых данных турнира
	newReq := entities.UpdateTournamentRequest{
		ID:               req.ID,
		Executor:         req.Executor,
		TournamentTypeID: req.TournamentTypeID,
	}
	if req.CityID != nil {
		newReq.CityID = req.CityID
	} else {
		newReq.CityID = &tournament.CityID
	}
	if req.SeasonID != nil {
		newReq.SeasonID = req.SeasonID
	} else {
		newReq.SeasonID = &tournament.SeasonID
	}
	if req.Name != nil {
		newReq.Name = req.Name
	} else {
		newReq.Name = &tournament.Name
	}
	if req.Rules != nil {
		newReq.Rules = req.Rules
	} else {
		newReq.Rules = &tournament.Rules
	}

	// если в запросе были новые команды, то заново их проверить на соответствие городу
	if len(req.TeamsIDs) > 0 {
		err := s.checkTeamsCityIds(ctx, logger, req.TeamsIDs, tournament, newReq.CityID)
		if err != nil {
			return err
		}
		newReq.TeamsIDs = req.TeamsIDs
	} else {
		newReq.TeamsIDs = tournament.TeamIDs
	}

	// вызываем второй раз для валидации собранных данных турнира
	err = helpers.ValidatePlayoffTournamentRulesOnUpdate(req, &newReq)
	if err != nil {
		logger.Error().Err(err).Msg("helpers.ValidatePlayoffTournamentRulesOnUpdate")
		return err
	}

	tx, err := s.rwdbOperations.BeginTx(ctx, logger)
	if err != nil {
		return err
	}

	err = s.rwdbOperations.UpdateTournament(logger, ctx, newReq, tx)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	// если пришли новые команды, изменились правила или город - всё перезаписываем
	if len(req.TeamsIDs) > 0 || req.Rules != nil || req.CityID != nil {
		// старые команды отвязать от турнира
		if err = s.rwdbOperations.UnlinkTeamsFromTournament(logger, ctx, tournament.ID, tx); err != nil {
			tx.Rollback(ctx)
			return err
		}

		// удалить игры старых команд
		if err = s.rwdbOperations.DeleteTournamentGames(logger, ctx, tournament.ID, tx); err != nil {
			tx.Rollback(ctx)
			return err
		}

		// удалить этапы турнира
		if err = s.rwdbOperations.DeleteTournamentStages(logger, ctx, tournament.ID, tx); err != nil {
			tx.Rollback(ctx)
			return err
		}

		// добавим новые команды в турнир
		if err = s.rwdbOperations.AddTeamsToTournament(ctx, logger, newReq.TeamsIDs, tournament.ID, tx); err != nil {
			tx.Rollback(ctx)
			return err
		}

		// создаем номера этапов и параметр BestOf каждому из них
		stageNums, stageBestOfList, err := helpers.SortKeysValues(newReq.Rules.PlayOff.Stages)
		if err != nil {
			logger.Error().Err(err).Msg("helpers.SortKeysValues")
			err = errors.New("нарушена нумерация этапов")
			return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}

		// id первого этапа, чтобы потом игры в него сохранить
		var firstStageId int64

		for i, stageNum := range stageNums {

			// создаем строку вида <порядковый_номер_этапа>/<кол-во_этапов>, образец: 1/4, 2/4, 3/4, 4/4 - финал
			stageSlashNumberOfStages := strconv.FormatInt(stageNum, 10) + "/" + strconv.Itoa(len(newReq.Rules.PlayOff.Stages))

			// создаем этап
			stageId, err := s.rwdbOperations.CreateTournamentStage(logger, ctx, tournament.ID, stageSlashNumberOfStages, tx)
			if err != nil {
				tx.Rollback(ctx)
				return err
			}

			// сохраним id первого этапа, чтобы потом туда положить готовые игры
			if i == indexZero {
				firstStageId = stageId
			}
		}

		// расставляем пары для первого этапа
		// BestOf берётся из первого элемента мапы реквеста - stages
		teams1Ids, teams2Ids := helpers.GeneratePlayoffPairs(newReq.TeamsIDs, int(stageBestOfList[indexZero].Bo))
		// создаем игры первого этапа
		if err = s.rwdbOperations.CreateFutureTournamentStageGames(ctx, logger, firstStageId, *newReq.CityID, teams1Ids, teams2Ids, tx); err != nil {
			tx.Rollback(ctx)
			return err
		}
	}

	if err = tx.Commit(ctx); err != nil {
		logger.Error().Err(err).Msg("failed tx.Commit")
		tx.Rollback(ctx)
		return err
	}

	return nil
}

func (s *Service) updateRegularPlayoff(ctx context.Context, logger zerolog.Logger, req entities.UpdateTournamentRequest, tournament entities.Tournament) error {
	// вызываем первый раз для валидации только входящих данных, второй параметр поэтому равен nil
	err := helpers.ValidateRegularPlayoffTournamentRulesOnUpdate(req, nil)
	if err != nil {
		logger.Error().Err(err).Msg("failed to helpers.ValidateRegularPlayoffTournamentRulesOnUpdate")
		return err
	}

	// собираем запрос с учетом старых данных турнира
	newReq := entities.UpdateTournamentRequest{
		ID:               req.ID,
		Executor:         req.Executor,
		TournamentTypeID: req.TournamentTypeID,
	}
	if req.CityID != nil {
		newReq.CityID = req.CityID
	} else {
		newReq.CityID = &tournament.CityID
	}
	if req.SeasonID != nil {
		newReq.SeasonID = req.SeasonID
	} else {
		newReq.SeasonID = &tournament.SeasonID
	}
	if req.Name != nil {
		newReq.Name = req.Name
	} else {
		newReq.Name = &tournament.Name
	}
	if req.Rules != nil {
		newReq.Rules = req.Rules
	} else {
		newReq.Rules = &tournament.Rules
	}

	// если в запросе были новые команды, то заново их проверить на соответствие городу
	if len(req.TeamsIDs) > 0 {
		err := s.checkTeamsCityIds(ctx, logger, req.TeamsIDs, tournament, newReq.CityID)
		if err != nil {
			return err
		}
		newReq.TeamsIDs = req.TeamsIDs
	} else {
		newReq.TeamsIDs = tournament.TeamIDs
	}

	// вызываем второй раз для валидации собранных данных турнира
	err = helpers.ValidateRegularPlayoffTournamentRulesOnUpdate(req, &newReq)
	if err != nil {
		logger.Error().Err(err).Msg("failed to helpers.ValidateRegularPlayoffTournamentRulesOnUpdate")
		return err
	}

	tx, err := s.rwdbOperations.BeginTx(ctx, logger)
	if err != nil {
		return err
	}

	err = s.rwdbOperations.UpdateTournament(logger, ctx, newReq, tx)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	// если пришли новые команды, или если изменились правила или город - всё перезаписываем
	if len(req.TeamsIDs) > 0 || req.Rules != nil || req.CityID != nil {
		// старые команды отвязать от турнира
		if err = s.rwdbOperations.UnlinkTeamsFromTournament(logger, ctx, tournament.ID, tx); err != nil {
			tx.Rollback(ctx)
			return err
		}

		// удалить игры старых команд
		if err = s.rwdbOperations.DeleteTournamentGames(logger, ctx, tournament.ID, tx); err != nil {
			tx.Rollback(ctx)
			return err
		}

		// удалить этапы турнира
		if err = s.rwdbOperations.DeleteTournamentStages(logger, ctx, tournament.ID, tx); err != nil {
			tx.Rollback(ctx)
			return err
		}

		// добавим новые команды в турнир
		if err = s.rwdbOperations.AddTeamsToTournament(ctx, logger, newReq.TeamsIDs, tournament.ID, tx); err != nil {
			tx.Rollback(ctx)
			return err
		}

		playoffStageNums, _, err := helpers.SortKeysValues(newReq.Rules.RegularPlayoff.PlayOff.Stages)
		if err != nil {
			logger.Error().Err(err).Msg("failed helpers.SortKeysValues")
			err = errors.New("нарушена нумерация этапов")
			return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}

		// создаем regular-этап 1/N
		regularStage := strconv.FormatInt(constant.FirstStage, 10) + constant.SeparatorStageNumber + strconv.Itoa(constant.FirstStage+len(newReq.Rules.RegularPlayoff.PlayOff.Stages))

		regularStageId, err := s.rwdbOperations.CreateTournamentStage(logger, ctx, newReq.ID, regularStage, tx)
		if err != nil {
			tx.Rollback(ctx)
			return err
		}

		team1IDs, team2IDs := helpers.GeneratePairs(newReq.TeamsIDs, int(newReq.Rules.RegularPlayoff.Regular.BestOf))

		// игры regular-этапа
		err = s.rwdbOperations.CreateFutureTournamentStageGames(ctx, logger, regularStageId, *newReq.CityID, team1IDs, team2IDs, tx)
		if err != nil {
			tx.Rollback(ctx)
			return err
		}

		for i := 0; i < len(playoffStageNums); i++ {
			// создаём строку вида <порядковый_номер_этапа>/<кол-во_этапов>, образец: 2/4, 3/4, 4/4 - финал.
			// 1/N пропускаем, так как он уже создан в рамках regular
			stageSlashNumberOfStages := strconv.FormatInt(int64(i+2), 10) + constant.SeparatorStageNumber + strconv.Itoa(constant.FirstStage+len(newReq.Rules.RegularPlayoff.PlayOff.Stages))

			_, err = s.rwdbOperations.CreateTournamentStage(logger, ctx, newReq.ID, stageSlashNumberOfStages, tx)
			if err != nil {
				tx.Rollback(ctx)
				return err
			}
		}

		if err = tx.Commit(ctx); err != nil {
			tx.Rollback(ctx)
			return err
		}
	}

	return nil
}

func (s *Service) checkTournamentMasterCredentials(ctx context.Context, logger zerolog.Logger, req *entities.UpdateTournamentRequest, tournament entities.Tournament, games []entities.TournamentGame) error {
	master, err := s.rdbOperations.GetTournamentMasterByUserId(logger, ctx, req.Executor.ID)
	if err != nil {
		return err
	}

	if tournament.CityID != master.City.ID {
		err = fmt.Errorf(
			"город турнира(ID: %d) не совпадает с городом мастера по турнирам(ID: %d)",
			tournament.CityID,
			master.City.ID,
		)
		logger.Error().Err(err).Msg("Failed tournament.Update by tournament master")
		return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	if req.CityID != nil && *req.CityID != master.City.ID {
		err = fmt.Errorf(
			"город в запросе(ID: %d) не совпадает с городом мастера по турнирам(ID: %d)",
			*req.CityID,
			master.City.ID,
		)
		logger.Error().Err(err).Msg("Failed tournament.Update by tournament master")
		return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	if len(req.TeamsIDs) > 0 {
		for _, tId := range req.TeamsIDs {
			team, err := s.rdbOperations.GetTeamById(logger, ctx, tId, nil)
			if err != nil {
				return err
			}

			if pointer.GetValue(team.CityId) != master.City.ID {
				err = fmt.Errorf("город(id: %d) команды не совпадает с городом(id: %d) мастера по турнирам", pointer.GetValue(team.CityId), master.City.ID)
				return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
			}
		}
	}

	if len(req.PlayersIDs) > 0 {
		for _, pId := range req.PlayersIDs {
			player, err := s.rdbOperations.GetPlayerByID(logger, ctx, int(pId), nil)
			if err != nil {
				return err
			}

			if int64(pointer.GetValue(player.CityID)) != master.City.ID {
				err = fmt.Errorf("город(id: %d) игрока не совпадает с городом(id: %d) мастера по турнирам", pointer.GetValue(player.CityID), master.City.ID)
				return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
			}
		}
	}

	if haveFinishedStage(tournament.Stages) == true {
		err = errors.New("турнир не может меняться мастером по турнирам если имеет завершенные этапы")
		logger.Error().Err(err).Msg("Failed tournament.Update by tournament master")
		return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	for _, game := range games {
		if game.Date != nil && time.Now().After(*game.Date) {
			err = fmt.Errorf(
				"турнир \"%s\"(ID: %d) не может меняться мастером по турнирам если имеет завершенные/текущие игры",
				tournament.Name,
				tournament.ID,
			)
			logger.Error().Err(err).Msg("Failed tournament.Update by tournament master")
			return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}

		matches, err := s.rdbOperations.GetMatchListByGameID(ctx, logger, game.ID)
		if err != nil {
			logger.Error().Err(err).Msg("Failed tournament.Update by tournament master")
			return err
		}
		if len(matches) > 0 {
			err = fmt.Errorf(
				"турнир \"%s\"(ID: %d) не может меняться мастером по турнирам если есть сыгранные матчи",
				tournament.Name,
				tournament.ID,
			)
			logger.Error().Err(err).Msg("Failed tournament.Update by tournament master")
			return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}

		if game.CityID != master.City.ID {
			err = fmt.Errorf(
				"город игры(ID: %d) не совпадает с городом мастера по турнирам(ID: %d)",
				game.CityID,
				master.City.ID,
			)
			logger.Error().Err(err).Msg("Failed tournament.Update by tournament master")
			return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}
	}

	return nil
}

func (s *Service) checkTournamentMasterCredentialsOnDelete(ctx context.Context, logger zerolog.Logger, req *entities.DeleteTournamentRequest, tournament entities.Tournament) error {
	master, err := s.rdbOperations.GetTournamentMasterByUserId(logger, ctx, req.Executor.ID)
	if err != nil {
		return err
	}

	games, err := s.rdbOperations.GetTournamentGameList(logger, ctx, &entities.GetTournamentGameList{TournamentId: &tournament.ID})
	if err != nil {
		return err
	}

	if tournament.CityID != master.City.ID {
		err = fmt.Errorf(
			"город турнира(ID: %d) не совпадает с городом мастера по турнирам(ID: %d)",
			tournament.CityID,
			master.City.ID,
		)
		logger.Error().Err(err).Msg("Failed GetTournamentGameList")
		return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	if haveFinishedStage(tournament.Stages) == true {
		err = errors.New("турнир не может меняться мастером по турнирам если имеет завершенные этапы")
		logger.Error().Err(err).Msg("have finished stages")
		return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	for _, game := range games {
		if game.Date != nil && time.Now().After(*game.Date) {
			err = fmt.Errorf(
				"турнир \"%s\"(ID: %d) не может удаляться мастером по турнирам если имеет завершенные/текущие игры",
				tournament.Name,
				tournament.ID,
			)
			logger.Error().Err(err).Msg("game is played")
			return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}

		matches, err := s.rdbOperations.GetMatchListByGameID(ctx, logger, game.ID)
		if err != nil {
			logger.Error().Err(err).Msg("Failed GetMatchListByGameID")
			return err
		}
		if len(matches) > 0 {
			err = fmt.Errorf(
				"турнир \"%s\"(ID: %d) не может удаляться мастером по турнирам если есть сыгранные матчи",
				tournament.Name,
				tournament.ID,
			)
			logger.Error().Err(err).Msg("have matches")
			return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}

		if game.CityID != master.City.ID {
			err = fmt.Errorf(
				"город игры(ID: %d) не совпадает с городом мастера по турнирам(ID: %d)",
				game.CityID,
				master.City.ID,
			)
			logger.Error().Err(err).Msg("Game city not equal master city")
			return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}
	}

	return nil
}

func (s *Service) checkTeamsCityIds(ctx context.Context, logger zerolog.Logger, newTeamsIds []int64, tournament entities.Tournament, newCityId *int64) error {
	needCheckCity := false // чтобы знать нужно ли проверять города у команд

	for _, reqTeamId := range newTeamsIds {
		thisTeamIsOld := false // отметка что команда есть в изменяемом турнире(по дефолту ее там нет)

		// если команда есть, то обозначаем ее как новую и выходим из внутреннего цикла
		for _, trnTeamId := range tournament.TeamIDs {
			if reqTeamId == trnTeamId {
				thisTeamIsOld = true
				break
			}
		}

		// проверяем отметку о том новая ли команда, если да, то у нее надо проверить город, отмечаем и выходим
		if thisTeamIsOld == false {
			needCheckCity = true
			break
		}
	}

	// если входящий город отличается от базового
	if pointer.GetValue(newCityId) != tournament.CityID {
		needCheckCity = true
	}

	// ничто не мешает админу при обновлении положить иногороднюю команду, поэтому проверяем
	if needCheckCity == true {
		for _, reqTeamId := range newTeamsIds {
			team, err := s.rdbOperations.GetTeamById(logger, ctx, reqTeamId, nil)
			if err != nil {
				return err
			}

			if pointer.GetValue(team.CityId) != pointer.GetValue(newCityId) {
				err = fmt.Errorf("id города команды %s(%d) не совпадает с id города турнира(%d)", team.Name, pointer.GetValue(team.CityId), pointer.GetValue(newCityId))
				return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
			}
		}
	}

	return nil
}

func (s *Service) checkPlayersCityIds(ctx context.Context, logger zerolog.Logger, newPlayersIds []int64, newCityId *int64) error {
	for _, pId := range newPlayersIds {
		player, err := s.rdbOperations.GetPlayerByID(logger, ctx, int(pId), nil)
		if err != nil {
			return err
		}

		if int64(pointer.GetValue(player.CityID)) != pointer.GetValue(newCityId) {
			err = fmt.Errorf("id города игрока %s(%d) не совпадает с id города турнира(%d)", pointer.GetValue(player.Name), pointer.GetValue(player.CityID), pointer.GetValue(newCityId))
			return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}
	}

	return nil
}

func (s *Service) checkMasterAndTournamentCities(ctx context.Context, tournamentId, userId int64, role string) error {
	logger := s.logger.With().Str("service", "checkMasterAndTournamentCities").Logger()

	if role == constant.TournamentMaster {
		masterInfo, err := s.rdbOperations.GetTournamentMasterByUserId(logger, ctx, userId)
		if err != nil {
			return err
		}

		tournamentInfo, err := s.rdbOperations.GetTournamentById(ctx, logger, tournamentId)
		if err != nil {
			return err
		}

		if masterInfo.City.ID != tournamentInfo.CityID {
			err = errors.New(pkgerr.ErrCityTournamentAndMasterMismatch)
			logger.Error().Err(err).Msg("master city not equal tournament city")
			return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}
	}

	return nil
}

func (s *Service) startNextStagePlayoff(ctx context.Context, logger zerolog.Logger, tournament entities.Tournament) (int64, error) {
	// все этапы турнира
	stages, err := s.rdbOperations.GetTournamentStageList(ctx, logger, tournament.ID)
	if err != nil {
		return 0, err
	}

	orderedStages, err := s.getOrderedStagesWithWinners(ctx, logger, tournament, stages)
	if err != nil {
		return 0, err
	}

	nextStage, previousStage, foundNextStage, err := checkOrderedStages(orderedStages)
	if err != nil {
		return 0, err
	}

	// если такой не нашелся - либо играется последний этап, либо непредвиденный косяк
	if *foundNextStage == false {
		err = errors.New("не найден следующий этап")
		logger.Error().Err(err).Msg("next stage not found")
		return 0, error_templates.New(err.Error(), err, codes.FailedPrecondition, http.StatusConflict)
	}

	// если в предыдущем этапе посчиталось неправильное кол-во победителей,
	// то это либо получилась ничья, либо непредвиденный косяк
	err = helpers.CheckQtyWinnersInPreviousPlayoffStage(len(tournament.TeamIDs), len(previousStage.Winners), int(nextStage.Number), len(stages))
	if err != nil {
		logger.Error().Err(err).Msg("helpers.CheckQtyWinners")
		return 0, err
	}

	// если предыдущий этап не завершён, не даем создать новый
	if previousStage.Stage.IsFinished == false {
		err = errors.New("предыдущий этап не завершён")
		return 0, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	teams1Ids, teams2Ids := helpers.GeneratePlayoffPairs(previousStage.Winners, int(nextStage.BestOf))

	tx, err := s.rwdbOperations.BeginTx(ctx, logger)
	if err != nil {
		return 0, err
	}

	if err = s.rwdbOperations.CreateFutureTournamentStageGames(ctx, logger, nextStage.Stage.ID, tournament.CityID, teams1Ids, teams2Ids, tx); err != nil {
		tx.Rollback(ctx)
		return 0, err
	}

	if err = tx.Commit(ctx); err != nil {
		tx.Rollback(ctx)
		return 0, err
	}

	return nextStage.Stage.ID, nil
}

func (s *Service) startNextStageRegularPlayoff(ctx context.Context, logger zerolog.Logger, tournament entities.Tournament) (int64, error) {
	stages, err := s.rdbOperations.GetTournamentStageList(ctx, logger, tournament.ID)
	if err != nil {
		return 0, err
	}

	orderedStages, err := s.getOrderedStagesRegularPlayoff(ctx, logger, tournament, stages)
	if err != nil {
		return 0, err
	}

	nextStage, previousStage, foundNextStage, err := checkOrderedStages(orderedStages)
	if err != nil {
		return 0, err
	}

	// если такой не нашелся - либо играется последний этап, либо непредвиденный косяк
	if *foundNextStage == false {
		err = errors.New("не найден следующий этап")
		logger.Error().Err(err).Msg("next stage not found")
		return 0, error_templates.New(err.Error(), err, codes.FailedPrecondition, http.StatusConflict)
	}

	err = helpers.CheckQtyTeamsInRegularPlayoffStage(int(tournament.Rules.RegularPlayoff.PlayoffTeamsCountOnStart), len(previousStage.Winners), int(nextStage.Number), len(stages))
	if err != nil {
		logger.Error().Err(err).Msg("failed to helpers.CheckQtyTeamsInRegularPlayoffStage")
		return 0, err
	}

	// если предыдущий этап не завершён, не даем создать новый
	if previousStage.Stage.IsFinished == false {
		err = errors.New("предыдущий этап не завершён")
		return 0, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	var countFinishedPlayoffStages int32
	for _, stage := range orderedStages {
		if stage.Stage.IsFinished && stage.Number != constant.FirstStage {
			countFinishedPlayoffStages += 1
		}
	}

	// Логика следующая (старый номер этапа -> новый номер этапа):
	// Формула: КОЛ-ВО_КОМАНД_ДЛЯ_ПЕРЕХОДА_В_ПЛЕЙОФФ (задано при создании турнира) / (2 ^ КОЛ-ВО_ПРОЙДЕННЫХ_ЭТАПОВ)
	// Пусть при создании турнира задано, что в playoff перейдут 4 лучшие команды
	// При переходе из regular в playoff (1 -> 2): 4 / (2 ^ 0) = 4 (0 в степени двойки потому, что на данный момент пройден regular-этап => 0 пройденных playoff-этапов)
	// При переходе из playoff-1 в playoff-2 (2 -> 3): 4 / (2 ^ 1) = 2; далее этапов нет, то есть в этом этапе определится победитель

	neededTeamsCountForNextStage := tournament.Rules.RegularPlayoff.PlayoffTeamsCountOnStart / int32(math.Pow(2, float64(countFinishedPlayoffStages)))

	var teams1Ids, teams2Ids []int64
	if previousStage.Number == constant.FirstStage {
		// Для первого этапа playoff (то есть второй этап турнира) команды генерируются по особому алгоритму,
		// см. реализацию в GeneratePairsRegularPlayoff
		teams1Ids, teams2Ids = helpers.GeneratePairsRegularPlayoff(previousStage.Winners[:neededTeamsCountForNextStage], int(nextStage.BestOf))
	} else {
		teams1Ids, teams2Ids = helpers.GeneratePlayoffPairs(previousStage.Winners[:neededTeamsCountForNextStage], int(nextStage.BestOf))
	}

	tx, err := s.rwdbOperations.BeginTx(ctx, logger)
	if err != nil {
		return 0, err
	}

	if err = s.rwdbOperations.CreateFutureTournamentStageGames(ctx, logger, nextStage.Stage.ID, tournament.CityID, teams1Ids, teams2Ids, tx); err != nil {
		tx.Rollback(ctx)
		return 0, err
	}

	if err = tx.Commit(ctx); err != nil {
		tx.Rollback(ctx)
		return 0, err
	}

	return nextStage.Stage.ID, nil
}

func (s *Service) finishStageRegular(ctx context.Context, logger zerolog.Logger, request *entities.FinishStageRequest, tournament entities.Tournament, games []entities.TournamentGame) error {
	var err error

	if err = s.validateGames(ctx, logger, games); err != nil {
		return err
	}

	// игр должно быть определенное кол-во, которое задается при создании турнира в зависимости от параметра bestOf
	teams1ids, _ := helpers.GeneratePairs(tournament.TeamIDs, int(tournament.Rules.Regular.BestOf))
	if len(games) < len(teams1ids) {
		err = errors.New("в этапе турнирной сетки ожидается больше игр")
		logger.Error().Err(err).Msg("not enough games")

		return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	return nil
}

func (s *Service) finishStagePlayoff(ctx context.Context, logger zerolog.Logger, tournament entities.Tournament, games []entities.TournamentGame, curStage entities.TournamentStage) error {
	stages, err := s.rdbOperations.GetTournamentStageList(ctx, logger, tournament.ID)
	if err != nil {
		return err
	}

	if err = s.checkStages(logger, stages, curStage); err != nil {
		return err
	}

	if err = s.validateGames(ctx, logger, games); err != nil {
		return err
	}

	orderedStages, err := s.getOrderedStagesWithWinners(ctx, logger, tournament, stages)
	if err != nil {
		return err
	}

	nextStage, previousStage, foundNextStage, err := checkOrderedStages(orderedStages)
	if err != nil {
		return err
	}

	if err = checkForLastStageFinish(nextStage, foundNextStage); err != nil {
		logger.Error().Err(err).Msg(err.Error())
		return err
	}

	if nextStage == nil && previousStage == nil {
		// завершаем последний этап со всеми сыгранными играми
		return nil
	}

	// если в предыдущем этапе посчиталось неправильное кол-во победителей,
	// то это либо получилась ничья, либо непредвиденный косяк
	err = helpers.CheckQtyWinnersInPreviousPlayoffStage(len(tournament.TeamIDs), len(previousStage.Winners), int(nextStage.Number), len(stages))
	if err != nil {
		logger.Error().Err(err).Msg("failed to helpers.CheckQtyWinners")
		return err
	}

	if err = checkPlayoffTeamsCountWithGeneratedTeams(previousStage.Winners, int(nextStage.BestOf), helpers.GeneratePlayoffPairs); err != nil {
		logger.Error().Err(err).Msg(err.Error())

		return err
	}

	return nil
}

func (s *Service) finishStageRegularPlayoff(ctx context.Context, logger zerolog.Logger, tournament entities.Tournament, stageGames []entities.TournamentGame, curStage entities.TournamentStage) error {
	stages, err := s.rdbOperations.GetTournamentStageList(ctx, logger, tournament.ID)
	if err != nil {
		return err
	}

	if err = s.checkStages(logger, stages, curStage); err != nil {
		return err
	}

	if err = s.validateGames(ctx, logger, stageGames); err != nil {
		return err
	}

	orderedStages, err := s.getOrderedStagesRegularPlayoff(ctx, logger, tournament, stages)
	if err != nil {
		return err
	}

	nextStage, previousStage, foundNextStage, err := checkOrderedStages(orderedStages)
	if err != nil {
		return err
	}

	if err = checkForLastStageFinish(nextStage, foundNextStage); err != nil {
		logger.Error().Err(err).Msg(err.Error())
		return err
	}

	if nextStage == nil && previousStage == nil {
		// завершаем последний этап со всеми сыгранными играми
		return nil
	}

	err = helpers.CheckQtyTeamsInRegularPlayoffStage(int(tournament.Rules.RegularPlayoff.PlayoffTeamsCountOnStart), len(previousStage.Winners), int(nextStage.Number), len(stages))
	if err != nil {
		logger.Error().Err(err).Msg("failed to helpers.CheckQtyTeamsInRegularPlayoffStage")
		return err
	}

	return nil
}

// validateGames Проверяет, что все игры этапа начаты и у них есть хотя бы 1 матч.
// Используется при старте нового этапа или завершении текущего в playoff или regular+playoff турнирах.
func (s *Service) validateGames(ctx context.Context, logger zerolog.Logger, games []entities.TournamentGame) error {
	var err error

	for _, game := range games {
		if game.Date == nil || game.Date.After(time.Now()) {
			msg := fmt.Sprintf("дата игры с id=%d позже текущей даты", game.ID)
			err = errors.New(msg)
			logger.Error().Err(err).Msg(msg)

			return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}

		matches, err := s.rdbOperations.FetchMatches(logger, ctx, game.ID)
		if err != nil {
			return err
		}
		if game.TechLooseTeamID == nil && len(matches) == 0 {
			msg := fmt.Sprintf("обнаружена игра id=%d без матчей", game.ID)
			err = errors.New(msg)
			logger.Error().Err(err).Msg(msg)

			return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}
	}

	return nil
}

// getOrderedStagesWithWinners Проходит по всем играм всех этапов, определяет на каждом этапе победителей.
func (s *Service) getOrderedStagesWithWinners(ctx context.Context, logger zerolog.Logger, tournament entities.Tournament, stages []entities.TournamentStage) ([]entities.StageStat, error) {
	orderedStages := make([]entities.StageStat, len(stages))

	teams, err := s.rdbOperations.GetTournamentTeamList(ctx, logger, tournament.ID)
	if err != nil {
		return nil, err
	}

	teamsExtraPoints := make(map[int64]int64, len(teams))

	for _, team := range teams {
		extraPoints, err := s.rdbOperations.GetTournamentTeamExtraPointsCountV2(ctx, logger, team.ID, tournament.ID)
		if err != nil {
			return nil, err
		}

		teamsExtraPoints[team.ID] = extraPoints
	}

	slices.SortFunc(stages, func(a, b entities.TournamentStage) int {
		return int(a.ID - b.ID)
	})

	for _, stage := range stages {
		sSt, err := s.getGameStatsForStage(ctx, logger, stage.ID, teamsExtraPoints)
		if err != nil {
			return nil, err
		}

		sSt.Winners = helpers.ExtractWinners(sSt.GameStats)

		// парсим порядковый номер этапа
		orderNumberStr := strings.Split(stage.Number, constant.SeparatorStageNumber)
		orderNumber, err := strconv.ParseInt(orderNumberStr[indexZero], 10, 64)
		if err != nil {
			logger.Error().Err(err).Msg(err.Error())

			return nil, error_templates.New(err.Error(), err, codes.Internal, http.StatusInternalServerError)
		}
		if orderNumber <= 0 {
			msg := "failed to parse stage number"
			err := errors.New(msg)
			logger.Error().Err(err).Msg(msg)

			return nil, error_templates.New(err.Error(), err, codes.Internal, http.StatusInternalServerError)
		}

		sSt.Number = orderNumber
		sSt.Stage = stage

		// bestOf у каждого этапа свой
		sSt.BestOf = tournament.Rules.PlayOff.Stages[orderNumber].Bo

		// выстраиваем этапы по порядку
		orderedStages[orderNumber-1] = *sSt
	}

	return orderedStages, nil
}

// getOrderedStagesRegularPlayoff делает то же, что и getOrderedStagesWithWinners, только разница в получении списка победителей.
// Здесь извлекаются не победители, а отсортированные по убыванию рейтинга команды.
func (s *Service) getOrderedStagesRegularPlayoff(ctx context.Context, logger zerolog.Logger, tournament entities.Tournament, stages []entities.TournamentStage) ([]entities.StageStat, error) {
	orderedStages := make([]entities.StageStat, len(stages))

	teams, err := s.rdbOperations.GetTournamentTeamList(ctx, logger, tournament.ID)
	if err != nil {
		return nil, err
	}

	teamsExtraPoints := make(map[int64]int64, len(teams))

	for _, team := range teams {
		extraPoints, err := s.rdbOperations.GetTournamentTeamExtraPointsCountV2(ctx, logger, team.ID, tournament.ID)
		if err != nil {
			return nil, err
		}

		teamsExtraPoints[team.ID] = extraPoints
	}

	slices.SortFunc(stages, func(a, b entities.TournamentStage) int {
		return int(a.ID - b.ID)
	})

	for _, stage := range stages {
		sSt, err := s.getGameStatsForStage(ctx, logger, stage.ID, teamsExtraPoints)
		if err != nil {
			return nil, err
		}

		// парсим порядковый номер этапа
		orderNumberStr := strings.Split(stage.Number, constant.SeparatorStageNumber)
		orderNumber, err := strconv.ParseInt(orderNumberStr[indexZero], 10, 64)
		if err != nil {
			logger.Error().Err(err).Msg(err.Error())

			return nil, error_templates.New(err.Error(), err, codes.Internal, http.StatusInternalServerError)
		}
		if orderNumber <= 0 {
			msg := "failed to parse stage number"
			err := errors.New(msg)
			logger.Error().Err(err).Msg(msg)

			return nil, error_templates.New(err.Error(), err, codes.Internal, http.StatusInternalServerError)
		}

		sSt.Number = orderNumber
		sSt.Stage = stage

		if sSt.Number == constant.FirstStage {
			// для оконченного regular-этапа надо понять, какие топ N команд пройдут в первый playoff-этап (т.е. второй этап)
			sSt.Winners, err = helpers.ExtractWinnersRegularPlayoff(sSt.GameStats, teamsExtraPoints)
			if err != nil {
				return nil, err
			}

			sSt.BestOf = tournament.Rules.RegularPlayoff.Regular.BestOf
		} else {
			// для всех остальных playoff этапов
			sSt.Winners = helpers.ExtractWinners(sSt.GameStats)
			sSt.BestOf = tournament.Rules.RegularPlayoff.PlayOff.Stages[orderNumber-1].Bo
		}

		// выстраиваем этапы по порядку
		orderedStages[orderNumber-1] = *sSt
	}

	return orderedStages, nil
}

func (s *Service) getGameStatsForStage(ctx context.Context, logger zerolog.Logger, stageId int64, teamsExtraPoints map[int64]int64) (*entities.StageStat, error) {
	var sSt entities.StageStat

	stageGames, err := s.rdbOperations.GetTournamentStageGames(ctx, logger, stageId)
	if err != nil {
		return nil, err
	}

	if err = s.validateGames(ctx, logger, stageGames); err != nil {
		return nil, err
	}

	for _, g := range stageGames {
		var gs entities.GameStat
		gs.Game = g

		// если есть тех.поражение, то матчи не смотрим, проверяем кто проиграл и ставим отметки
		if g.TechLooseTeamID != nil {
			if *g.TechLooseTeamID == g.Team1ID {
				gs.WinnerId = g.Team2ID
			} else if *g.TechLooseTeamID == g.Team2ID {
				gs.WinnerId = g.Team1ID
			}

			gs.Matches = nil
		} else { // если тех.поражения нет, то проверяем матчи
			matches, err := s.rdbOperations.FetchMatches(logger, ctx, g.ID)
			if err != nil {
				return nil, err
			}

			var totalScoreTeam1 int64
			var totalScoreTeam2 int64

			gs.Matches = matches

			// суммируем все очки матчей для каждой команды
			for _, m := range matches {
				totalScoreTeam1 += m.ScoreTeam1
				totalScoreTeam2 += m.ScoreTeam2

				// учитываем заработанные доп.очки команды
				totalScoreTeam1 += teamsExtraPoints[m.Team1ID]
				totalScoreTeam2 += teamsExtraPoints[m.Team2ID]
			}

			// если у кого-то есть преимущество по очкам - то он победитель, иначе - ничего не делаем
			if totalScoreTeam1 > totalScoreTeam2 {
				gs.WinnerId = g.Team1ID
			} else if totalScoreTeam2 > totalScoreTeam1 {
				gs.WinnerId = g.Team2ID
			}
		}

		sSt.GameStats = append(sSt.GameStats, gs)
	}

	return &sSt, nil
}

// checkStages По номеру текущего этапа проверяет, что предыдущие завершены, а также смотрит, чтобы следующий от проверяемого этапа не был завершён.
func (s *Service) checkStages(logger zerolog.Logger, stages []entities.TournamentStage, currentStage entities.TournamentStage) error {
	currentStageNumberInt, _, err := helpers.ParseTournamentStageNumber(currentStage.Number, constant.SeparatorStageNumber)
	if err != nil {
		return err
	}

	for _, stage := range stages {
		stageNumberInt, _, err := helpers.ParseTournamentStageNumber(stage.Number, constant.SeparatorStageNumber)
		if err != nil {
			return err
		}

		if stageNumberInt < currentStageNumberInt && !stage.IsFinished {
			msg := fmt.Sprintf("предыдущий этап %s не завершён", stage.Number)
			err := errors.New(msg)
			logger.Error().Err(err).Msg(msg)

			return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		} else if stageNumberInt > currentStageNumberInt && stage.IsFinished {
			msg := "следующий этап уже завершён"
			err := errors.New(msg)
			logger.Error().Err(err).Msg(msg)

			return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}

		if stageNumberInt > currentStageNumberInt {
			break
		}
	}

	return nil
}

/*local functions*/

func validateRulesTypesIds(rules entities.TournamentRule, typeId int64, teamsIds, playersIds []int64) error {
	regular := rules.Regular
	playOff := rules.PlayOff
	regularPlayoff := rules.RegularPlayoff

	if regular == nil && playOff == nil && regularPlayoff == nil {
		err := errors.New("правила не установлены")
		return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	if len(playersIds) > 0 && len(teamsIds) > 0 {
		err := errors.New("в запросе могут только игроки или только команды")
		return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	if typeId == constant.RegularTournamentTypeID {
		if len(teamsIds) == 0 {
			err := errors.New("для турнира типа Regular команды обязательны")
			return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}

		if regular == nil {
			err := errors.New("правила для турнира типа Regular не заполнены")
			return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}

		if playOff != nil {
			err := errors.New("для турнира типа Regular поле \"playOff\" не должно быть заполнено")
			return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}

		return nil
	}

	if typeId == constant.RegularOneVsOneTournamentTypeID {
		if len(playersIds) == 0 && len(teamsIds) == 0 {
			err := errors.New("для турнира типа Regular.OneVsOne обязательны или игроки или команды")
			return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}

		if regular == nil {
			err := errors.New("правила для турнира типа Regular.OneVsOne не заполнены")
			return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}

		if playOff != nil {
			err := errors.New("для турнира типа Regular.OneVsOne поле \"playOff\" не должно быть заполнено")
			return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}

		return nil
	}

	return nil
}

func haveFinishedStage(stages []entities.TournamentStage) bool {
	for _, stage := range stages {
		if stage.IsFinished == true {
			return true
		}
	}
	return false
}

func haveStartedGames(games []entities.TournamentGame) bool {
	for _, game := range games {
		if game.Date != nil && time.Now().After(*game.Date) {
			return true
		}
	}
	return false
}

func checkPlayoffTeamsCountWithGeneratedTeams(winners []int64, bestOf int, generatorPairsFunc func([]int64, int) ([]int64, []int64)) error {
	teams1Ids, teams2Ids := generatorPairsFunc(winners, bestOf)

	if len(winners)*bestOf != (len(teams1Ids) + len(teams2Ids)) {
		msg := "не совпадает текущее количество команд в этапе с рассчитанным для этого же этапа"
		err := errors.New(msg)

		return error_templates.New(err.Error(), err, codes.FailedPrecondition, http.StatusConflict)
	}

	return nil
}

func checkForLastStageFinish(nextStage *entities.StageStat, foundNextStage *bool) error {
	if *foundNextStage == false && nextStage != nil {
		// если этап не найден, но есть заполненный следующий этап, то произошёл косяк

		err := errors.New("не найден следующий этап")
		return error_templates.New(err.Error(), err, codes.Internal, http.StatusInternalServerError)
	}

	// завершаем работу метода, т.к. все условия для завершения последнего этапа соблюдены
	return nil
}

// checkOrderedStages проверяет orderedStages при старте нового этапа или завершении текущего в StartNextStage или FinishStage
func checkOrderedStages(orderedStages []entities.StageStat) (*entities.StageStat, *entities.StageStat, *bool, error) {
	var err error

	var nextStage *entities.StageStat
	var previousStage *entities.StageStat
	var foundNextStage bool

	for i, stage := range orderedStages {
		if len(stage.GameStats) == 0 && i == indexZero {
			err = errors.New("первый этап не может быть без игр")
			return nil, nil, nil, error_templates.New(err.Error(), err, codes.Internal, http.StatusInternalServerError)
		}

		if len(stage.GameStats) == 0 && i > indexZero {
			nextStage = &stage
			previousStage = &orderedStages[i-1]

			if len(previousStage.GameStats) == 0 {
				err = errors.New("предыдущий этап не может быть без игр")
				return nil, nil, nil, error_templates.New(err.Error(), err, codes.Internal, http.StatusInternalServerError)
			}

			for _, g := range previousStage.GameStats {
				if g.Game.Date == nil || time.Now().Before(*g.Game.Date) {
					err = errors.New("предыдущий этап содержит игру с будущей датой")
					return nil, nil, nil, error_templates.New(err.Error(), err, codes.FailedPrecondition, http.StatusConflict)
				}

				if g.Game.TechLooseTeamID == nil && len(g.Matches) == 0 {
					err = errors.New("предыдущий этап содержит игру без матчей и без присуждения тех. поражения")
					return nil, nil, nil, error_templates.New(err.Error(), err, codes.FailedPrecondition, http.StatusConflict)
				}

				if g.Game.IsTiebreak && len(g.Matches) == 0 {
					err = errors.New("есть несыгранные ничьи в предыдущем этапе")
					return nil, nil, nil, error_templates.New(err.Error(), err, codes.FailedPrecondition, http.StatusConflict)
				}
			}

			foundNextStage = true
			break
		}
	}

	return nextStage, previousStage, &foundNextStage, nil
}
