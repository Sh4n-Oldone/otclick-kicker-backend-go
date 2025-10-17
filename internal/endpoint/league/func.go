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

func makeGetList(s league.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		// reqID, ctx := middleware.GetRequestID(ctx)
		// serviceLogger := s.GetLogger().With().Str("Source", "makeCreate").Logger()

		leagues, err := s.GetList(ctx, request.(*entities.GetLeagueListRequest).CityID)
		if err != nil {
			return nil, err
		}

		response := &entities.GetLeagueListResponse{}
		response.Leagues = leagues

		return response, nil
	}
}

func makeCreate(s league.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeCreate").Logger()

		err := helpers.ValidateCreateLeagueRequest(request.(*entities.CreateLeagueRequest))
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		req := request.(*entities.CreateLeagueRequest)

		id, err := s.Create(ctx, req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed to league.makeCreate")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		response := &entities.CreateLeagueResponse{}
		response.ID = *id

		return response, nil
	}
}

func makeUpdate(s league.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		// reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeUpdate").Logger()

		err := helpers.ValidateUpdateLeagueRequest(request.(*entities.UpdateLeagueRequest))
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
			return nil, err
		}

		league := helpers.ConvertUpdateLeagueRequestToLeague(request.(*entities.UpdateLeagueRequest))

		err = s.Update(ctx, *league, request.(*entities.UpdateLeagueRequest).Teams)
		if err != nil {
			return nil, err
		}

		response := &entities.UpdateLeagueResponse{}
		response.ID = league.ID

		return response, nil
	}
}

func makeDelete(s league.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		// reqID, ctx := middleware.GetRequestID(ctx)
		// serviceLogger := s.GetLogger().With().Str("Source", "makeCreate").Logger()

		err := s.Delete(ctx, request.(*entities.DeleteLeagueRequest).ID)
		if err != nil {
			return nil, err
		}

		return nil, nil
	}
}

func makeRecalc(s league.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		// reqID, ctx := middleware.GetRequestID(ctx)
		// serviceLogger := s.GetLogger().With().Str("Source", "makeCreate").Logger()

		err := s.Recalc(ctx, request.(*entities.RecalcLeagueRequest).ID)
		if err != nil {
			return nil, err
		}

		return nil, nil
	}
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////////////

func makeCreateExtraPoints(s league.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeCreateExtraPoints").Logger()

		req, err := helpers.CastRequest[*entities.CreateExtraPointsRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed to cast request")
			return false, error_templates.WrapErrorEndpoint(err, reqID)
		}

		err = helpers.ValidateCreateExtraPointsRequest(req)
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
			return false, error_templates.WrapErrorEndpoint(err, reqID)
		}

		res, err := s.CreateExtraPoints(ctx, req)
		if err != nil {
			return res, error_templates.WrapErrorEndpoint(err, reqID)
		}

		response := &struct {
			Id int64 `json:"id"`
		}{}
		response.Id = res
		return response, nil
	}
}

func makeUpdateExtraPoints(s league.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeUpdateExtraPoints").Logger()

		req, err := helpers.CastRequest[*entities.UpdateExtraPointsRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed to cast request")
			return false, error_templates.WrapErrorEndpoint(err, reqID)
		}

		err = helpers.ValidateUpdateExtraPointsRequest(req)
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
			return false, error_templates.WrapErrorEndpoint(err, reqID)
		}

		res, err := s.UpdateExtraPoints(ctx, req)
		if err != nil {
			return res, error_templates.WrapErrorEndpoint(err, reqID)
		}

		response := &struct {
			Success bool `json:"success"`
		}{}
		response.Success = res

		return response, nil
	}
}

func makeDeleteExtraPoints(s league.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeDeleteExtraPoints").Logger()

		req, err := helpers.CastRequest[*entities.IdRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed to cast request")
			return false, error_templates.WrapErrorEndpoint(err, reqID)
		}

		err = helpers.ValidateIdRequest(req)
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
			return false, error_templates.WrapErrorEndpoint(err, reqID)
		}

		res, err := s.DeleteExtraPoints(ctx, req.Id)
		if err != nil {
			return res, error_templates.WrapErrorEndpoint(err, reqID)
		}

		response := &struct {
			Success bool `json:"success"`
		}{}
		response.Success = res

		return response, nil
	}
}

func makeGetExtraPointsListByTeamAndLeagueId(s league.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeGetExtraPointsListByTeamAndLeagueId").Logger()

		req, err := helpers.CastRequest[*entities.TeamLeagueIdRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed to cast request")
			return false, error_templates.WrapErrorEndpoint(err, reqID)
		}

		err = helpers.ValidateTeamLeagueIdRequest(req)
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
			return false, error_templates.WrapErrorEndpoint(err, reqID)
		}

		res, err := s.GetExtraPointsListByTeamAndLeagueId(ctx, req.TeamId, req.LeagueId)
		if err != nil {
			return res, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return res, nil
	}
}

func makeGetExtraPointsById(s league.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeGetExtraPointsById").Logger()

		req, err := helpers.CastRequest[*entities.IdRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed to cast request")
			return false, error_templates.WrapErrorEndpoint(err, reqID)
		}

		err = helpers.ValidateIdRequest(req)
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
			return false, error_templates.WrapErrorEndpoint(err, reqID)
		}

		res, err := s.GetExtraPointsById(ctx, req.Id)
		if err != nil {
			return res, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return res, nil
	}
}
