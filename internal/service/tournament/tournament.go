package tournament

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/constant"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/convert"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/error_templates"
	pkgerr "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/errors"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/helpers"
)

const (
	bestOfOne int64 = 1
	indexZero int64 = 0
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

	// чтобы не валидировать соответствие типа и конструктора турнира, выявляем тип сами тут
	request.TournamentTypeID, err = setTournamentType(request.Rules)
	if err != nil {
		logger.Error().Err(err).Msg("failed to set tournament type")
		return 0, err
	}

	// если запрос от мастера по ткрнирам, то тщательно проверяем соответствие города мастера
	// городу из запроса, которого ожидается, что не будет указано,но на всякий случай
	if request.Creator.Role.Name == constant.TournamentMaster {
		master, err := s.rdbOperations.GetTournamentMasterByUserId(logger, ctx, request.Creator.ID, &s.config.RDB)
		if err != nil {
			return 0, err
		}

		if request.CityIDParam == "" {

			request.CityID = master.City.ID

		} else {

			cityId, err := strconv.ParseInt(request.CityIDParam, 10, 64)
			if err != nil {
				err = errors.New(pkgerr.WrongParameterError + ": " + "cityId")
				logger.Error().Err(err).Msg("Failed strconv.ParseInt in tournament.Create")
				return 0, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
			}

			if cityId != master.City.ID {
				err = errors.New(pkgerr.ErrCityIdNotEqualMasterCityId)
				logger.Error().Err(err).Msg("cityIdParam not equal masterCityId in tournament.Create")
				return 0, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
			}

			request.CityID = cityId
		}

	} else {
		// если это суперюзер или админ(мидлвэар так настроен) то валидируем город
		// валидация перенесена в сервисный слой из-за мастера
		if request.CityIDParam == "" {
			err = errors.New(pkgerr.EmptyParameterError + ": " + "cityId")
			logger.Error().Err(err).Msg("CityId is required in tournament.Create")
			return 0, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}

		cityId, err := strconv.ParseInt(request.CityIDParam, 10, 64)
		if err != nil {
			err = errors.New(pkgerr.WrongParameterError + ": " + "cityId")
			logger.Error().Err(err).Msg("Failed strconv.ParseInt in tournament.Create")
			return 0, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}

		request.CityID = cityId
	}

	// создание турнира типа Regular, BestOf=1 - каждый с каждым по разу
	if request.Rules.Regular != nil && request.TournamentTypeID == constant.RegularTournamentTypeID && request.Rules.Regular.BestOf == bestOfOne {
		tournamentID, err := s.createRegularBestOfOne(ctx, request, logger)
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

	tournament, err := s.rdbOperations.GetTournamentById(logger, ctx, request.ID, &s.config.RDB)
	if err != nil {
		return err
	}

	games, err := s.rdbOperations.GetTournamentGameList(logger, ctx, tournament.ID, &s.config.RDB)
	if err != nil {
		return err
	}

	// обновление мастером возможно только в неначатых турнирах
	if request.Executor.Role.Name == constant.TournamentMaster {
		err = s.checkTournamentMasterCredentials(ctx, request, logger, tournament, games)
		if err != nil {
			return err
		}
	} else {

		// если это админы и турнир начат можно обновить минимальную инфу без перетасовки команд и изменения типа и сетки турнира
		if helpers.HaveFinishedStage(tournament.Stages) == true || helpers.HaveStartedGames(games) == true {
			if request.CityID != nil || request.Rules != nil || request.TeamIDs != nil {
				err = fmt.Errorf("турнир с завершенными или начатыми играми не могут обновляться поля: cityId, rules, teamIds")
				logger.Error().Err(err).Msg("Failed tournament.Update: finished stage or started games	")
				return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
			}

			err = s.rwdbOperations.UpdateTournament(logger, ctx, request, &s.config.RDB)
			if err != nil {
				return err
			}

			return nil
		}
	}

	// если турнир не начат, то менять его могут и мастер и админы
	// на основании старого турнира и запроса на изменения формируем данные для обновления(приоритет у новых данных само собой)
	newReq, err := buildNewRequest(request, tournament)
	if err != nil {
		return err
	}

	// обновление пока только для турнира типа Regular, BestOf=1 - каждый с каждым по разу
	if (newReq.TournamentTypeID != nil && *newReq.TournamentTypeID == constant.RegularTournamentTypeID) && (newReq.Rules.Regular != nil && newReq.Rules.Regular.BestOf == bestOfOne) {
		err = s.updateRegularBestOfOne(ctx, newReq, tournament.Stages[indexZero].ID, logger)
		if err != nil {
			return err
		}

		return nil
	}

	return errors.New("not implemented")
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	logger := s.logger.With().Str("service", "Delete").Logger()

	games, err := s.rdbOperations.GetTournamentGameList(logger, ctx, id, &s.config.RDB)
	if err != nil {
		return err
	}

	err = s.rwdbOperations.DeleteTournamentRating(logger, ctx, id, &s.config.RWDB)
	if err != nil {
		return err
	}

	for _, g := range games {
		err = s.rwdbOperations.DeleteGameMatches(logger, ctx, g.ID, &s.config.RWDB)
		if err != nil {
			return err
		}
	}

	err = s.rwdbOperations.DeleteTournamentCascade(logger, ctx, id, &s.config.RWDB)
	if err != nil {
		return err
	}

	return nil
}

/*local methods*/

func (s *Service) createRegularBestOfOne(ctx context.Context, request *entities.CreateTournamentRequest, logger zerolog.Logger) (int64, error) {
	id, err := s.rwdbOperations.CreateRegularTournament(logger, ctx, request, &s.config.RWDB)
	if err != nil {
		return 0, err
	}

	stageId, err := s.rwdbOperations.CreateTournamentStage(logger, ctx, id, false, &s.config.RWDB)
	if err != nil {
		return 0, err
	}

	team1IDs, team2IDs := helpers.GenerateUniquePairs(request.TeamIDs)

	err = s.rwdbOperations.CreateFutureTournamentStageGames(logger, ctx, stageId, request.CityID, team1IDs, team2IDs, &s.config.RWDB)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *Service) updateRegularBestOfOne(ctx context.Context, newReq *entities.UpdateTournamentRequest, stageId int64, logger zerolog.Logger) error {

	// если роль не мастер(для мастера проверки в другом месте), то надо проверить соответствие городов
	if newReq.Executor.Role.Name != constant.TournamentMaster {
		err := s.cityIdTeamCityIdComparison(ctx, logger, newReq.CityID, newReq.TeamIDs, newReq.ID)
		if err != nil {
			return err
		}
	}

	err := s.rwdbOperations.DeleteTournamentGamesTeamLinks(logger, ctx, newReq.ID, &s.config.RWDB)
	if err != nil {
		return err
	}

	team1IDs, team2IDs := helpers.GenerateUniquePairs(newReq.TeamIDs)

	err = s.rwdbOperations.CreateFutureTournamentStageGames(logger, ctx, stageId, *newReq.CityID, team1IDs, team2IDs, &s.config.RWDB)
	if err != nil {
		return err
	}

	err = s.rwdbOperations.UpdateTournamentTeamsLinks(logger, ctx, newReq.TeamIDs, newReq.ID, &s.config.RWDB)
	if err != nil {
		return err
	}

	err = s.rwdbOperations.UpdateTournament(logger, ctx, newReq, &s.config.RWDB)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) checkTournamentMasterCredentials(ctx context.Context, request *entities.UpdateTournamentRequest, logger zerolog.Logger, tournament entities.Tournament, games []entities.TournamentGame) error {
	master, err := s.rdbOperations.GetTournamentMasterByUserId(logger, ctx, request.Executor.ID, &s.config.RDB)
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

	if request.CityID != nil && *request.CityID != master.City.ID {
		err = fmt.Errorf(
			"город в запросе(ID: %d) не совпадает с городом мастера по турнирам(ID: %d)",
			*request.CityID,
			master.City.ID,
		)
		logger.Error().Err(err).Msg("Failed tournament.Update by tournament master")
		return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	if helpers.HaveFinishedStage(tournament.Stages) == true {
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

func (s *Service) cityIdTeamCityIdComparison(ctx context.Context, logger zerolog.Logger, cityId *int64, teamIDs []int64, tournamentId int64) error {
	var cityIdLocal int64
	if cityId != nil {
		cityIdLocal = *cityId
	} else {
		tournament, err := s.rdbOperations.GetTournamentById(logger, ctx, tournamentId, &s.config.RDB)
		if err != nil {
			return err
		}
		cityIdLocal = tournament.CityID
	}

	for _, tId := range teamIDs {
		team, err := s.rdbOperations.GetTeamById(logger, ctx, tId, &s.config.RDB)
		if err != nil {
			return err
		}
		if team.CityId == nil || *team.CityId != cityIdLocal {
			err = fmt.Errorf("город турнира(ID: %d) и город команды(ID: %s) не совпадают",
				cityId,
				convert.IntPtrToStr(team.CityId),
			)
			logger.Error().Err(err).Msg("failed teamId and cityId comparison")
			return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}
	}

	return nil
}

/*local functions*/

func setTournamentType(rules entities.TournamentRule) (int64, error) {
	if rules.Regular != nil && rules.PlayOff == nil {
		return constant.RegularTournamentTypeID, nil
	} else if rules.Regular == nil && rules.PlayOff != nil {
		return constant.PlayoffTournamentTypeID, nil
	} else {
		err := errors.New("турнир может быть только регулярным либо только на вылет")
		return 0, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}
}

func buildNewRequest(request *entities.UpdateTournamentRequest, tournament entities.Tournament) (*entities.UpdateTournamentRequest, error) {
	newReq := &entities.UpdateTournamentRequest{
		Executor: request.Executor,
		ID:       request.ID,
	}

	if request.CityID != nil {
		newReq.CityID = request.CityID
	} else {
		newReq.CityID = &tournament.CityID
	}

	if request.SeasonID != nil {
		newReq.SeasonID = request.SeasonID
	} else {
		newReq.SeasonID = &tournament.SeasonID
	}

	if request.Name != nil {
		newReq.Name = request.Name
	} else {
		newReq.Name = &tournament.Name
	}

	if request.TeamIDs != nil {
		newReq.TeamIDs = request.TeamIDs
	} else {
		newReq.TeamIDs = tournament.TeamIDs
	}

	if request.Rules != nil {
		newReq.Rules = request.Rules
	} else {
		newReq.Rules = &tournament.Rules
	}

	typeId, err := setTournamentType(*newReq.Rules)
	if err != nil {
		return nil, err
	}

	newReq.TournamentTypeID = &typeId

	return newReq, nil
}
