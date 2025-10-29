package tournament

import (
	"context"
	"errors"
	"fmt"
	"net/http"
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
		master, err := s.rdbOperations.GetTournamentMasterByUserId(logger, ctx, request.Creator.ID, &s.config.RDB)
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
		tournamentID, err := s.createRegular(ctx, *request, logger)
		if err != nil {
			return 0, err
		}

		return tournamentID, nil

		// создание турнира типа Regular.OneVsOne каждый с каждым по одному игроку в команде
	} else if request.TournamentTypeID == constant.RegularOneVsOneTournamentTypeID {
		tournamentID, err := s.createRegularOneVsOne(ctx, *request, logger)
		if err != nil {
			return 0, err
		}

		return tournamentID, nil

		// создание турнира типа Playoff
		// автоматом создаются все этапы турнира и все игры первого этапа турнира
	} else if request.TournamentTypeID == constant.PlayoffTournamentTypeID {
		tournamentID, err := s.createPlayoff(ctx, *request, logger)
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

		// если это админы и турнир начат можно обновить минимальную инфу без перетасовки команд и изменения типа и сетки турнира
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
	}

	if request.TournamentTypeID == constant.RegularOneVsOneTournamentTypeID {
		err = s.updateRegularOneVsOne(ctx, logger, *request, tournament)
		if err != nil {
			return err
		}

		return nil
	}

	if request.TournamentTypeID == constant.PlayoffTournamentTypeID {
		err = s.updatePlayoff(ctx, logger, *request, tournament)
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
	tournament, err := s.rdbOperations.GetTournamentById(logger, ctx, request.ID, &s.config.RDB)
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

	stage, err := s.rdbOperations.GetTournamentStage(logger, ctx, request.ID, &s.config.RDB)
	if err != nil {
		return err
	}

	if stage.IsFinished == true {
		err = errors.New("этап уже завершён")
		logger.Error().Err(err).Msg("stage already finished")
		return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	tournament, err := s.rdbOperations.GetTournamentById(logger, ctx, stage.TournamentID, &s.config.RDB)
	if err != nil {
		return err
	}

	games, err := s.rdbOperations.GetTournamentStageGames(logger, ctx, request.ID, &s.config.RDB)
	if err != nil {
		return err
	}

	// проверка соответствия города у мастера
	if request.Finisher.Role.Name == constant.TournamentMaster {
		master, err := s.rdbOperations.GetTournamentMasterByUserId(logger, ctx, request.Finisher.ID, &s.config.RDB)
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

		for _, g := range games {
			// игра не должна быть позже времени финиша этапа
			if g.Date.After(time.Now()) {
				err = errors.New("дата игры позже текущей даты")
				logger.Error().Err(err).Msg("game date after now")
				return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
			}

			// сыгранная игра не может быть без матчей
			matches, err := s.rdbOperations.FetchMatches(logger, ctx, g.ID)
			if err != nil {
				return err
			}
			if len(matches) == 0 {
				err = errors.New("этап с игрой без матчей не может быть завершен")
				logger.Error().Err(err).Msg("game without matches")
				return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
			}
		}

		// игр должно быть определенное кол-во, которое задается при создании турнира в зависимости от параметра bestOf
		teams1ids, _ := helpers.GeneratePairs(tournament.TeamIDs, int(tournament.Rules.Regular.BestOf))
		if len(games) < len(teams1ids) {
			err = errors.New("в этапе турнирной сетки ожидается больше игр")
			logger.Error().Err(err).Msg("not enough games")
			return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}

		err = s.rwdbOperations.UpdateTournamentStage(logger, ctx, entities.NullableStage{
			ID:           request.ID,
			TournamentID: nil,
			IsFinished:   pointer.GetPointer(true),
		})
		if err != nil {
			return err
		}

		return nil
	}

	return errors.New("tournament types 2, 3, 4 not implemented")
}

func (s *Service) GetTournamentStageList(ctx context.Context, id int64) ([]entities.TournamentStageItem, error) {
	logger := s.logger.With().Str("service", "GetTournamentStageList").Logger()

	stages, err := s.rdbOperations.GetTournamentStageList(logger, ctx, id)
	if err != nil {
		return nil, err
	}

	var stageItems []entities.TournamentStageItem

	for _, stage := range stages {
		games, err := s.rdbOperations.GetTournamentStageGames(logger, ctx, stage.ID, &s.config.RDB)
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

/*local methods*/

func (s *Service) createRegular(ctx context.Context, request entities.CreateTournamentRequest, logger zerolog.Logger) (int64, error) {
	id, err := s.rwdbOperations.CreateTournament(logger, ctx, request, nil)
	if err != nil {
		return 0, err
	}

	stageId, err := s.rwdbOperations.CreateTournamentStage(logger, ctx, id, oneStage, nil)
	if err != nil {
		return 0, err
	}

	team1IDs, team2IDs := helpers.GeneratePairs(request.TeamsIDs, int(request.Rules.Regular.BestOf))

	err = s.rwdbOperations.CreateFutureTournamentStageGames(logger, ctx, stageId, *request.CityID, team1IDs, team2IDs, nil)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *Service) createRegularOneVsOne(ctx context.Context, request entities.CreateTournamentRequest, logger zerolog.Logger) (int64, error) {
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

		// получаем все команды игрока, если команд нет,ошибки быть не должно
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

		// добавим существующую команду
		if teamAlreadyExist == true {

			teamIds = append(teamIds, int64(existingTeam.ID))

			// если такой команды не нашлось, создаем новую по имени
		} else {

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

	err = s.rwdbOperations.CreateFutureTournamentStageGames(logger, ctx, stageId, *request.CityID, team1IDs, team2IDs, tx)
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

func (s *Service) createPlayoff(ctx context.Context, request entities.CreateTournamentRequest, logger zerolog.Logger) (int64, error) {
	err := helpers.ValidatePlayoffTournamentRulesOnCreate(&request)
	if err != nil {
		logger.Error().Err(err).Msg("helpers.ValidatePlayoffTournamentRulesOnCreate")
		return 0, err
	}

	stageNums, stageBestOfList, err := helpers.SortKeysValues(request.Rules.PlayOff.Stages)
	if err != nil {
		logger.Error().Err(err).Msg("helpers.SortKeysValues")
		err = errors.New("нарушена нумерация этапов")
		return 0, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	tx, err := s.rwdbOperations.BeginTx(ctx, logger)
	if err != nil {
		return 0, err
	}

	tournamentId, err := s.rwdbOperations.CreateTournamentTx(logger, ctx, request, tx)
	if err != nil {
		tx.Rollback(ctx)
		return 0, err
	}

	if err = s.rwdbOperations.AddTeamsToTournamentTx(logger, ctx, request.TeamsIDs, tournamentId, tx); err != nil {
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

		// сохраним id первого этапа,  чтобы потом туда положить готовые игры
		if i == indexZero {
			firstStageId = stageId
		}
	}

	// расставляем пары для первого этапа
	// BestOf берётся из первого элемента мапы реквеста - stages
	teams1Ids, teams2Ids := helpers.GeneratePlayoffPairs(request.TeamsIDs, int(stageBestOfList[indexZero].Bo))
	// создаем игры первого этапа
	if err = s.rwdbOperations.CreateFutureTournamentStageGames(logger, ctx, firstStageId, *request.CityID, teams1Ids, teams2Ids, tx); err != nil {
		tx.Rollback(ctx)
		return 0, err
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

	//todo: рефакторинг транзакций делать тут в первую оередь
	err = s.rwdbOperations.UpdateTournament(logger, ctx, newReq, nil)
	if err != nil {
		return err
	}

	err = s.rwdbOperations.DeleteTournamentGamesTeamLinks(logger, ctx, newReq.ID, nil)
	if err != nil {
		return err
	}

	team1IDs, team2IDs := helpers.GeneratePairs(newReq.TeamsIDs, int(newReq.Rules.Regular.BestOf))

	err = s.rwdbOperations.CreateFutureTournamentStageGames(logger, ctx, tournament.Stages[indexZero].ID, *newReq.CityID, team1IDs, team2IDs, nil)
	if err != nil {
		return err
	}

	err = s.rwdbOperations.UpdateTournamentTeamsLinks(logger, ctx, newReq.TeamsIDs, newReq.ID, nil)
	if err != nil {
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
		err = s.rwdbOperations.CreateFutureTournamentStageGames(logger, ctx, tournament.Stages[indexZero].ID, *newReq.CityID, team1IDs, team2IDs, tx)
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
		err = s.rwdbOperations.CreateFutureTournamentStageGames(logger, ctx, tournament.Stages[indexZero].ID, *newReq.CityID, team1IDs, team2IDs, tx)
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
		if err = s.rwdbOperations.CreateFutureTournamentStageGames(logger, ctx, tournament.Stages[indexZero].ID, *newReq.CityID, team1IDs, team2IDs, tx); err != nil {
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

	// если пришли новые команды, или изменились правила или город - всё перезаписываем
	if len(req.TeamsIDs) > 0 || req.Rules != nil || req.CityID != nil {
		// старые команды отвязать от турнира,
		if err = s.rwdbOperations.UnlinkTeamsFromTournament(logger, ctx, tournament.ID, tx); err != nil {
			tx.Rollback(ctx)
			return err
		}

		// удалить игры старых команд
		if err = s.rwdbOperations.DeleteTournamentGames(logger, ctx, tournament.ID, tx); err != nil {
			tx.Rollback(ctx)
			return err
		}

		// удалить этапы утрнира
		if err = s.rwdbOperations.DeleteTournamentStages(logger, ctx, tournament.ID, tx); err != nil {
			tx.Rollback(ctx)
			return err
		}

		// добавим новые команды в турнир
		if err = s.rwdbOperations.AddTeamsToTournamentTx(logger, ctx, newReq.TeamsIDs, tournament.ID, tx); err != nil {
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
		if err = s.rwdbOperations.CreateFutureTournamentStageGames(logger, ctx, firstStageId, *newReq.CityID, teams1Ids, teams2Ids, tx); err != nil {
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

func (s *Service) checkTournamentMasterCredentials(ctx context.Context, logger zerolog.Logger, req *entities.UpdateTournamentRequest, tournament entities.Tournament, games []entities.TournamentGame) error {
	master, err := s.rdbOperations.GetTournamentMasterByUserId(logger, ctx, req.Executor.ID, &s.config.RDB)
	if err != nil {
		return err
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

		matches, err := s.rdbOperations.GetMatchListByGameID(ctx, logger, game.ID, &s.config.RDB)
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
	master, err := s.rdbOperations.GetTournamentMasterByUserId(logger, ctx, req.Executor.ID, &s.config.RDB)
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

		matches, err := s.rdbOperations.GetMatchListByGameID(ctx, logger, game.ID, &s.config.RDB)
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

/*local functions*/

func validateRulesTypesIds(rules entities.TournamentRule, typeId int64, teamsIds, playersIds []int64) error {
	regular := rules.Regular
	playOff := rules.PlayOff

	if regular == nil && playOff == nil {
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
