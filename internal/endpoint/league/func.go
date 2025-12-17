package league

import (
	"context"

	"github.com/go-kit/kit/endpoint"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/middleware"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/error_templates"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/errors"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/helpers"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/league"
)

// makeGetList deprecated
func makeGetList(s league.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeCreate").Str("request_id", reqID).Logger()

		leagues, err := s.GetList(ctx, request.(*entities.GetLeagueListRequest).CityID)
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).
				Msg("failed to league.GetList")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		response := &entities.GetLeagueListResponse{}
		response.Leagues = leagues

		return response, nil
	}
}

// makeCreate deprecated
func makeCreate(s league.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeCreate").Str("request_id", reqID).Logger()

		err := helpers.ValidateCreateLeagueRequest(request.(*entities.CreateLeagueRequest))
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		req := request.(*entities.CreateLeagueRequest)

		id, err := s.Create(ctx, req)
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).
				Msg("failed to league.Create")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		response := &entities.CreateLeagueResponse{}
		response.ID = *id

		return response, nil
	}
}

// makeUpdate deprecated
func makeUpdate(s league.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeUpdate").Str("request_id", reqID).Logger()

		err := helpers.ValidateUpdateLeagueRequest(request.(*entities.UpdateLeagueRequest))
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
			return nil, err
		}

		league := helpers.ConvertUpdateLeagueRequestToLeague(request.(*entities.UpdateLeagueRequest))

		err = s.Update(ctx, *league, request.(*entities.UpdateLeagueRequest).Teams)
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).
				Msg("failed to league.Update")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		response := &entities.UpdateLeagueResponse{}
		response.ID = league.ID

		return response, nil
	}
}

// makeDelete deprecated
func makeDelete(s league.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeCreate").Str("request_id", reqID).Logger()

		err := s.Delete(ctx, request.(*entities.DeleteLeagueRequest).ID)
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).
				Msg("failed to league.Delete")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return nil, nil
	}
}

// makeRecalc deprecated
func makeRecalc(s league.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeCreate").Str("request_id", reqID).Logger()

		err := s.Recalc(ctx, request.(*entities.RecalcLeagueRequest).ID)
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).
				Msg("failed to league.Recalc")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return nil, nil
	}
}

// makeCreateExtraPoints deprecated
func makeCreateExtraPoints(s league.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeCreateExtraPoints").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[*entities.CreateExtraPointsRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(errors.FailedCastRequest)
			return false, error_templates.WrapErrorEndpoint(err, reqID)
		}

		err = helpers.ValidateCreateExtraPointsRequest(req)
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
			return false, error_templates.WrapErrorEndpoint(err, reqID)
		}

		res, err := s.CreateExtraPoints(ctx, req)
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).
				Msg("failed to league.CreateExtraPoints")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		response := &struct {
			Id int64 `json:"id"`
		}{}
		response.Id = res
		return response, nil
	}
}

// makeUpdateExtraPoints deprecated
func makeUpdateExtraPoints(s league.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeUpdateExtraPoints").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[*entities.UpdateExtraPointsRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(errors.FailedCastRequest)
			return false, error_templates.WrapErrorEndpoint(err, reqID)
		}

		err = helpers.ValidateUpdateExtraPointsRequest(req)
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
			return false, error_templates.WrapErrorEndpoint(err, reqID)
		}

		res, err := s.UpdateExtraPoints(ctx, req)
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).
				Msg("failed to league.UpdateExtraPoints")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		response := &struct {
			Success bool `json:"success"`
		}{}
		response.Success = res

		return response, nil
	}
}

// makeDeleteExtraPoints deprecated
func makeDeleteExtraPoints(s league.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeDeleteExtraPoints").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[*entities.IdRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(errors.FailedCastRequest)
			return false, error_templates.WrapErrorEndpoint(err, reqID)
		}

		err = helpers.ValidateIdRequest(req)
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
			return false, error_templates.WrapErrorEndpoint(err, reqID)
		}

		res, err := s.DeleteExtraPoints(ctx, req.Id)
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).
				Msg("failed to league.DeleteExtraPoints")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		response := &struct {
			Success bool `json:"success"`
		}{}
		response.Success = res

		return response, nil
	}
}

// makeGetExtraPointsListByTeamAndLeagueId deprecated
func makeGetExtraPointsListByTeamAndLeagueId(s league.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeGetExtraPointsListByTeamAndLeagueId").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[*entities.TeamLeagueIdRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(errors.FailedCastRequest)
			return false, error_templates.WrapErrorEndpoint(err, reqID)
		}

		err = helpers.ValidateTeamLeagueIdRequest(req)
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
			return false, error_templates.WrapErrorEndpoint(err, reqID)
		}

		res, err := s.GetExtraPointsListByTeamAndLeagueId(ctx, req.TeamId, req.LeagueId)
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).
				Msg("failed to league.GetExtraPointsListByTeamAndLeagueId")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return res, nil
	}
}

// makeGetExtraPointsById deprecated
func makeGetExtraPointsById(s league.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeGetExtraPointsById").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[*entities.IdRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(errors.FailedCastRequest)
			return false, error_templates.WrapErrorEndpoint(err, reqID)
		}

		err = helpers.ValidateIdRequest(req)
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
			return false, error_templates.WrapErrorEndpoint(err, reqID)
		}

		res, err := s.GetExtraPointsById(ctx, req.Id)
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).
				Msg("failed to league.GetExtraPointsById")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return res, nil
	}
}
