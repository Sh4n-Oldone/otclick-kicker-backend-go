package game

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-kit/kit/endpoint"
	"google.golang.org/grpc/codes"

	cnst "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/constant"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/game"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/middleware"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/error_templates"
	pkgerr "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/errors"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/helpers"
)

func makeCreate(s game.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "game.makeCreate").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[entities.CreateGameRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(pkgerr.FailedCastRequest)
			return nil, error_templates.WrapErrorEndpoint(
				error_templates.New(pkgerr.FailedCastRequest, err, codes.InvalidArgument, http.StatusBadRequest), reqID)
		}

		// Checking matches only for filled
		if len(req.Matches) > 0 {
			for _, m := range req.Matches {
				if m.Team1ID != req.Team1ID || m.Team2ID != req.Team2ID {
					err = error_templates.New(pkgerr.ErrDifferentTeams, errors.New(pkgerr.ErrDifferentTeams), codes.InvalidArgument, http.StatusBadRequest)
					serviceLogger.Error().Err(err).Msg("failed to game.makeCreate")
					return nil, error_templates.WrapErrorEndpoint(err, reqID)
				}
			}
		}

		err = s.GetValidator().Struct(req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(pkgerr.FailedValidateRequest)
			return nil, error_templates.WrapErrorEndpoint(
				error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest), reqID)
		}

		resp, err := s.Create(ctx, req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("failed to game.Create")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return resp, nil
	}
}

func makeDelete(s game.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "game.makeDelete").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[entities.DeleteGameRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(pkgerr.FailedCastRequest)
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		err = s.GetValidator().Struct(req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(pkgerr.FailedValidateRequest)
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		err = s.Delete(ctx, &req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("failed to game.Delete")
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
		serviceLogger := s.GetLogger().With().Str("Source", "game.makeGet").Str("request_id", reqID).Logger()

		gameId, err := helpers.CastRequest[int](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(pkgerr.FailedCastRequest)
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		gameResp, err := s.Get(ctx, gameId)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("failed to game.Get")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return gameResp, nil
	}
}

func makeUpdate(s game.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "game.makeUpdate").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[entities.UpdateGameRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(pkgerr.FailedCastRequest)
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		err = s.GetValidator().Struct(req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(pkgerr.FailedValidateRequest)
			return nil, error_templates.WrapErrorEndpoint(
				error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest), reqID)
		}

		err = s.Update(ctx, req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("failed to game.Update")
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
		serviceLogger := s.GetLogger().With().Str("Source", "game.makeFind").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[entities.FindGameRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(pkgerr.FailedCastRequest)
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		gameResp, err := s.Find(ctx, req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("failed to game.Find")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return gameResp, nil
	}
}

func makeUpdateFutureGame(s game.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "game.makeUpdateFutureGame").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[entities.UpdateFutureGameRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(pkgerr.FailedCastRequest)
			return nil, error_templates.WrapErrorEndpoint(
				error_templates.New(pkgerr.FailedCastRequest, err, codes.InvalidArgument, http.StatusBadRequest), reqID)
		}

		err = s.GetValidator().Struct(req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(pkgerr.FailedValidateRequest)
			return nil, error_templates.WrapErrorEndpoint(
				error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest), reqID)
		}

		err = s.UpdateFutureGame(ctx, req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("failed to game.UpdateFutureGame")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return entities.OkResponse{
			Message: "OK",
		}, nil
	}
}

func makeGetYears(s game.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)

		resp, err := s.GetGamesYears(ctx)
		if err != nil {
			s.GetLogger().Error().Err(err).Msg("failed to game.GetGameYears")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return resp, nil
	}
}

func makeGetGameList(s game.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "game.makeGetGameList").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[entities.GetGameListRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(pkgerr.FailedCastRequest)
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		err = helpers.ValidateGetGameList(req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(pkgerr.FailedValidateRequest)
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		resp, err := s.GetGameList(ctx, req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("failed to game.GetGameList")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return resp, nil
	}
}

func makeGetComingGames(s game.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "game.makeGetComingGames").Str("request_id", reqID).Logger()

		resp, err := s.GetComingGames(ctx)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("failed to game.GetComingGames")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return resp, nil
	}
}

func makeGetFutureGames(s game.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "game.makeGetFutureGames").Str("request_id", reqID).Logger()

		cityID, err := helpers.CastRequest[int](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(pkgerr.FailedCastRequest)
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		resp, err := s.GetFutureGames(ctx, cityID)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("failed to game.GetFutureGames")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return resp, nil
	}
}

func makeCreateFutureGame(s game.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "game.makeCreateFutureGame").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[entities.CreateFutureGameRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(pkgerr.FailedCastRequest)
			return nil, error_templates.WrapErrorEndpoint(
				error_templates.New(pkgerr.FailedCastRequest, err, codes.InvalidArgument, http.StatusBadRequest), reqID)
		}

		err = s.GetValidator().Struct(req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(pkgerr.FailedValidateRequest)
			return nil, error_templates.WrapErrorEndpoint(
				error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest), reqID)
		}

		resp, err := s.CreateFutureGame(ctx, req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("failed to game.CreateFutureGame")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return resp, nil
	}
}

func makeGetTeamGames(s game.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "game.makeGetTeamGames").Str("request_id", reqID).Logger()

		teamID, err := helpers.CastRequest[int](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(pkgerr.FailedCastRequest)
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		// Captain-Flow
		// limitations for the role 'captain'
		// if teamID of captain not equal team1 or team2 from request then user unauthorized error
		role := ctx.Value(cnst.RoleNameContextKey)
		ctxTeamID := ctx.Value(cnst.TeamIDContextKey)
		if role == cnst.CaptainRole && ctxTeamID != int64(teamID) {
			serviceLogger.Error().Err(err).Msg("failed to captain request")
			err = error_templates.New(pkgerr.WrongUserRole, errors.New(pkgerr.WrongUserRole), codes.Unauthenticated, http.StatusUnauthorized)
			return nil, err
		}

		resp, err := s.GetTeamGames(ctx, teamID)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("failed to game.GetTeamGames")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return resp, nil
	}
}

func makeDeleteFutureGame(s game.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "game.makeDeleteFutureGame").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[entities.DeleteFutureGameRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(pkgerr.FailedCastRequest)
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		err = s.GetValidator().Struct(req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(pkgerr.FailedValidateRequest)
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		err = s.DeleteFutureGame(ctx, req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("failed to game.DeleteFutureGame")
			return nil, error_templates.WrapErrorEndpoint(
				error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest), reqID)
		}

		return entities.OkResponse{
			Message: "OK",
		}, nil
	}
}

func makeCreateFutureTournamentGame(s game.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeCreateFutureTournamentGame").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[*entities.CreateFutureTournamentGameRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(pkgerr.FailedCastRequest)
			return nil, error_templates.WrapErrorEndpoint(
				error_templates.New(pkgerr.FailedCastRequest, err, codes.InvalidArgument, http.StatusBadRequest), reqID)
		}

		err = s.GetValidator().Struct(req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(pkgerr.FailedValidateRequest)
			return nil, error_templates.WrapErrorEndpoint(
				error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest), reqID)
		}

		id, err := s.CreateFutureTournamentGame(ctx, req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("failed to game.CreateFutureTournamentGame")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return struct {
			Id int64 `json:"id"`
		}{
			Id: id,
		}, nil
	}
}

func makeUpdateFutureTournamentGame(s game.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeUpdateFutureTournamentGame").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[*entities.UpdateFutureTournamentGameRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(pkgerr.FailedCastRequest)
			return nil, error_templates.WrapErrorEndpoint(
				error_templates.New(pkgerr.FailedCastRequest, err, codes.InvalidArgument, http.StatusBadRequest), reqID)
		}

		err = s.GetValidator().Struct(req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(pkgerr.FailedValidateRequest)
			return nil, error_templates.WrapErrorEndpoint(
				error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest), reqID)
		}

		err = s.UpdateFutureTournamentGame(ctx, req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("failed to game.UpdateFutureTournamentGame")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return struct {
			Message int64 `json:"updatedGameId"`
		}{
			Message: req.GameID,
		}, nil
	}
}

func makeDeleteFutureTournamentGame(s game.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeDeleteFutureTournamentGame").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[*entities.DeleteFutureTournamentGameRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(pkgerr.FailedCastRequest)
			return nil, error_templates.WrapErrorEndpoint(
				error_templates.New(pkgerr.FailedCastRequest, err, codes.InvalidArgument, http.StatusBadRequest), reqID)
		}

		err = s.GetValidator().Struct(req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(pkgerr.FailedValidateRequest)
			return nil, error_templates.WrapErrorEndpoint(
				error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest), reqID)
		}

		err = s.DeleteFutureTournamentGame(ctx, req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("failed to game.DeleteFutureTournamentGame")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return struct {
			Message int64 `json:"deletedGameId"`
		}{
			Message: req.GameID,
		}, nil
	}
}

func makeCreatePlayedTournamentGame(s game.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeCreatePlayedTournamentGame").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[*entities.CreatePlayedTournamentGameRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(pkgerr.FailedCastRequest)
			return nil, error_templates.WrapErrorEndpoint(
				error_templates.New(pkgerr.FailedCastRequest, err, codes.InvalidArgument, http.StatusBadRequest), reqID)
		}

		err = s.GetValidator().Struct(req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(pkgerr.FailedValidateRequest)
			return nil, error_templates.WrapErrorEndpoint(
				error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest), reqID)
		}

		resp, err := s.CreatePlayedTournamentGame(ctx, req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("failed to game.CreatePlayedTournamentGame")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return resp, nil
	}
}

func makeUpdatePlayedTournamentGame(s game.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeUpdatePlayedTournamentGame").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[*entities.UpdatePlayedTournamentGameRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(pkgerr.FailedCastRequest)
			return nil, error_templates.WrapErrorEndpoint(
				error_templates.New(pkgerr.FailedCastRequest, err, codes.InvalidArgument, http.StatusBadRequest), reqID)
		}

		err = s.GetValidator().Struct(req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(pkgerr.FailedValidateRequest)
			return nil, error_templates.WrapErrorEndpoint(
				error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest), reqID)
		}

		resp, err := s.UpdatePlayedTournamentGame(ctx, req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("failed to game.UpdatePlayedTournamentGame")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return resp, nil
	}
}

func makeDeletePlayedTournamentGame(s game.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeDeletePlayedTournamentGame").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[*entities.DeletePlayedTournamentGameRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(pkgerr.FailedCastRequest)
			return nil, error_templates.WrapErrorEndpoint(
				error_templates.New(pkgerr.FailedCastRequest, err, codes.InvalidArgument, http.StatusBadRequest), reqID)
		}

		err = s.GetValidator().Struct(req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(pkgerr.FailedValidateRequest)
			return nil, error_templates.WrapErrorEndpoint(
				error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest), reqID)
		}

		err = s.DeletePlayedTournamentGame(ctx, req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("failed to game.DeletePlayedTournamentGame")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return struct {
			Message int64 `json:"deletedGameId"`
		}{
			Message: req.GameID,
		}, nil
	}
}

func makeGetTournamentGameList(s game.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeGetTournamentGameList").Str("request_id", reqID).Logger()

		id, err := helpers.CastRequest[int64](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(pkgerr.FailedCastRequest)
			return nil, error_templates.WrapErrorEndpoint(
				error_templates.New(pkgerr.FailedCastRequest, err, codes.InvalidArgument, http.StatusBadRequest), reqID)
		}

		games, err := s.GetTournamentGameList(ctx, id)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("failed to game.GetTournamentGameList")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return struct {
			Id    int64                         `json:"tournamentId"`
			Games []entities.FullTournamentGame `json:"games"`
		}{
			Id:    id,
			Games: games,
		}, nil
	}
}

func makeGetFutureTournamentGameList(s game.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeGetFutureTournamentGameList").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[*entities.GetTournamentGameList](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(pkgerr.FailedCastRequest)
			return nil, error_templates.WrapErrorEndpoint(
				error_templates.New(pkgerr.FailedCastRequest, err, codes.InvalidArgument, http.StatusBadRequest), reqID)
		}

		games, err := s.GetFutureTournamentGameList(ctx, req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("failed to game.GetFutureTournamentGameList")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return struct {
			Games []entities.TournamentGame `json:"games"`
		}{
			Games: games,
		}, nil
	}
}

func makeGetPlayedTournamentGameList(s game.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeGetPlayedTournamentGameList").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[*entities.GetTournamentGameList](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(pkgerr.FailedCastRequest)
			return nil, error_templates.WrapErrorEndpoint(
				error_templates.New(pkgerr.FailedCastRequest, err, codes.InvalidArgument, http.StatusBadRequest), reqID)
		}

		games, err := s.GetPlayedTournamentGameList(ctx, req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("failed to game.GetPlayedTournamentGameList")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return struct {
			Games []entities.FullTournamentGame `json:"games"`
		}{
			Games: games,
		}, nil
	}
}
