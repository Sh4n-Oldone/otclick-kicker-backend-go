package game

import (
	"context"
	stderr "errors"
	"github.com/go-kit/kit/endpoint"
	"google.golang.org/grpc/codes"
	"net/http"

	cnst "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/constant"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/entity"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/game"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/middleware"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/error_templates"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/errors"
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

		req, err := helpers.CastRequest[entity.UpdateFutureGameRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed to cast request")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		// Captain-Flow
		// limitations for the role 'captain'
		// if teamID of captain not equal team1 or team2 from request then user unauthorized error
		role := ctx.Value(cnst.RoleNameContextKey)
		teamID := ctx.Value(cnst.TeamIDContextKey)
		rTeam1ID := int64(req.Team1ID)
		if role == cnst.CaptainRole && teamID != rTeam1ID {
			serviceLogger.Error().Err(err).Msg("Failed to captain request")
			err = error_templates.New(errors.WrongUserRole, stderr.New(errors.WrongUserRole), codes.Unauthenticated, http.StatusUnauthorized)
			return nil, err
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

func makeGetComingGames(s game.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "game.makeGetComingGames").Logger()

		resp, err := s.GetComingGames(ctx)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed to game.UpdateFutureGame")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return resp, nil
	}
}

func makeGetFutureGames(s game.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "game.makeGetFutureGames").Logger()

		cityID, err := helpers.CastRequest[int](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed to cast request")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		resp, err := s.GetFutureGames(ctx, cityID)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed to game.GetFutureGames")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return resp, nil
	}
}

func makeCreateFutureGame(s game.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "game.makeCreateFutureGame").Logger()

		req, err := helpers.CastRequest[entity.CreateFutureGameRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed to cast request")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		// Captain-Flow
		// limitations for the role 'captain'
		// if teamID of captain not equal team1 or team2 from request then user unauthorized error
		role := ctx.Value(cnst.RoleNameContextKey)
		teamID := ctx.Value(cnst.TeamIDContextKey)
		rTeam1ID := int64(req.Team1ID)
		if role == cnst.CaptainRole && teamID != rTeam1ID {
			serviceLogger.Error().Err(err).Msg("Failed to captain request")
			err = error_templates.New(errors.WrongUserRole, stderr.New(errors.WrongUserRole), codes.Unauthenticated, http.StatusUnauthorized)
			return nil, err
		}

		resp, err := s.CreateFutureGame(ctx, req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed to game.CreateFutureGame")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return resp, nil
	}
}

func makeGetTeamGames(s game.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "game.makeGetTeamGames").Logger()

		teamID, err := helpers.CastRequest[int](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed to cast request")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		resp, err := s.GetTeamGames(ctx, teamID)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed to game.GetTeamGames")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return resp, nil
	}
}
