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

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// Create ExtraPoints
func decodeCreateExtraPointsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := &entity.CreateExtraPointsRequest{}

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

// Update ExtraPoints
func decodeUpdateExtraPointsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := &entity.UpdateExtraPointsRequest{}

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

// Delete ExtraPoints
func decodeDeleteExtraPointsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := &entity.IdRequest{}

	idParam := chi.URLParam(r, "id")
	if idParam == "" {
		err := stderr.New(errors.EmptyParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	extraPointsId, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		err := stderr.New(errors.WrongParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	request.Id = extraPointsId
	return request, nil
}

// Get ExtraPointsListByTeamAndLeagueId
func decodeGetExtraPointsListByTeamAndLeagueIdRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := &entity.TeamLeagueIdRequest{}

	teamIdParam := chi.URLParam(r, "team_id")
	if teamIdParam == "" {
		err := stderr.New(errors.EmptyParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}
	teamId, err := strconv.ParseInt(teamIdParam, 10, 64)
	if err != nil {
		err := stderr.New(errors.WrongParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	leagueIdParam := chi.URLParam(r, "league_id")
	if leagueIdParam == "" {
		err := stderr.New(errors.EmptyParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}
	leagueId, err := strconv.ParseInt(leagueIdParam, 10, 64)
	if err != nil {
		err := stderr.New(errors.WrongParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	request.TeamId = teamId
	request.LeagueId = leagueId

	return request, nil
}

// Get ExtraPointsById
func decodeGetExtraPointsByIdRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := &entity.IdRequest{}

	idParam := chi.URLParam(r, "id")
	if idParam == "" {
		err := stderr.New(errors.EmptyParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	extraPointsId, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		err := stderr.New(errors.WrongParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	request.Id = extraPointsId
	return request, nil
}
