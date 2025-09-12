package league

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

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/error_templates"
	pkgerr "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/errors"
)

func decodeGetListRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	request := &entities.GetLeagueListRequest{}

	cityIDParam := r.URL.Query().Get("cityId")
	if cityIDParam == "" {
		err := errors.New(pkgerr.EmptyParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	cityID, err := strconv.ParseInt(cityIDParam, 10, 64)
	if err != nil {
		err := errors.New(pkgerr.WrongParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	request.CityID = cityID

	return request, nil
}

func decodeCreateRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := &entities.CreateLeagueRequest{}

	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	cityIDParam := r.URL.Query().Get("cityId")
	if cityIDParam == "" {
		err := errors.New(pkgerr.EmptyParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	cityID, err := strconv.ParseInt(cityIDParam, 10, 64)
	if err != nil {
		err = errors.New(pkgerr.WrongParameterError)
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
	request := &entities.UpdateLeagueRequest{}

	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	idParam := chi.URLParam(r, "id")
	if idParam == "" {
		err := errors.New(pkgerr.EmptyParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		err := errors.New(pkgerr.WrongParameterError)
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
	request := &entities.DeleteLeagueRequest{}

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

func decodeRecalcRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := &entities.RecalcLeagueRequest{}

	idParam := chi.URLParam(r, "id")
	if idParam == "" {
		err := errors.New(pkgerr.EmptyParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		err := errors.New(pkgerr.WrongParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	request.ID = id

	return request, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func decodeCreateExtraPointsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := &entities.CreateExtraPointsRequest{}

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

	return request, nil
}

func decodeUpdateExtraPointsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := &entities.UpdateExtraPointsRequest{}

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

	return request, nil
}

func decodeDeleteExtraPointsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := &entities.IdRequest{}

	idParam := chi.URLParam(r, "id")
	if idParam == "" {
		err := errors.New(pkgerr.EmptyParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	extraPointsId, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		err := errors.New(pkgerr.WrongParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	request.Id = extraPointsId
	return request, nil
}

func decodeGetExtraPointsListByTeamAndLeagueIdRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := &entities.TeamLeagueIdRequest{}

	teamIdParam := chi.URLParam(r, "team_id")
	if teamIdParam == "" {
		err := errors.New(pkgerr.EmptyParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}
	teamId, err := strconv.ParseInt(teamIdParam, 10, 64)
	if err != nil {
		err := errors.New(pkgerr.WrongParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	leagueIdParam := chi.URLParam(r, "league_id")
	if leagueIdParam == "" {
		err := errors.New(pkgerr.EmptyParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}
	leagueId, err := strconv.ParseInt(leagueIdParam, 10, 64)
	if err != nil {
		err := errors.New(pkgerr.WrongParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	request.TeamId = teamId
	request.LeagueId = leagueId

	return request, nil
}

func decodeGetExtraPointsByIdRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := &entities.IdRequest{}

	idParam := chi.URLParam(r, "id")
	if idParam == "" {
		err := errors.New(pkgerr.EmptyParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	extraPointsId, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		err := errors.New(pkgerr.WrongParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	request.Id = extraPointsId
	return request, nil
}
