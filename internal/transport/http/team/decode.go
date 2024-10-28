package team

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

////////////////////////////////////////////////////////////////////////////////////
// GetTeam {id}
// GetTeams
// GetTeamsByCity {city_id}
// GetTeamsByLeague {league_id}
// GetTeamVsTeamTable {city_id}

// Create
// Update
// Delete {id}

// AddPlayerIntoTeam
// RemovePlayerFromTeam
////////////////////////////////////////////////////////////////////////////////////

// GetTeam{id}
func decodeGetTeamRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := &entity.GetTeamRequest{}

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

// // GetTeams
func decodeGetTeamsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := &entity.GetTeamsRequest{}

	var onlyFree bool = false
	onlyFreeParam := r.URL.Query().Get("onlyFree")
	if onlyFreeParam == "true" {
		onlyFree = true
	}

	cityIdParam := r.URL.Query().Get("cityId")
	if cityIdParam != "" {
		cityID, err := strconv.ParseInt(cityIdParam, 10, 64)
		if err != nil {
			err := stderr.New(errors.WrongParameterError)
			return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}
		request.CityId = cityID
	}

	request.OnlyFree = onlyFree
	return request, nil
}

// // GetTeamsByCity{city_id}
func decodeGetTeamsByCityRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := &entity.GetTeamsByCityRequest{}

	idParam := chi.URLParam(r, "city_id")
	if idParam == "" {
		err := stderr.New(errors.EmptyParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	var onlyFree bool = false
	onlyFreeParam := r.URL.Query().Get("onlyFree")
	if onlyFreeParam == "true" {
		onlyFree = true
	}

	cityID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		err := stderr.New(errors.WrongParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	request.OnlyFree = onlyFree
	request.CityID = cityID

	return request, nil
}

// // GetTeamsByLeague{league_id}
func decodeGetTeamsByLeagueRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := &entity.GetTeamsByLeagueRequest{}

	idParam := chi.URLParam(r, "league_id")
	if idParam == "" {
		err := stderr.New(errors.EmptyParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	var onlyFree bool = false
	onlyFreeParam := r.URL.Query().Get("onlyFree")
	if onlyFreeParam == "true" {
		onlyFree = true
	}

	leagueID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		err := stderr.New(errors.WrongParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	request.OnlyFree = onlyFree
	request.LeagueID = leagueID

	return request, nil
}

func decodeGetTeamVsTeamTableRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := &entity.GetTeamVsTeamTableRequest{}

	cityIdParam := r.URL.Query().Get("cityId")
	if cityIdParam == "" {
		err := stderr.New(errors.EmptyParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}
	cityID, err := strconv.ParseInt(cityIdParam, 10, 64)
	if err != nil {
		err := stderr.New(errors.WrongParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	yearParam := r.URL.Query().Get("year")
	if yearParam == "" {
		err := stderr.New(errors.EmptyParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}
	year, err := strconv.ParseInt(yearParam, 10, 64)
	if err != nil {
		err := stderr.New(errors.WrongParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	request.CityID = cityID
	request.Year = year

	return request, nil
}

////////////////////////////////////////////////////////////////////////////////////

// Create
func decodeCreateRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := &entity.CreateTeamRequest{}

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

// Update
func decodeUpdateRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := &entity.UpdateTeamRequest{}

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

// Delete {id}
func decodeDeleteRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := &entity.DeleteTeamRequest{}

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

////////////////////////////////////////////////////////////////////////////////////

// AddPlayerIntoTeam
func decodeAddPlayerIntoTeamRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := &entity.PlayerTeam{}

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

// // RemovePlayerFromTeam
func decodeRemovePlayerFromTeamRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := &entity.PlayerTeam{}

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
