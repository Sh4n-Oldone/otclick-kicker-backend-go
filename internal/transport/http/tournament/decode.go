package tournament

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

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/constant"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/error_templates"
	pkgerr "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/errors"
)

func decodeGetTournamentTypeListRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := &entities.GetTournamentTypeListRequest{}

	request.WithDeleted = r.URL.Query().Get("withDeleted") == "true"

	return request, nil
}

func decodeCreateRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := &entities.CreateTournamentRequest{Creator: entities.User{Role: &entities.Role{}}}
	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	userId, ok := r.Context().Value(constant.UserIDContextKey).(int64)
	if !ok || userId == 0 {
		err := errors.New(pkgerr.ErrUserIdToken)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}
	request.Creator.ID = userId

	role, ok := r.Context().Value(constant.RoleNameContextKey).(string)
	if !ok || role == "" {
		err := errors.New(pkgerr.ErrRoleToken)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}
	request.Creator.Role.Name = role

	request.CityIDParam = r.URL.Query().Get("cityId")

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
	request := &entities.UpdateTournamentRequest{Executor: entities.User{Role: &entities.Role{}}}
	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	idParam := chi.URLParam(r, "id")
	if idParam == "" {
		err := errors.New(pkgerr.EmptyParameterError + ": id")
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		err = errors.New(pkgerr.WrongParameterError + ": id")
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	request.ID = id

	userId, ok := r.Context().Value(constant.UserIDContextKey).(int64)
	if !ok || userId == 0 {
		err := errors.New(pkgerr.ErrUserIdToken)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}
	request.Executor.ID = userId

	role, ok := r.Context().Value(constant.RoleNameContextKey).(string)
	if !ok || role == "" {
		err = errors.New(pkgerr.ErrRoleToken)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}
	request.Executor.Role.Name = role

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

func decodeDeleteRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := &entities.DeleteTournamentRequest{}
	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	idParam := chi.URLParam(r, "id")
	if idParam == "" {
		err := errors.New(pkgerr.EmptyParameterError)
		return nil, error_templates.New(err.Error(), err, http.StatusBadRequest, http.StatusBadRequest)
	}

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil || id == 0 {
		err = errors.New(pkgerr.WrongParameterError)
		return nil, error_templates.New(err.Error(), err, http.StatusBadRequest, http.StatusBadRequest)
	}

	request.ID = id

	return request, nil
}

func decodeFinishStageRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := &entities.FinishStageRequest{Finisher: entities.User{Role: &entities.Role{}}}
	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	idParam := chi.URLParam(r, "id")
	if idParam == "" {
		err := errors.New(pkgerr.EmptyParameterError + ": id")
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		err = errors.New(pkgerr.WrongParameterError + ": id")
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	request.ID = id

	userId, ok := r.Context().Value(constant.UserIDContextKey).(int64)
	if !ok || userId == 0 {
		err := errors.New(pkgerr.ErrUserIdToken)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}
	request.Finisher.ID = userId

	role, ok := r.Context().Value(constant.RoleNameContextKey).(string)
	if !ok || role == "" {
		err := errors.New(pkgerr.ErrRoleToken)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}
	request.Finisher.Role.Name = role

	return request, nil
}

func decodeIdRequest(_ context.Context, r *http.Request) (interface{}, error) {
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
