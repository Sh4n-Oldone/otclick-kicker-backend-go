package player

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

func decodeCreateRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := &entities.CreatePlayerRequest{Creator: entities.User{Role: &entities.Role{}}}
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

	_, err := io.Copy(buf, r.Body) // buf.ReadFrom(r.Body)
	if err != nil {
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	err = json.Unmarshal(buf.Bytes(), &request)
	if err != nil {
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	request.CityIdParam = r.URL.Query().Get("cityId")

	return request, nil
}

func decodeDeleteRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := &entities.DeletePlayerRequest{}
	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	paramID := chi.URLParam(r, "id")
	if paramID == "" {
		err := errors.New(pkgerr.EmptyParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	id, err := strconv.Atoi(paramID)
	if err != nil || id <= 0 {
		err = errors.New(pkgerr.WrongParameterError)
		return nil, error_templates.New(err.Error(), err, http.StatusBadRequest, http.StatusBadRequest)
	}

	request.ID = id

	return request, nil
}

func decodeUpdateRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := &entities.UpdatePlayerRequest{}
	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	paramID := chi.URLParam(r, "id")
	if paramID == "" {
		err := errors.New(pkgerr.EmptyParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	id, err := strconv.ParseInt(paramID, 10, 64)
	if err != nil || id <= 0 {
		err = errors.New(pkgerr.WrongParameterError)
		return nil, error_templates.New(err.Error(), err, http.StatusBadRequest, http.StatusBadRequest)
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

func decodeRecoverRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := &entities.RecoverPlayerRequest{}

	paramID := chi.URLParam(r, "id")
	if paramID == "" {
		err := errors.New(pkgerr.EmptyParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	id, err := strconv.Atoi(paramID)
	if err != nil || id <= 0 {
		err = errors.New(pkgerr.WrongParameterError)
		return nil, error_templates.New(err.Error(), err, http.StatusBadRequest, http.StatusBadRequest)
	}

	request.ID = id

	return request, nil
}

func decodeFindPlayersRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := &entities.FindPlayersRequest{}
	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	paramLeagueID := r.URL.Query().Get("league")
	if paramLeagueID != "" {
		leagueID, err := strconv.Atoi(paramLeagueID)
		if err != nil || leagueID <= 0 {
			err = errors.New(pkgerr.WrongParameterError)
			return nil, error_templates.New(err.Error(), err, http.StatusBadRequest, http.StatusBadRequest)
		}
		request.LeagueID = &leagueID
	}

	paramFindAny := r.URL.Query().Get("findAny")
	request.FindAny = &paramFindAny

	paramGamesPlayedNumber := r.URL.Query().Get("gamesPlayedNumber")
	if paramGamesPlayedNumber != "" {
		gamesPlayedNumber, err := strconv.Atoi(paramGamesPlayedNumber)
		if err != nil || gamesPlayedNumber < 0 {
			err = errors.New(pkgerr.WrongParameterError)
			return nil, error_templates.New(err.Error(), err, http.StatusBadRequest, http.StatusBadRequest)
		}
		request.GamesPlayedNumber = &gamesPlayedNumber
	}

	paramRating := r.URL.Query().Get("rating")
	if paramRating != "" {
		rating, err := strconv.Atoi(paramRating)
		if err != nil || rating < 0 {
			err = errors.New(pkgerr.WrongParameterError)
			return nil, error_templates.New(err.Error(), err, http.StatusBadRequest, http.StatusBadRequest)
		}
		request.Rating = &rating
	}

	paramCityID := r.URL.Query().Get("cityId")
	if paramCityID != "" {
		cityID, err := strconv.Atoi(paramCityID)
		if err != nil || cityID <= 0 {
			err = errors.New(pkgerr.WrongParameterError)
			return nil, error_templates.New(err.Error(), err, http.StatusBadRequest, http.StatusBadRequest)
		}
		request.CityID = &cityID
	}

	var trueValue = true
	var falseValue = false

	paramWithDeleted := r.URL.Query().Get("withDeleted")
	if paramWithDeleted == "true" {
		request.WithDeleted = &trueValue
	}
	if paramWithDeleted == "false" {
		request.WithDeleted = &falseValue
	}

	paramOnlyFree := r.URL.Query().Get("onlyFree")
	if paramOnlyFree == "true" {
		request.OnlyFree = &trueValue
	}
	if paramOnlyFree == "false" {
		request.OnlyFree = &falseValue
	}

	paramKeepSimple := r.URL.Query().Get("keepSimple")
	if paramKeepSimple == "true" {
		request.KeepSimple = &trueValue
	}
	if paramKeepSimple == "false" {
		request.KeepSimple = &falseValue
	}

	return request, nil
}

func decodeGetRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := &entities.GetPlayerRequest{}
	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	paramID := chi.URLParam(r, "id")
	if paramID == "" {
		err := errors.New(pkgerr.EmptyParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	id, err := strconv.Atoi(paramID)
	if err != nil || id <= 0 {
		err = errors.New(pkgerr.WrongParameterError)
		return nil, error_templates.New(err.Error(), err, http.StatusBadRequest, http.StatusBadRequest)
	}

	request.ID = id

	return request, nil
}

func decodeGetByTeamIDRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := &entities.GetPlayersByTeamIDRequest{}
	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	paramID := chi.URLParam(r, "teamId")
	if paramID == "" {
		err := errors.New(pkgerr.EmptyParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	id, err := strconv.Atoi(paramID)
	if err != nil || id <= 0 {
		err = errors.New(pkgerr.WrongParameterError)
		return nil, error_templates.New(err.Error(), err, http.StatusBadRequest, http.StatusBadRequest)
	}

	request.TeamID = id

	return request, nil
}

func decodeGetTournamentPlayerListRequest(_ context.Context, r *http.Request) (interface{}, error) {
	req := &entities.GetTournamentPlayerListRequest{}

	idParam := chi.URLParam(r, "id")
	if idParam == "" {
		err := errors.New(pkgerr.EmptyParameterError + ": id")
		return nil, error_templates.New(err.Error(), err, http.StatusBadRequest, http.StatusBadRequest)
	}

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil || id <= 0 {
		err = errors.New(pkgerr.WrongParameterError + ": id")
		return nil, error_templates.New(err.Error(), err, http.StatusBadRequest, http.StatusBadRequest)
	}

	req.TournamentID = id

	req.WithDeleted = r.URL.Query().Get("withDeleted") == "true"

	return req, nil
}
