package tournament

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"net/http"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/helpers/pointer"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/constant"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/error_templates"
	pkgerr "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/errors"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/helpers"
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
	err = validateRulesTypesIds(request.Rules, request.TournamentTypeID, request.TeamsIDs, request.PlayersIDs)
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

	for _, tId := range request.TeamsIDs {
		team, err := s.rdbOperations.GetTeamById(logger, ctx, tId, &s.config.RDB)
		if err != nil {
			return 0, err
		}

		if pointer.GetValue(team.CityId) != request.CityID {
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

	games, err := s.rdbOperations.GetTournamentGameList(logger, ctx, tournament.ID, &s.config.RDB)
	if err != nil {
		return err
	}

	haveMatches := false
	for _, game := range games {
		matches, err := s.rdbOperations.FetchMatches(logger, ctx, game.ID, &s.config.RDB)
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
			if request.CityID != nil || request.Rules != nil || request.TeamsIDs != nil {
				err = fmt.Errorf("турнир с завершенными или начатыми играми не могут обновляться поля: cityId, rules, teamIds")
				logger.Error().Err(err).Msg("Failed tournament.Update: finished stage or started games	")
				return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
			}

			err = s.rwdbOperations.UpdateTournament(logger, ctx, *request, &s.config.RDB)
			if err != nil {
				return err
			}

			return nil
		}
	}

	// обновление для турнира типа Regular, у него может быть только один этап
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

	return errors.New("not implemented")
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	logger := s.logger.With().Str("service", "Delete").Logger()

	games, err := s.rdbOperations.GetTournamentGameList(logger, ctx, id, &s.config.RDB)
	if err != nil {
		return err
	}

	tournament, err := s.rdbOperations.GetTournamentById(logger, ctx, id, &s.config.RDB)
	if err != nil {
		return err
	}

	// удаляем рейтинг
	err = s.rwdbOperations.DeleteTournamentRating(logger, ctx, id, &s.config.RWDB)
	if err != nil {
		return err
	}

	for _, g := range games {
		// удаляем матчи игр турнира
		err = s.rwdbOperations.DeleteGame(logger, ctx, g.ID, &s.config.RWDB)
		if err != nil {
			return err
		}
	}

	// если это турнир 1vs1, то и все его команды удаляем
	for _, tId := range tournament.TeamIDs {
		if tournament.TypeID == constant.RegularOneVsOneTournamentTypeID {
			err = s.rwdbOperations.DeleteTournamentTeamCascade(logger, ctx, tId, &s.config.RWDB)
			if err != nil {
				return err
			}
		}
	}

	// удаляются связи турнир-команда, игры, этапы, и сам турнир
	err = s.rwdbOperations.DeleteTournamentCascade(logger, ctx, id, &s.config.RWDB)
	if err != nil {
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
			matches, err := s.rdbOperations.FetchMatches(logger, ctx, g.ID, &s.config.RDB)
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
		}, &s.config.RDB)
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

/*local methods*/

func (s *Service) createRegular(ctx context.Context, request entities.CreateTournamentRequest, logger zerolog.Logger) (int64, error) {
	id, err := s.rwdbOperations.CreateTournament(logger, ctx, request, &s.config.RWDB)
	if err != nil {
		return 0, err
	}

	stageId, err := s.rwdbOperations.CreateTournamentStage(logger, ctx, id, &s.config.RWDB)
	if err != nil {
		return 0, err
	}

	team1IDs, team2IDs := helpers.GeneratePairs(request.TeamsIDs, int(request.Rules.Regular.BestOf))

	err = s.rwdbOperations.CreateFutureTournamentStageGames(logger, ctx, stageId, request.CityID, team1IDs, team2IDs, &s.config.RWDB)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *Service) createRegularOneVsOne(ctx context.Context, request entities.CreateTournamentRequest, logger zerolog.Logger) (int64, error) {
	teamIds := make([]int64, 0, len(request.TeamsIDs))

	players := make([]entities.Player, 0, len(request.PlayersIDs))

	for _, pId := range request.PlayersIDs {

		player, err := s.rdbOperations.GetPlayerByID(logger, ctx, int(pId), &s.config.RWDB)
		if err != nil {
			return 0, err
		}

		if int64(pointer.GetValue(player.CityID)) != request.CityID {
			err = fmt.Errorf("id города игрока %s(%d) не совпадает с id города турнира(%d)", pointer.GetValue(player.Name), pointer.GetValue(player.CityID), request.CityID)
			return 0, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}

		players = append(players, player)
	}

	for _, p := range players {
		suffix, err := s.rdbOperations.GetSuffix(logger, ctx, &s.config.RDB)
		if err != nil {
			return 0, err
		}

		name, shortName := buildTeamName(p, suffix)

		teamId, err := s.rwdbOperations.CreateTeam(logger, ctx, entities.CreateTeamRequest{
			Name:      name,
			ShortName: shortName,
			CityId:    request.CityID,
		}, &s.config.RWDB)
		if err != nil {
			return 0, err
		}

		_, err = s.rwdbOperations.AddPlayerIntoTeam(logger, ctx, int64(p.ID), teamId, &s.config.RWDB)
		if err != nil {
			return 0, err
		}

		teamIds = append(teamIds, teamId)
	}

	request.TeamsIDs = teamIds

	id, err := s.rwdbOperations.CreateTournament(logger, ctx, request, &s.config.RWDB)
	if err != nil {
		return 0, err
	}

	stageId, err := s.rwdbOperations.CreateTournamentStage(logger, ctx, id, &s.config.RWDB)
	if err != nil {
		return 0, err
	}

	team1IDs, team2IDs := helpers.GeneratePairs(request.TeamsIDs, int(request.Rules.Regular.BestOf))

	err = s.rwdbOperations.CreateFutureTournamentStageGames(logger, ctx, stageId, request.CityID, team1IDs, team2IDs, &s.config.RWDB)
	if err != nil {
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
	if err = s.rwdbOperations.CreateFutureTournamentStageGamesTx(logger, ctx, stageId, request.CityID, teams1Ids, teams2Ids, tx); err != nil {
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

	err = s.rwdbOperations.UpdateTournament(logger, ctx, newReq, &s.config.RWDB)
	if err != nil {
		return err
	}

	err = s.rwdbOperations.DeleteTournamentGamesTeamLinks(logger, ctx, newReq.ID, &s.config.RWDB)
	if err != nil {
		return err
	}

	team1IDs, team2IDs := helpers.GeneratePairs(newReq.TeamsIDs, int(newReq.Rules.Regular.BestOf))

	err = s.rwdbOperations.CreateFutureTournamentStageGames(logger, ctx, tournament.Stages[indexZero].ID, *newReq.CityID, team1IDs, team2IDs, &s.config.RWDB)
	if err != nil {
		return err
	}

	err = s.rwdbOperations.UpdateTournamentTeamsLinks(logger, ctx, newReq.TeamsIDs, newReq.ID, &s.config.RWDB)
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

	err := validateRulesTypesIds(*newReq.Rules, newReq.TournamentTypeID, newReq.TeamsIDs, newReq.PlayersIDs)
	if err != nil {
		return err
	}

	// сначала обновляем турнир, если неправильный сезон, то именно тут возникнет ошибка (транзакций не хватает сильно)
	// в этом запросе логика команд и правил не участвует, если в играх/командах/связях перезапишуться данные
	// а в этом запросе всё упадет из-за сезона, то данные окажутся несогласованными
	// сезон приходит извне, поэтому важно чтобы запрос делался первым
	err = s.rwdbOperations.UpdateTournament(logger, ctx, newReq, &s.config.RWDB)
	if err != nil {
		return err
	}

	var finalTeamIds []int64

	// если пришли новые игроки,то делаем из них новые команды, перезаписываем всё что с ними связано и выходим
	if len(req.PlayersIDs) > 0 {
		for _, pId := range req.PlayersIDs {

			player, err := s.rdbOperations.GetPlayerByID(logger, ctx, int(pId), &s.config.RWDB)
			if err != nil {
				return err
			}

			suffix, err := s.rdbOperations.GetSuffix(logger, ctx, &s.config.RDB)
			if err != nil {
				return err
			}

			name, shortName := buildTeamName(player, suffix)

			teamId, err := s.rwdbOperations.CreateTeam(logger, ctx, entities.CreateTeamRequest{
				Name:      name,
				ShortName: shortName,
				CityId:    *newReq.CityID,
			}, &s.config.RWDB)
			if err != nil {
				return err
			}

			_, err = s.rwdbOperations.AddPlayerIntoTeam(logger, ctx, int64(player.ID), teamId, &s.config.RWDB)
			if err != nil {
				return err
			}

			finalTeamIds = append(finalTeamIds, teamId)
		}

		// перезаписываем всё, и команды в том числе удаляем
		err = s.rewriteGamesCommandsTeamsLinks(ctx, logger, newReq, tournament, finalTeamIds)
		if err != nil {
			return err
		}

		return nil
	}

	// если в запросе пришли команды то тоже перезаписываем все данные и выходим
	if len(req.TeamsIDs) > 0 {
		for _, tId := range newReq.TeamsIDs {
			team, err := s.rdbOperations.GetTeamById(logger, ctx, tId, &s.config.RWDB)
			if err != nil {
				return err
			}

			finalTeamIds = append(finalTeamIds, team.Id)
		}

		// перезаписываем игры и связи, команды неудаляем
		err = s.rewriteGamesTeamsLinks(ctx, logger, newReq, tournament, finalTeamIds)
		if err != nil {
			return err
		}

		return nil
	}

	// если команды не поменялись, но поменялись правила формирования турнира или город, то перезаписываем всё и выходим
	if len(req.TeamsIDs) == 0 && (req.Rules != nil || req.CityID != nil) {
		for _, tId := range tournament.TeamIDs {
			team, err := s.rdbOperations.GetTeamById(logger, ctx, tId, &s.config.RWDB)
			if err != nil {
				return err
			}

			finalTeamIds = append(finalTeamIds, team.Id)
		}

		// перезаписываем игры и связи, команды неудаляем
		err = s.rewriteGamesTeamsLinks(ctx, logger, newReq, tournament, finalTeamIds)
		if err != nil {
			return err
		}

		return nil
	}

	// команды не менялись, правила турнира и город тоже
	return nil
}

func (s *Service) checkTournamentMasterCredentials(ctx context.Context, logger zerolog.Logger, req *entities.UpdateTournamentRequest, tournament entities.Tournament, games []entities.TournamentGame) error {
	master, err := s.rdbOperations.GetTournamentMasterByUserId(logger, ctx, req.Executor.ID, &s.config.RDB)
	if err != nil {
		return err
	}

	if len(req.TeamsIDs) > 0 {
		for _, tId := range req.TeamsIDs {
			team, err := s.rdbOperations.GetTeamById(logger, ctx, tId, &s.config.RDB)
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
			player, err := s.rdbOperations.GetPlayerByID(logger, ctx, int(pId), &s.config.RDB)
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
			team, err := s.rdbOperations.GetTeamById(logger, ctx, reqTeamId, &s.config.RDB)
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
		player, err := s.rdbOperations.GetPlayerByID(logger, ctx, int(pId), &s.config.RDB)
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

func (s *Service) rewriteGamesCommandsTeamsLinks(ctx context.Context, logger zerolog.Logger, newReq entities.UpdateTournamentRequest, tournament entities.Tournament, finalTeamIds []int64) error {
	// удаляем игры старых команд
	err := s.rwdbOperations.DeleteTournamentGamesTeamLinks(logger, ctx, newReq.ID, &s.config.RWDB)
	if err != nil {
		return err
	}

	// удаляем старые команды
	for _, tId := range tournament.TeamIDs {
		err = s.rwdbOperations.DeleteTournamentTeamCascade(logger, ctx, tId, &s.config.RWDB)
		if err != nil {
			return err
		}
	}

	// делаем новые пары
	team1IDs, team2IDs := helpers.GeneratePairs(finalTeamIds, int(newReq.Rules.Regular.BestOf))

	// создаем новые игры
	err = s.rwdbOperations.CreateFutureTournamentStageGames(logger, ctx, tournament.Stages[indexZero].ID, *newReq.CityID, team1IDs, team2IDs, &s.config.RWDB)
	if err != nil {
		return err
	}

	// обновляем связи турнира и новых команд
	err = s.rwdbOperations.UpdateTournamentTeamsLinks(logger, ctx, finalTeamIds, newReq.ID, &s.config.RWDB)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) rewriteGamesTeamsLinks(ctx context.Context, logger zerolog.Logger, newReq entities.UpdateTournamentRequest, tournament entities.Tournament, finalTeamIds []int64) error {
	// удаляем связь турнира с командами и игры турнира
	err := s.rwdbOperations.DeleteTournamentGamesTeamLinks(logger, ctx, newReq.ID, &s.config.RWDB)
	if err != nil {
		return err
	}

	// делаем новые пары
	team1IDs, team2IDs := helpers.GeneratePairs(finalTeamIds, int(newReq.Rules.Regular.BestOf))

	// создаем новые игры
	err = s.rwdbOperations.CreateFutureTournamentStageGames(logger, ctx, tournament.Stages[indexZero].ID, *newReq.CityID, team1IDs, team2IDs, &s.config.RWDB)
	if err != nil {
		return err
	}

	// обновляем связи турнира и новых команд
	err = s.rwdbOperations.UpdateTournamentTeamsLinks(logger, ctx, finalTeamIds, newReq.ID, &s.config.RWDB)
	if err != nil {
		return err
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

func buildTeamName(player entities.Player, suffix string) (string, string) {
	firstRune := func(s string) string {
		bukvy := []rune{'Й', 'Ы', 'Ъ', 'Ь'}

		for _, r := range s {
			if unicode.IsLetter(r) {
				return string(unicode.ToUpper(r))
			} else {
				return string(bukvy[rand.IntN(len(bukvy))])
			}
		}
		return string(bukvy[rand.IntN(len(bukvy))])
	}

	var nameArr []string
	var letters []string

	if player.Name != nil {
		nameArr = append(nameArr, *player.Name)
		letters = append(letters, firstRune(*player.Name))
	}

	if player.SecondName != nil {
		nameArr = append(nameArr, *player.SecondName)
		letters = append(letters, firstRune(*player.SecondName))
	}

	nameArr = append(nameArr, player.LastName)
	letters = append(letters, firstRune(player.LastName))

	name := strings.Join(append(nameArr, suffix), " ")
	short := strings.Join(append(letters, suffix), "")

	return name, short
}
