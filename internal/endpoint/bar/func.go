package bar

import (
	"context"

	"github.com/go-kit/kit/endpoint"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/bar"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/middleware"
	errTmpls "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/error_templates"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/errors"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/helpers"
)

func makeGetList(s bar.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeGetList Bar").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[*entities.GetBarListRequest](request)
		if err != nil {
			serviceLogger.Error().Stack().Err(errTmpls.ErrorDetailFromError(err)).Msg(errors.FailedCastRequest)
			return nil, err
		}

		records, err := s.GetList(ctx, req.CityID, req.WithDeleted)
		if err != nil {
			serviceLogger.Error().Stack().Err(errTmpls.ErrorDetailFromError(err)).
				Msg("failed to bar.GetList")
			return nil, errTmpls.WrapErrorEndpoint(err, reqID)
		}

		return &entities.GetBarListResponse{Bars: records}, nil
	}
}

func makeCreate(s bar.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeCreate Bar").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[*entities.CreateBarRequest](request)
		if err != nil {
			serviceLogger.Error().Stack().Err(errTmpls.ErrorDetailFromError(err)).Msg(errors.FailedCastRequest)
			return nil, err
		}

		err = helpers.ValidateCreateBarRequest(req)
		if err != nil {
			serviceLogger.Error().Stack().Err(errTmpls.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
			return nil, err
		}

		rel := &entities.City{
			ID: req.CityID,
		}
		entity := &entities.Bar{
			City:        *rel,
			Name:        req.Name,
			Description: req.Description,
		}

		id, err := s.Create(ctx, *entity)
		if err != nil {
			serviceLogger.Error().Stack().Err(errTmpls.ErrorDetailFromError(err)).
				Msg("failed to bar.Create")
			return nil, errTmpls.WrapErrorEndpoint(err, reqID)
		}

		response := &struct {
			ID int64 `json:"id"`
		}{}

		response.ID = *id

		return response, nil
	}
}

func makeUpdate(s bar.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeUpdate Bar").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[*entities.UpdateBarRequest](request)
		if err != nil {
			serviceLogger.Error().Stack().Err(errTmpls.ErrorDetailFromError(err)).Msg(errors.FailedCastRequest)
			return nil, err
		}

		err = helpers.ValidateUpdateBarRequest(req)
		if err != nil {
			serviceLogger.Error().Stack().Err(errTmpls.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
			return nil, err
		}

		eBar := entities.UpdateBarRequest{
			ID:          req.ID,
			Name:        req.Name,
			Description: req.Description,
		}

		err = s.Update(ctx, eBar)
		if err != nil {
			serviceLogger.Error().Stack().Err(errTmpls.ErrorDetailFromError(err)).
				Msg("failed to bar.Update")
			return nil, errTmpls.WrapErrorEndpoint(err, reqID)
		}

		response := &struct {
			ID int64 `json:"id"`
		}{}

		response.ID = req.ID

		return response, nil
	}
}

func makeDelete(s bar.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeDelete Bar").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[*entities.DeleteBarRequest](request)
		if err != nil {
			serviceLogger.Error().Stack().Err(errTmpls.ErrorDetailFromError(err)).Msg(errors.FailedCastRequest)
			return nil, err
		}

		err = helpers.ValidateDeleteBarRequest(req)
		if err != nil {
			serviceLogger.Error().Stack().Err(errTmpls.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
			return nil, err
		}

		err = s.Delete(ctx, req.ID)
		if err != nil {
			serviceLogger.Error().Stack().Err(errTmpls.ErrorDetailFromError(err)).
				Msg("failed to bar.Delete")
			return nil, errTmpls.WrapErrorEndpoint(err, reqID)
		}

		return nil, nil
	}
}
