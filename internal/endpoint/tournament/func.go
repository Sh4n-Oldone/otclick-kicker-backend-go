package tournament

import (
	"context"
	"net/http"

	"github.com/go-kit/kit/endpoint"
	"google.golang.org/grpc/codes"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/tournament"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/middleware"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/error_templates"
	pkgerr "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/errors"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/helpers"
)

func makeGetTournamentTypeList(s tournament.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeGetTeam").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[*entities.GetTournamentTypeListRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed to cast request")
			return nil, error_templates.WrapErrorEndpoint(
				error_templates.New(pkgerr.FailedCastRequest, err, codes.InvalidArgument, http.StatusBadRequest), reqID)
		}

		list, err := s.GetTournamentTypeList(ctx, req.WithDeleted)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed GetTournamentTypeList")
			return nil, error_templates.WrapErrorEndpoint(
				error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest), reqID)
		}

		return struct {
			TournamentTypeList []entities.TournamentType `json:"tournamentTypeList"`
		}{
			TournamentTypeList: list,
		}, nil
	}
}

func makeCreate(s tournament.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeCreate").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[*entities.CreateTournamentRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed to cast request")
			return nil, error_templates.WrapErrorEndpoint(
				error_templates.New(pkgerr.FailedCastRequest, err, codes.InvalidArgument, http.StatusBadRequest), reqID)
		}

		err = s.GetValidator().Struct(req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed validation in makeCreate")
			return nil, error_templates.WrapErrorEndpoint(
				error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest), reqID)
		}

		id, err := s.Create(ctx, req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed Create in team.makeCreate")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return struct {
			Id int64 `json:"id"`
		}{
			Id: id,
		}, nil
	}
}

func makeUpdate(s tournament.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeUpdate").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[*entities.UpdateTournamentRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed to cast request")
			return nil, error_templates.WrapErrorEndpoint(
				error_templates.New(pkgerr.FailedCastRequest, err, codes.InvalidArgument, http.StatusBadRequest), reqID)
		}

		err = s.GetValidator().Struct(req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed validation in makeUpdate")
			return nil, error_templates.WrapErrorEndpoint(
				error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest), reqID)
		}

		err = s.Update(ctx, req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed Create in team.makeUpdate")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return struct {
			Message int64 `json:"updatedTournamentId"`
		}{
			Message: req.ID,
		}, nil
	}
}

func makeDelete(s tournament.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeDelete").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[*entities.DeleteTournamentRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed to cast request")
			return nil, error_templates.WrapErrorEndpoint(
				error_templates.New(pkgerr.FailedCastRequest, err, codes.InvalidArgument, http.StatusBadRequest), reqID)
		}

		err = s.GetValidator().Struct(req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed validation in makeDelete")
			return nil, error_templates.WrapErrorEndpoint(
				error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest), reqID)
		}

		err = s.Delete(ctx, req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed Delete")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return struct {
			Message int64 `json:"deletedTournamentId"`
		}{
			Message: req.ID,
		}, nil
	}
}

func makeFinishStage(s tournament.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeFinishStage").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[*entities.FinishStageRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed to cast request")
			return nil, error_templates.WrapErrorEndpoint(
				error_templates.New(pkgerr.FailedCastRequest, err, codes.InvalidArgument, http.StatusBadRequest), reqID)
		}

		err = s.GetValidator().Struct(req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed validation in makeFinishStage")
			return nil, error_templates.WrapErrorEndpoint(
				error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest), reqID)
		}

		err = s.FinishStage(ctx, req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed s.FinishStage")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return struct {
			Id int64 `json:"finishedStageId"`
		}{
			Id: req.ID,
		}, nil
	}
}

func makeGetTournamentStageList(s tournament.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeGetTournamentStageList").Str("request_id", reqID).Logger()

		id, err := helpers.CastRequest[int64](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed to cast request")
			return nil, error_templates.WrapErrorEndpoint(
				error_templates.New(pkgerr.FailedCastRequest, err, codes.InvalidArgument, http.StatusBadRequest), reqID)
		}

		stages, err := s.GetTournamentStageList(ctx, id)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed s.GetTournamentStageList")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return struct {
			Stages []entities.TournamentStageItem `json:"stages"`
		}{
			Stages: stages,
		}, nil
	}
}

func makeGetTournamentList(s tournament.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeGetTournamentList").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[*entities.GetTournamentListRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed to cast request")
			return nil, error_templates.WrapErrorEndpoint(
				error_templates.New(pkgerr.FailedCastRequest, err, codes.InvalidArgument, http.StatusBadRequest), reqID)
		}

		err = s.GetValidator().Struct(req)

		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed validation in makeGetTournamentList")
			return nil, error_templates.WrapErrorEndpoint(
				error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest), reqID)
		}

		tournaments, count, err := s.GetTournamentList(ctx, req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed validation in makeGetTournamentList")
			return nil, error_templates.WrapErrorEndpoint(
				error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest), reqID)
		}

		return struct {
			Count       int64                      `json:"count"`
			Tournaments []entities.TournamentShort `json:"tournaments"`
		}{
			Count:       count,
			Tournaments: tournaments,
		}, nil
	}
}

func makeRecalc(s tournament.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeRecalc").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[*entities.RecalcTournamentRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed to cast request")
			return nil, error_templates.WrapErrorEndpoint(
				error_templates.New(pkgerr.FailedCastRequest, err, codes.InvalidArgument, http.StatusBadRequest), reqID)
		}

		err = s.GetValidator().Struct(req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed validation in makeRecalc")
			return nil, error_templates.WrapErrorEndpoint(
				error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest), reqID)
		}

		err = s.Recalc(ctx, req.Id)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed s.makeRecalc")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return nil, nil
	}
}
