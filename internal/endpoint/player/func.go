package player

import (
	"context"
	"net/http"

	"github.com/go-kit/kit/endpoint"
	"google.golang.org/grpc/codes"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/player"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/middleware"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/error_templates"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/errors"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/helpers"
)

func makeCreate(s player.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "player.makeCreate").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[*entities.CreatePlayerRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(errors.FailedCastRequest)
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		err = s.GetValidator().Struct(req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(errors.FailedValidateRequest)
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		id, err := s.Create(ctx, req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("failed to player.Create")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return entities.CreatePlayerResponse{ID: id}, nil
	}
}

func makeDelete(s player.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "player.makeDelete").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[*entities.DeletePlayerRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(errors.FailedCastRequest)
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		err = s.Delete(ctx, req.ID)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("failed to player.Delete")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return entities.OkResponse{
			Message: "OK",
		}, nil
	}
}

func makeRecover(s player.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "player.makeRecover").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[*entities.RecoverPlayerRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(errors.FailedCastRequest)
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		err = s.Recover(ctx, req.ID)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("failed to player.Recover")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return entities.OkResponse{
			Message: "OK",
		}, nil
	}
}

func makeUpdate(s player.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "player.makeUpdate").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[*entities.UpdatePlayerRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(errors.FailedCastRequest)
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		err = s.GetValidator().Struct(req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(errors.FailedValidateRequest)
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		err = s.Update(ctx, *req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("failed to player.Update")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return entities.OkResponse{
			Message: "OK",
		}, nil
	}
}

func makeGet(s player.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "player.makeGet").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[*entities.GetPlayerRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(errors.FailedCastRequest)
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		playerResponse, err := s.Get(ctx, req.ID)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("failed to player.Get")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return playerResponse, nil
	}
}

func makeGetByTeamID(s player.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "player.makeGetByTeamID").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[*entities.GetPlayersByTeamIDRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(errors.FailedCastRequest)
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		playerResponse, err := s.GetByTeamID(ctx, req.TeamID)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("failed to player.Get")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return playerResponse, nil
	}
}

func makeFindPlayers(s player.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "player.makeFindPlayers").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[*entities.FindPlayersRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(errors.FailedCastRequest)
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		//TODO: в случае если метод рабочий, переименовать его в Find, старые удалить из сервиса и из базы
		playersResponse, err := s.FindV2(ctx, *req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("failed to player.Find")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return playersResponse, nil
	}
}

func makeGetTournamentPlayerList(s player.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "player.makeFindPlayers").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[*entities.GetTournamentPlayerListRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(errors.FailedCastRequest)
			return nil, error_templates.WrapErrorEndpoint(
				error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest), reqID)
		}

		tournamentPlayers, err := s.GetTournamentPlayerList(ctx, req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("failed to player.GetTournamentPlayerList")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return struct {
			Players []entities.TournamentPlayerItem `json:"players"`
		}{
			Players: tournamentPlayers,
		}, nil
	}
}
