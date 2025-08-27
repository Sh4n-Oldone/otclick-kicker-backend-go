package place

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

func decodeGetListRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := &entities.GetPlaceListRequest{}

	attrBarID := r.URL.Query().Get("barId")
	if attrBarID != "" {
		barID, err := strconv.ParseInt(attrBarID, 10, 64)
		if err != nil || barID < 1 {
			err = errors.New(pkgerr.WrongParameterError)
			return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}
		request.BarID = &barID
	}
	attrTableID := r.URL.Query().Get("tableId")
	if attrTableID != "" {
		tableID, err := strconv.ParseInt(attrTableID, 10, 64)
		if err != nil || tableID < 1 {
			err = errors.New(pkgerr.WrongParameterError)
			return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}
		request.TableID = &tableID
	}
	attrCityID := r.URL.Query().Get("cityId")
	if attrCityID != "" {
		cityID, err := strconv.ParseInt(attrCityID, 10, 64)
		if err != nil || cityID < 1 {
			err = errors.New(pkgerr.WrongParameterError)
			return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}
		request.CityID = &cityID
	}

	if request.CityID != nil && request.BarID != nil {
		err := errors.New("одновременный поиск по барам и городам невозможен")
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	withDeleted := r.URL.Query().Get("withDeleted")
	if withDeleted == "true" {
		request.WithDeleted = true
	}

	return request, nil
}

func decodeCreateRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := &entities.CreatePlaceRequest{Creator: entities.User{Role: &entities.Role{}}}

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

	_, err := io.Copy(buf, r.Body) // buf.ReadFrom(r.Body)
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
	request := &entities.UpdatePlaceRequest{Executor: entities.User{Role: &entities.Role{}}}

	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	userId, ok := r.Context().Value(constant.UserIDContextKey).(int64)
	if !ok || userId == 0 {
		err := errors.New(pkgerr.ErrUserIdToken)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}
	request.Executor.ID = userId

	role, ok := r.Context().Value(constant.RoleNameContextKey).(string)
	if !ok || role == "" {
		err := errors.New(pkgerr.ErrRoleToken)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}
	request.Executor.Role.Name = role

	_, err := io.Copy(buf, r.Body) // buf.ReadFrom(r.Body)
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
	request := &entities.DeletePlaceRequest{Executor: entities.User{Role: &entities.Role{}}}

	attrID := chi.URLParam(r, "id")
	if attrID == "" {
		err := errors.New(pkgerr.EmptyParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	id, err := strconv.ParseInt(attrID, 10, 64)
	if err != nil {
		err = errors.New(pkgerr.WrongParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	userId, ok := r.Context().Value(constant.UserIDContextKey).(int64)
	if !ok || userId == 0 {
		err := errors.New(pkgerr.ErrUserIdToken)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}
	request.Executor.ID = userId

	role, ok := r.Context().Value(constant.RoleNameContextKey).(string)
	if !ok || role == "" {
		err := errors.New(pkgerr.ErrRoleToken)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}
	request.Executor.Role.Name = role

	request.ID = id

	return request, nil
}
