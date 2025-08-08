package player

import (
	"context"

	"github.com/go-kit/kit/endpoint"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/player"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/middleware"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/error_templates"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/helpers"
)

func makeCreate(s player.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "player.makeCreate").Logger()

		req, err := helpers.CastRequest[*entities.CreatePlayerRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed to cast request")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		err = s.GetValidator().Struct(req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed validate request")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		id, err := s.Create(ctx, req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed to player.Create")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return entities.CreatePlayerResponse{ID: id}, nil
	}
}

func makeDelete(s player.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "player.makeDelete").Logger()

		req, err := helpers.CastRequest[*entities.DeletePlayerRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed to cast request")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		err = s.Delete(ctx, req.ID)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed to player.Delete")
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
		serviceLogger := s.GetLogger().With().Str("Source", "player.makeRecover").Logger()

		req, err := helpers.CastRequest[*entities.RecoverPlayerRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed to cast request")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		err = s.Recover(ctx, req.ID)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed to player.Recover")
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
		serviceLogger := s.GetLogger().With().Str("Source", "player.makeUpdate").Logger()

		req, err := helpers.CastRequest[*entities.UpdatePlayerRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed to cast request")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		err = s.Update(ctx, *req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed to player.Update")
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
		serviceLogger := s.GetLogger().With().Str("Source", "player.makeGet").Logger()

		req, err := helpers.CastRequest[*entities.GetPlayerRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed to cast request")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		playerResponse, err := s.Get(ctx, req.ID)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed to player.Get")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return playerResponse, nil
	}
}

func makeGetByTeamID(s player.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "player.makeGetByTeamID").Logger()

		req, err := helpers.CastRequest[*entities.GetPlayersByTeamIDRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed to cast request")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		playerResponse, err := s.GetByTeamID(ctx, req.TeamID)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed to player.Get")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return playerResponse, nil
	}
}

func makeFindPlayers(s player.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "player.makeFindPlayers").Logger()

		req, err := helpers.CastRequest[*entities.FindPlayersRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed to cast request")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		playersResponse, err := s.Find(ctx, *req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed to player.Find")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return playersResponse, nil
	}
}
