package season

import (
	"context"
	"net/http"

	"github.com/go-kit/kit/endpoint"
	"google.golang.org/grpc/codes"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/season"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/middleware"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/error_templates"
	pkgerr "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/errors"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/helpers"
)

func makeGetList(s season.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeGetList").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[*entities.GetSeasonListRequest](request)
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(pkgerr.FailedCastRequest)
			return nil, error_templates.WrapErrorEndpoint(
				error_templates.New(pkgerr.FailedCastRequest, err, codes.InvalidArgument, http.StatusBadRequest), reqID)
		}

		seasons, err := s.GetList(ctx, req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("failed to season.GetList")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		response := &entities.GetSeasonResponse{}
		response.SeasonList = seasons

		return response, nil
	}
}

func makeCreate(s season.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeCreate").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[*entities.CreateSeasonRequest](request)
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(pkgerr.FailedCastRequest)
			return nil, error_templates.WrapErrorEndpoint(
				error_templates.New(pkgerr.FailedCastRequest, err, codes.InvalidArgument, http.StatusBadRequest), reqID)
		}

		err = helpers.ValidateCreateSeasonRequest(req)
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(pkgerr.FailedValidateRequest)
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		entityReq := &entities.Season{
			Name:        req.Name,
			Description: req.Description,
		}

		id, err := s.Create(ctx, *entityReq)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("failed to season.Create")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		response := &entities.CreateSeasonResponse{}
		response.Id = *id

		return response, nil
	}
}

func makeUpdate(s season.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeUpdate").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[*entities.UpdateSeasonRequest](request)
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(pkgerr.FailedCastRequest)
			return nil, error_templates.WrapErrorEndpoint(
				error_templates.New(pkgerr.FailedCastRequest, err, codes.InvalidArgument, http.StatusBadRequest), reqID)
		}

		err = helpers.ValidateUpdateSeasonRequest(req)
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(pkgerr.FailedValidateRequest)
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		entityReq := entities.UpdateSeasonRequest{
			ID:          req.ID,
			Name:        req.Name,
			Description: req.Description,
		}

		err = s.Update(ctx, entityReq)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("failed to season.Update")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		response := &struct {
			ID int64 `json:"id"`
		}{}

		response.ID = req.ID

		return response, nil
	}
}

func makeDelete(s season.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeDelete").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[*entities.DeleteSeasonRequest](request)
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(pkgerr.FailedCastRequest)
			return nil, error_templates.WrapErrorEndpoint(
				error_templates.New(pkgerr.FailedCastRequest, err, codes.InvalidArgument, http.StatusBadRequest), reqID)
		}

		err = helpers.ValidateDeleteSeasonRequest(req)
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(pkgerr.FailedValidateRequest)
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		res, err := s.Delete(ctx, req.ID)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("failed to season.Delete")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		response := &struct {
			Success bool `json:"success"`
		}{}

		response.Success = res

		return response, nil
	}
}
