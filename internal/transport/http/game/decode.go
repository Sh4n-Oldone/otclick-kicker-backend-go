package game

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/valyala/bytebufferpool"
	"google.golang.org/grpc/codes"
	"io"
	"net/http"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/constant"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/error_templates"
	pkgerr "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/errors"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/helpers"
	"strconv"
	"time"
)

var validate = helpers.NewCustomValidator()

func decodeCreateRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := entities.CreateGameRequest{}
	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	cityIdParam := r.URL.Query().Get("cityId")
	if cityIdParam == "" {
		err := errors.New(pkgerr.EmptyParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	cityId, err := strconv.Atoi(cityIdParam)
	if err != nil {
		err = errors.New(pkgerr.WrongParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	isTiebreakParam := r.URL.Query().Get("isTiebreak")
	if isTiebreakParam != "" {
		isTiebreak, err := strconv.ParseBool(isTiebreakParam)
		if err != nil {
			err = errors.New(pkgerr.WrongParameterError)
			return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}
		request.IsTiebreak = isTiebreak
	}
	_, err = io.Copy(buf, r.Body)
	if err != nil {
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	err = json.Unmarshal(buf.Bytes(), &request)
	if err != nil {
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	request.CityID = cityId

	return request, nil
}

func decodeIdParamRequest(_ context.Context, r *http.Request) (interface{}, error) {
	idParam := chi.URLParam(r, "id")
	if idParam == "" {
		err := errors.New(pkgerr.EmptyParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil || id < 1 {
		err = errors.New(pkgerr.WrongParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	return id, nil
}

func decodeUpdateRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := entities.UpdateGameRequest{}
	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	_, err := io.Copy(buf, r.Body)
	if err != nil {
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	err = json.Unmarshal(buf.Bytes(), &request)
	if err != nil {
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	err = validate.Struct(request)
	if err != nil {
		err = errors.New(pkgerr.ValidationErr + ": " + err.Error())
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	if request.Matches != nil {
		for _, match := range request.Matches {
			if match.Team1ID != request.Team1ID || match.Team2ID != request.Team2ID {
				err = errors.New(pkgerr.ValidationErr + ": " + "id команд в игре и матче не совпадают")
				return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
			}
		}
	}

	return request, nil
}

func decodeFindRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := entities.FindGameRequest{}
	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	team1IdParam := r.URL.Query().Get("team1Id")
	if team1IdParam == "" {
		err := errors.New(pkgerr.EmptyParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	team1Id, err := strconv.Atoi(team1IdParam)
	if err != nil || team1Id < 1 {
		err = errors.New(pkgerr.WrongParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	team2IdParam := r.URL.Query().Get("team2Id")
	if team2IdParam == "" {
		err = errors.New(pkgerr.EmptyParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	team2Id, err := strconv.Atoi(team2IdParam)
	if err != nil || team2Id < 1 {
		err = errors.New(pkgerr.WrongParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	request.Team1ID = team1Id
	request.Team2ID = team2Id

	return request, nil
}

func decodeUpdateFutureGameRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := entities.UpdateFutureGameRequest{}
	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	_, err := io.Copy(buf, r.Body)
	if err != nil {
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	err = json.Unmarshal(buf.Bytes(), &request)
	if err != nil {
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	err = validate.Struct(request)
	if err != nil {
		err = errors.New(pkgerr.ValidationErr + ": " + err.Error())
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	if request.PlaceID != nil {
		if *request.PlaceID <= 0 {
			err = errors.New(pkgerr.ValidationErr + ": " + pkgerr.WrongPlaceIdError)
			return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}
	}

	if request.Date != nil {
		date := *request.Date
		if date.Equal(time.Unix(0, 0)) || date.IsZero() || date.Year() == 0 {
			err = errors.New(pkgerr.ValidationErr + ": " + pkgerr.ErrWrongDate)
			return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}
	}

	return request, nil
}

func decodeGamesYearsRequest(_ context.Context, _ *http.Request) (interface{}, error) {
	return nil, nil
}

func decodeGetComingGamesRequest(_ context.Context, _ *http.Request) (interface{}, error) {
	return nil, nil
}

func decodeGetFutureGamesRequest(_ context.Context, r *http.Request) (interface{}, error) {
	cityIdParam := r.URL.Query().Get("cityId")
	if cityIdParam == "" {
		err := errors.New(pkgerr.EmptyParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	cityId, err := strconv.Atoi(cityIdParam)
	if err != nil {
		err = errors.New(pkgerr.WrongParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	if cityId < 1 {
		err = errors.New(pkgerr.WrongParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	return cityId, nil
}

func decodeGetGameListRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := entities.GetGameListRequest{}

	queryParams := r.URL.Query()
	var err error

	// числовые параметры
	if cityIdParam := queryParams.Get("cityId"); cityIdParam != "" {
		if request.CityId, err = parseIntParam(cityIdParam); err != nil {
			return nil, error_templates.BadRequestError(err)
		}
	}

	if leagueIdParam := queryParams.Get("leagueId"); leagueIdParam != "" {
		if request.LeagueId, err = parseIntParam(leagueIdParam); err != nil {
			return nil, error_templates.BadRequestError(err)
		}
	}

	if seasonIdParam := queryParams.Get("seasonId"); seasonIdParam != "" {
		if request.SeasonId, err = parseIntParam(seasonIdParam); err != nil {
			return nil, error_templates.BadRequestError(err)
		}
	}

	if placeIdParam := queryParams.Get("placeId"); placeIdParam != "" {
		if request.PlaceId, err = parseIntParam(placeIdParam); err != nil {
			return nil, error_templates.BadRequestError(err)
		}
	}

	if team1IdParam := queryParams.Get("team1Id"); team1IdParam != "" {
		if request.Team1Id, err = parseIntParam(team1IdParam); err != nil {
			return nil, error_templates.BadRequestError(err)
		}
	}

	if team2IdParam := queryParams.Get("team2Id"); team2IdParam != "" {
		if request.Team2Id, err = parseIntParam(team2IdParam); err != nil {
			return nil, error_templates.BadRequestError(err)
		}
	}

	if sortFieldParam := queryParams.Get("sortField"); sortFieldParam != "" {
		if request.SortField, err = parseIntParam(sortFieldParam); err != nil {
			return nil, error_templates.BadRequestError(err)
		}
	}

	if sortTypeParam := queryParams.Get("sortType"); sortTypeParam != "" {
		if request.SortType, err = parseIntParam(sortTypeParam); err != nil {
			return nil, error_templates.BadRequestError(err)
		}
	}

	if limitParam := queryParams.Get("limit"); limitParam != "" {
		if request.Limit, err = parseIntParam(limitParam); err != nil {
			return nil, error_templates.BadRequestError(err)
		}
	}

	if offsetParam := queryParams.Get("offset"); offsetParam != "" {
		if request.Offset, err = parseIntParam(offsetParam); err != nil {
			return nil, error_templates.BadRequestError(err)
		}
	}

	// булевый параметр
	if isTiebreakParam := queryParams.Get("isTiebreak"); isTiebreakParam != "" {
		if request.IsTiebreak, err = parseBoolParam(isTiebreakParam); err != nil {
			return nil, error_templates.BadRequestError(err)
		}
	}

	// даты
	if dateFromParam := queryParams.Get("dateFrom"); dateFromParam != "" {
		if request.DateFrom, err = parseDateParam(dateFromParam); err != nil {
			return nil, error_templates.BadRequestError(err)
		}
	}

	if dateToParam := queryParams.Get("dateTo"); dateToParam != "" {
		if request.DateTo, err = parseDateParam(dateToParam); err != nil {
			return nil, error_templates.BadRequestError(err)
		}
	}

	return request, nil
}

func decodeCreateFutureGameRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := entities.CreateFutureGameRequest{}
	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	_, err := io.Copy(buf, r.Body)
	if err != nil {
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	err = json.Unmarshal(buf.Bytes(), &request)
	if err != nil {
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	cityIdParam := r.URL.Query().Get("cityId")
	if cityIdParam == "" {
		err = errors.New(pkgerr.EmptyParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	cityId, err := strconv.Atoi(cityIdParam)
	if err != nil {
		err = errors.New(pkgerr.WrongParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	request.CityID = cityId

	isTiebreakParam := r.URL.Query().Get("isTiebreak")
	if isTiebreakParam != "" {
		isTiebreak, err := strconv.ParseBool(isTiebreakParam)
		if err != nil {
			err = errors.New(pkgerr.WrongParameterError)
			return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}
		request.IsTiebreak = isTiebreak
	}

	err = validate.Struct(request)
	if err != nil {
		err = errors.New(pkgerr.ValidationErr + ": " + err.Error())
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	if request.PlaceID != nil {
		if *request.PlaceID <= 0 {
			err = errors.New(pkgerr.ValidationErr + ": " + pkgerr.WrongPlaceIdError)
			return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}
	}

	if request.Date != nil {
		date := *request.Date
		if date.Equal(time.Unix(0, 0)) || date.IsZero() || date.Year() == 0 {
			err = errors.New(pkgerr.ValidationErr + ": " + pkgerr.ErrWrongDate)
			return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}
	}

	return request, nil
}

func decodeGetTeamIDRequest(_ context.Context, r *http.Request) (interface{}, error) {
	paramID := chi.URLParam(r, "team_id")
	if paramID == "" {
		err := errors.New(pkgerr.EmptyParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	id, err := strconv.Atoi(paramID)
	if err != nil || id <= 0 {
		err = errors.New(pkgerr.WrongParameterError)
		return nil, error_templates.New(err.Error(), err, http.StatusBadRequest, http.StatusBadRequest)
	}

	return id, nil
}

func decodeDeleteFutureGameRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := entities.DeleteFutureGameRequest{}

	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	idParam := chi.URLParam(r, "id")
	if idParam == "" {
		err := errors.New(pkgerr.EmptyParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		err = errors.New(pkgerr.WrongParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	_, err = io.Copy(buf, r.Body)
	if err != nil {
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	err = json.Unmarshal(buf.Bytes(), &request)
	if err != nil {
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	request.ID = id

	return request, nil
}

func decodeCreateFutureTournamentGameRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := &entities.CreateFutureTournamentGameRequest{Creator: entities.User{Role: &entities.Role{}, Team: &entities.Team{}}}
	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	reqCtx := r.Context()

	userId, ok := reqCtx.Value(constant.UserIDContextKey).(int64)
	if !ok || userId == 0 {
		err := errors.New(pkgerr.ErrUserIdToken)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}
	request.Creator.ID = userId

	role, ok := reqCtx.Value(constant.RoleNameContextKey).(string)
	if !ok || role == "" {
		err := errors.New(pkgerr.ErrRoleToken)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}
	request.Creator.Role.Name = role

	teamId, ok := reqCtx.Value(constant.TeamIDContextKey).(int64)
	request.Creator.Team.ID = teamId

	_, err := io.Copy(buf, r.Body)
	if err != nil {
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	err = json.Unmarshal(buf.Bytes(), &request)
	if err != nil {
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	return request, nil
}

func decodeUpdateFutureTournamentGameRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := &entities.UpdateFutureTournamentGameRequest{Executor: entities.User{Role: &entities.Role{}, Team: &entities.Team{}}}
	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	reqCtx := r.Context()

	idParam := chi.URLParam(r, "id")
	if idParam == "" {
		err := errors.New(pkgerr.EmptyParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		err = errors.New(pkgerr.WrongParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}
	request.GameID = id

	userId, ok := reqCtx.Value(constant.UserIDContextKey).(int64)
	if !ok || userId == 0 {
		err = errors.New(pkgerr.ErrUserIdToken)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}
	request.Executor.ID = userId

	role, ok := reqCtx.Value(constant.RoleNameContextKey).(string)
	if !ok || role == "" {
		err := errors.New(pkgerr.ErrRoleToken)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}
	request.Executor.Role.Name = role

	teamId, ok := reqCtx.Value(constant.TeamIDContextKey).(int64)
	request.Executor.Team.ID = teamId

	_, err = io.Copy(buf, r.Body)
	if err != nil {
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	err = json.Unmarshal(buf.Bytes(), &request)
	if err != nil {
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	return request, nil
}

func decodeDeleteFutureTournamentGameRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := &entities.DeleteFutureTournamentGameRequest{Executor: entities.User{Role: &entities.Role{}, Team: &entities.Team{}}}
	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	reqCtx := r.Context()

	idParam := chi.URLParam(r, "id")
	if idParam == "" {
		err := errors.New(pkgerr.EmptyParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		err = errors.New(pkgerr.WrongParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}
	request.GameID = id

	userId, ok := reqCtx.Value(constant.UserIDContextKey).(int64)
	if !ok || userId == 0 {
		err = errors.New(pkgerr.ErrUserIdToken)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}
	request.Executor.ID = userId

	role, ok := reqCtx.Value(constant.RoleNameContextKey).(string)
	if !ok || role == "" {
		err := errors.New(pkgerr.ErrRoleToken)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}
	request.Executor.Role.Name = role

	teamId, ok := reqCtx.Value(constant.TeamIDContextKey).(int64)
	request.Executor.Team.ID = teamId

	_, err = io.Copy(buf, r.Body)
	if err != nil {
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	err = json.Unmarshal(buf.Bytes(), &request)
	if err != nil {
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	return request, nil
}

func decodeCreatePlayedTournamentGameRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := &entities.CreatePlayedTournamentGameRequest{Creator: entities.User{Role: &entities.Role{}, Team: &entities.Team{}}}
	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	reqCtx := r.Context()

	userId, ok := reqCtx.Value(constant.UserIDContextKey).(int64)
	if !ok || userId == 0 {
		err := errors.New(pkgerr.ErrUserIdToken)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}
	request.Creator.ID = userId

	role, ok := reqCtx.Value(constant.RoleNameContextKey).(string)
	if !ok || role == "" {
		err := errors.New(pkgerr.ErrRoleToken)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}
	request.Creator.Role.Name = role

	teamId, ok := reqCtx.Value(constant.TeamIDContextKey).(int64)
	request.Creator.Team.ID = teamId

	_, err := io.Copy(buf, r.Body)
	if err != nil {
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	err = json.Unmarshal(buf.Bytes(), &request)
	if err != nil {
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	return request, nil
}

func decodeUpdatePlayedTournamentGameRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := &entities.UpdatePlayedTournamentGameRequest{Executor: entities.User{Role: &entities.Role{}, Team: &entities.Team{}}}
	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	reqCtx := r.Context()

	idParam := chi.URLParam(r, "id")
	if idParam == "" {
		err := errors.New(pkgerr.EmptyParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		err = errors.New(pkgerr.WrongParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}
	request.GameID = id

	userId, ok := reqCtx.Value(constant.UserIDContextKey).(int64)
	if !ok || userId == 0 {
		err = errors.New(pkgerr.ErrUserIdToken)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}
	request.Executor.ID = userId

	role, ok := reqCtx.Value(constant.RoleNameContextKey).(string)
	if !ok || role == "" {
		err := errors.New(pkgerr.ErrRoleToken)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}
	request.Executor.Role.Name = role

	teamId, ok := reqCtx.Value(constant.TeamIDContextKey).(int64)
	request.Executor.Team.ID = teamId

	_, err = io.Copy(buf, r.Body)
	if err != nil {
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	err = json.Unmarshal(buf.Bytes(), &request)
	if err != nil {
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	return request, nil
}

func decodeDeletePlayedTournamentGameRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := &entities.DeletePlayedTournamentGameRequest{Executor: entities.User{Role: &entities.Role{}, Team: &entities.Team{}}}
	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	reqCtx := r.Context()

	idParam := chi.URLParam(r, "id")
	if idParam == "" {
		err := errors.New(pkgerr.EmptyParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		err = errors.New(pkgerr.WrongParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}
	request.GameID = id

	userId, ok := reqCtx.Value(constant.UserIDContextKey).(int64)
	if !ok || userId == 0 {
		err = errors.New(pkgerr.ErrUserIdToken)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}
	request.Executor.ID = userId

	role, ok := reqCtx.Value(constant.RoleNameContextKey).(string)
	if !ok || role == "" {
		err := errors.New(pkgerr.ErrRoleToken)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}
	request.Executor.Role.Name = role

	teamId, ok := reqCtx.Value(constant.TeamIDContextKey).(int64)
	request.Executor.Team.ID = teamId

	_, err = io.Copy(buf, r.Body)
	if err != nil {
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	err = json.Unmarshal(buf.Bytes(), &request)
	if err != nil {
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	return request, nil
}

func decodeGetTournamentGameListRequest(_ context.Context, r *http.Request) (interface{}, error) {
	tournamentIdParam := chi.URLParam(r, "id")
	if tournamentIdParam == "" {
		err := errors.New(pkgerr.EmptyParameterError + ": id")
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	tournamentId, err := strconv.ParseInt(tournamentIdParam, 10, 64)
	if err != nil {
		err = errors.New(pkgerr.WrongParameterError + ": id")
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	return tournamentId, nil
}

/*local functions-helpers*/

func parseIntParam(param string) (*int, error) {
	val, err := strconv.Atoi(param)
	if err != nil {
		return nil, errors.New(pkgerr.WrongParameterError)
	}
	return &val, nil
}

func parseBoolParam(param string) (*bool, error) {
	val, err := strconv.ParseBool(param)
	if err != nil {
		return nil, errors.New(pkgerr.WrongParameterError)
	}
	return &val, nil
}

func parseDateParam(param string) (*time.Time, error) {
	// "2006-01-02" для формата YYYY-MM-DD
	val, err := time.Parse("2006-01-02", param)
	if err != nil {
		return nil, errors.New(pkgerr.ErrWrongDate)
	}
	return &val, nil
}
