package team

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/valyala/bytebufferpool"
	"google.golang.org/grpc/codes"

	cnst "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/constant"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/error_templates"
	pkgerr "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/errors"
)

func decodeGetTeamRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := &entities.GetTeamRequest{}

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

	request.ID = id

	return request, nil
}

func decodeGetTeamsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := &entities.GetTeamsRequest{}

	var onlyFree bool = false
	onlyFreeParam := r.URL.Query().Get("onlyFree")
	if onlyFreeParam == "true" {
		onlyFree = true
	}

	cityIdParam := r.URL.Query().Get("cityId")
	if cityIdParam != "" {
		cityID, err := strconv.ParseInt(cityIdParam, 10, 64)
		if err != nil {
			err = errors.New(pkgerr.WrongParameterError)
			return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}
		request.CityId = cityID
	}

	request.OnlyFree = onlyFree
	return request, nil
}

func decodeGetTeamsByCityRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := &entities.GetTeamsByCityRequest{}

	idParam := chi.URLParam(r, "city_id")
	if idParam == "" {
		err := errors.New(pkgerr.EmptyParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	var onlyFree bool = false
	onlyFreeParam := r.URL.Query().Get("onlyFree")
	if onlyFreeParam == "true" {
		onlyFree = true
	}

	cityID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		err = errors.New(pkgerr.WrongParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	request.OnlyFree = onlyFree
	request.CityID = cityID

	return request, nil
}

func decodeGetTeamsByLeagueRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := &entities.GetTeamsByLeagueRequest{}

	idParam := chi.URLParam(r, "league_id")
	if idParam == "" {
		err := errors.New(pkgerr.EmptyParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	var onlyFree bool = false
	onlyFreeParam := r.URL.Query().Get("onlyFree")
	if onlyFreeParam == "true" {
		onlyFree = true
	}

	leagueID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		err = errors.New(pkgerr.WrongParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	request.OnlyFree = onlyFree
	request.LeagueID = leagueID

	return request, nil
}

func decodeGetTeamVsTeamTableRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := &entities.GetTeamVsTeamTableRequest{}

	cityIdParam := r.URL.Query().Get("cityId")
	if cityIdParam == "" {
		err := errors.New(pkgerr.EmptyParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}
	cityID, err := strconv.ParseInt(cityIdParam, 10, 64)
	if err != nil {
		err := errors.New(pkgerr.WrongParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	seasonIdParam := r.URL.Query().Get("season")
	if seasonIdParam == "" {
		err := errors.New(pkgerr.EmptyParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}
	seasonId, err := strconv.ParseInt(seasonIdParam, 10, 64)
	if err != nil {
		err := errors.New(pkgerr.WrongParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	request.CityID = cityID
	request.SeasonID = seasonId

	return request, nil
}

func decodeCreateRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := &entities.CreateTeamRequest{Creator: entities.User{Role: &entities.Role{}}}
	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	userId, ok := r.Context().Value(cnst.UserIDContextKey).(int64)
	if !ok || userId == 0 {
		err := errors.New(pkgerr.ErrUserIdToken)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}
	request.Creator.ID = userId

	role, ok := r.Context().Value(cnst.RoleNameContextKey).(string)
	if !ok || role == "" {
		err := errors.New(pkgerr.ErrRoleToken)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}
	request.Creator.Role.Name = role

	request.CityIdParam = r.URL.Query().Get("cityId")

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

func decodeUpdateRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := &entities.UpdateTeamRequest{Updater: entities.User{Role: &entities.Role{}}}
	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	userId, ok := r.Context().Value(cnst.UserIDContextKey).(int64)
	if !ok || userId == 0 {
		err := errors.New(pkgerr.ErrUserIdToken)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}
	request.Updater.ID = userId

	role, ok := r.Context().Value(cnst.RoleNameContextKey).(string)
	if !ok || role == "" {
		err := errors.New(pkgerr.ErrRoleToken)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}
	request.Updater.Role.Name = role

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

func decodeDeleteRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := &entities.DeleteTeamRequest{}

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

	request.ID = id

	return request, nil
}

func decodeAddPlayerIntoTeamRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := &entities.MovingPlayerTeam{Executor: entities.User{Role: &entities.Role{}}}

	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	userId, ok := r.Context().Value(cnst.UserIDContextKey).(int64)
	if !ok || userId == 0 {
		err := errors.New(pkgerr.ErrUserIdToken)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}
	request.Executor.ID = userId

	role, ok := r.Context().Value(cnst.RoleNameContextKey).(string)
	if !ok || role == "" {
		err := errors.New(pkgerr.ErrRoleToken)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}
	request.Executor.Role.Name = role

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

func decodeRemovePlayerFromTeamRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := &entities.MovingPlayerTeam{Executor: entities.User{Role: &entities.Role{}}}

	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	userId, ok := r.Context().Value(cnst.UserIDContextKey).(int64)
	if !ok || userId == 0 {
		err := errors.New(pkgerr.ErrUserIdToken)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}
	request.Executor.ID = userId

	role, ok := r.Context().Value(cnst.RoleNameContextKey).(string)
	if !ok || role == "" {
		err := errors.New(pkgerr.ErrRoleToken)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}
	request.Executor.Role.Name = role

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
