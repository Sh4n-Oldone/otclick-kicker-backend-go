package game

import (
	"context"
	"encoding/json"
	stderr "errors"
	"github.com/go-chi/chi/v5"
	"github.com/valyala/bytebufferpool"
	"google.golang.org/grpc/codes"
	"io"
	"net/http"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/error_templates"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/errors"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/helpers"
	"strconv"
)

var validate = helpers.NewCustomValidator()

func decodeCreateRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := entities.CreateGameRequest{}
	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	cityIdParam := r.URL.Query().Get("cityId")
	if cityIdParam == "" {
		err := stderr.New(errors.EmptyParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	cityId, err := strconv.Atoi(cityIdParam)
	if err != nil {
		err = stderr.New(errors.WrongParameterError)
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

	request.CityID = cityId

	err = validate.Struct(request)
	if err != nil {
		err = stderr.New(errors.ValidationErr + ": " + err.Error())
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	return request, nil
}

func decodeIdParamRequest(_ context.Context, r *http.Request) (interface{}, error) {
	gameIdParam := chi.URLParam(r, "id")
	if gameIdParam == "" {
		err := stderr.New(errors.EmptyParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	gameId, err := strconv.Atoi(gameIdParam)
	if err != nil {
		err = stderr.New(errors.WrongParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	if gameId < 1 {
		err = stderr.New(errors.WrongParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	return gameId, nil
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
		err = stderr.New(errors.ValidationErr + ": " + err.Error())
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	return request, nil
}

func decodeFindRequest(_ context.Context, r *http.Request) (interface{}, error) {
	request := entities.FindGameRequest{}
	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	team1IdParam := r.URL.Query().Get("team1Id")
	if team1IdParam == "" {
		err := stderr.New(errors.EmptyParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	team1Id, err := strconv.Atoi(team1IdParam)
	if err != nil || team1Id < 1 {
		err = stderr.New(errors.WrongParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	team2IdParam := r.URL.Query().Get("team2Id")
	if team2IdParam == "" {
		err = stderr.New(errors.EmptyParameterError)
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	team2Id, err := strconv.Atoi(team2IdParam)
	if err != nil || team2Id < 1 {
		err = stderr.New(errors.WrongParameterError)
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
		err = stderr.New(errors.ValidationErr + ": " + err.Error())
		return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	return request, nil
}
