package game

import (
	"context"
	"github.com/go-kit/kit/endpoint"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/game"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/middleware"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/error_templates"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/helpers"
)

func makeCreate(s game.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "game.makeCreate").Logger()

		req, err := helpers.CastRequest[entities.CreateGameRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed to cast request")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		resp, err := s.Create(ctx, req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed to game.Create")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return resp, nil
	}
}

func makeDelete(s game.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "game.makeDelete").Logger()

		gameId, err := helpers.CastRequest[int](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed to cast request")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		err = s.Delete(ctx, gameId)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed to game.Delete")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return entities.OkResponse{
			Message: "OK",
		}, nil
	}
}

func makeGet(s game.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "game.makeGet").Logger()

		gameId, err := helpers.CastRequest[int](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed to cast request")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		gameResp, err := s.Get(ctx, gameId)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed to game.Get")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return gameResp, nil
	}
}

func makeUpdate(s game.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "game.makeUpdate").Logger()

		req, err := helpers.CastRequest[entities.UpdateGameRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed to cast request")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		err = s.Update(ctx, req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed to game.Update")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return entities.OkResponse{
			Message: "OK",
		}, nil
	}
}

func makeFind(s game.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "game.makeFind").Logger()

		req, err := helpers.CastRequest[entities.FindGameRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed to cast request")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		gameResp, err := s.Find(ctx, req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed to game.Find")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return gameResp, nil
	}
}

func makeUpdateFutureGame(s game.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "game.makeUpdateFutureGame").Logger()

		req, err := helpers.CastRequest[entities.UpdateFutureGameRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed to cast request")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		err = s.UpdateFutureGame(ctx, req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed to game.UpdateFutureGame")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return entities.OkResponse{
			Message: "OK",
		}, nil
	}
}

func makeGetYears(s game.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		//reqID, ctx := middleware.GetRequestID(ctx)

		resp, err := s.GetGamesYears(ctx)
		if err != nil {
			s.GetLogger().Error().Err(err).Msg("Failed to game.GetGameYears")
			return nil, err
		}

		return resp, nil
	}
}
