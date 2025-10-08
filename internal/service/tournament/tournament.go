package tournament

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/constant"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/error_templates"
	pkgerr "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/errors"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/helpers"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/helpers/pointer"
)

const (
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

	// валидируем соответствие типа турнира входящим параметрам
	if err = validateRulesTypesIds(request.Rules, request.TournamentTypeID, request.TeamsIDs, request.PlayersIDs); err != nil {
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
		if request.Creator.Role.Name != constant.SuperUserRole || request.Creator.Role.Name != constant.AdminRole {
			logger.Error().Err(err).Msgf("forbidden for this role: %s", request.Creator.Role.Name)
			return 0, error_templates.New(err.Error(), err, codes.Unauthenticated, http.StatusForbidden)
		}
	}

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

	games, err := s.rdbOperations.GetTournamentGameList(logger, ctx, tournament.ID)
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
				err = fmt.Errorf("турнир с завершенными или начатыми играми не могут обновляться поля: cityId, rules, teamIds, playersIDs")
				logger.Error().Err(err).Msg("Failed tournament.Update: finished stage or started games	")
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

func (s *Service) Delete(ctx context.Context, id int64) error {
	logger := s.logger.With().Str("service", "Delete").Logger()

	tournament, err := s.rdbOperations.GetTournamentById(logger, ctx, id, &s.config.RDB)
	if err != nil {
		return err
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
		games, err := s.rdbOperations.GetTournamentStageGames(logger, ctx, stage.TournamentID, &s.config.RDB)
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

	stageId, err := s.rwdbOperations.CreateTournamentStage(logger, ctx, id, nil)
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
			err = fmt.Errorf("id города игрока %s(%d) не совпадает с id города турнира(%d)", pointer.GetValue(player.Name), pointer.GetValue(player.CityID), request.CityID)
			tx.Rollback(ctx)
			return 0, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}

		players = append(players, player)
	}

	for _, p := range players {
		teamAlreadyExist, existingTeam := false, entities.TeamItem{}

		// получаем все команды игрока, если команл нет,ошибки быть не должно
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

	stageId, err := s.rwdbOperations.CreateTournamentStage(logger, ctx, id, tx)
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

	// создаем первый этап
	stageId, err := s.rwdbOperations.CreateTournamentStageTx(logger, ctx, tournamentId, tx)
	if err != nil {
		tx.Rollback(ctx)
		return 0, err
	}

	// расставляем пары для первого этапа
	teams1Ids, teams2Ids := helpers.GeneratePlayoffPairs(request.TeamsIDs, int(request.Rules.PlayOff.BestOf))

	// создаем игры первого этапа
	if err = s.rwdbOperations.CreateFutureTournamentStageGames(logger, ctx, stageId, *request.CityID, teams1Ids, teams2Ids, tx); err != nil {
		tx.Rollback(ctx)
		return 0, err
	}

	// Т.к. кол-во команд равно степени двойки, то можем вычислить кол-во этапов общее.
	// Создаем второй и последующие этапов, без игр, т.к. победителей заранее знать не можем.
	// Делим команды на два пока их не остнется две (финальный этап между финалистами)
	for numGames := len(teams1Ids) / 2; numGames > 1; numGames = numGames / 2 {
		_, err = s.rwdbOperations.CreateTournamentStageTx(logger, ctx, tournamentId, tx)
		if err != nil {
			tx.Rollback(ctx)
			return 0, err
		}
	}

	if err = tx.Commit(ctx); err != nil {
		logger.Error().Err(err).Msg("failed tx.Commit")
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
	if len(req.PlayersIDs) > 0 {
		err := errors.New("для обновления тура типа Playoff playersIDs не должны быть заполнены")
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

	err = s.rwdbOperations.UpdateTournament(logger, ctx, newReq, tx)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	// если пришли новые команды, игры удалить, удалить этапы
	if len(req.TeamsIDs) > 0 {
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
		if err = s.rwdbOperations.AddTeamsToTournamentTx(logger, ctx, req.TeamsIDs, tournament.ID, tx); err != nil {
			tx.Rollback(ctx)
			return err
		}

		// первый этап
		stageId, err := s.rwdbOperations.CreateTournamentStageTx(logger, ctx, tournament.ID, tx)
		if err != nil {
			tx.Rollback(ctx)
			return err
		}

		teams1Ids, teams2Ids := helpers.GeneratePlayoffPairs(req.TeamsIDs, int(newReq.Rules.PlayOff.BestOf))

		// игры первого этапа
		if err = s.rwdbOperations.CreateFutureTournamentStageGames(logger, ctx, stageId, *newReq.CityID, teams1Ids, teams2Ids, tx); err != nil {
			tx.Rollback(ctx)
			return err
		}

		// Т.к. кол-во команд равно степени двойки, то можем вычислить кол-во этапов общее.
		// Создаем второй и последующие этапов, без игр, т.к. победителей заранее знать не можем.
		// Делим команды на два пока их не остнется две (финальный этап между финалистами)
		for numGames := len(teams1Ids) / 2; numGames > 1; numGames = numGames / 2 {
			_, err = s.rwdbOperations.CreateTournamentStageTx(logger, ctx, tournament.ID, tx)
			if err != nil {
				tx.Rollback(ctx)
				return err
			}
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
				err = fmt.Errorf("город(id: %d) команды не совпадает с городом(id: %d) мастера по турнирам)", pointer.GetValue(team.CityId), master.City.ID)
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

	if typeId == constant.PlayoffTournamentTypeID {
		if len(teamsIds) == 0 || len(playersIds) > 0 {
			err := errors.New("для турнира типа Playoff команды обязательны а игроков быть не должно")
			return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}

		if playOff == nil {
			err := errors.New("правила для турнира типа Playoff не заполнены")
			return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		} else {
			if helpers.IsPowTwo(len(teamsIds)) == false {
				err := errors.New("в турнире типа Playoff команд должно быть 2^N штук")
				return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
			}
		}

		if regular != nil {
			err := errors.New("для турнира типа Playoff поле \"regular\" не должно быть заполнено")
			return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}
	} else {
		return errors.New("для типов 3, 4 функционал пока не реализоан")
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
