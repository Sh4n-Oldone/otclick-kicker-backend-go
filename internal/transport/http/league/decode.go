package league

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	stderr "errors"

	"github.com/go-chi/chi/v5"
	"github.com/valyala/bytebufferpool"
	"google.golang.org/grpc/codes"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/entity"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/error_templates"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/errors"
)

func decodeGetListRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	request := &entity.GetLeagueListRequest{}

	cityIDParam := r.URL.Query().Get("cityId")
	if cityIDParam == "" {
		err := stderr.New(errors.EmptyParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	cityID, err := strconv.ParseInt(cityIDParam, 10, 64)
	if err != nil {
		err := stderr.New(errors.WrongParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	request.CityID = cityID

	return request, nil
}

func decodeCreateRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := &entity.CreateLeagueRequest{}

	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	cityIDParam := r.URL.Query().Get("cityId")
	if cityIDParam == "" {
		err := stderr.New(errors.EmptyParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	cityID, err := strconv.ParseInt(cityIDParam, 10, 64)
	if err != nil {
		err := stderr.New(errors.WrongParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	_, err = io.Copy(buf, r.Body) // buf.ReadFrom(r.Body)
	if err != nil {
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	err = json.Unmarshal(buf.Bytes(), &request)
	if err != nil {
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	request.CityID = cityID

	return request, nil
}

func decodeUpdateRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := &entity.UpdateLeagueRequest{}

	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	idParam := chi.URLParam(r, "id")
	if idParam == "" {
		err := stderr.New(errors.EmptyParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		err := stderr.New(errors.WrongParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	_, err = io.Copy(buf, r.Body) // buf.ReadFrom(r.Body)
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

func decodeDeleteRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := &entity.DeleteLeagueRequest{}

	idParam := chi.URLParam(r, "id")
	if idParam == "" {
		err := stderr.New(errors.EmptyParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		err := stderr.New(errors.WrongParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	request.ID = id

	return request, nil
}

func decodeRecalcRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := &entity.RecalcLeagueRequest{}

	idParam := chi.URLParam(r, "id")
	if idParam == "" {
		err := stderr.New(errors.EmptyParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		err := stderr.New(errors.WrongParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	request.ID = id

	return request, nil
}
