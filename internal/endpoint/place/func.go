package place

import (
	"context"
	"strconv"

	"github.com/go-kit/kit/endpoint"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/place"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/middleware"
	errTmpls "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/error_templates"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/errors"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/helpers"
)

func makeGetList(s place.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		logger := s.GetLogger().With().Str("Source", "makeGetList Place").Logger()

		req, err := helpers.CastRequest[*entities.GetPlaceListRequest](request)
		if err != nil {
			logger.Error().Stack().Err(errTmpls.ErrorDetailFromError(err)).Msg(errors.FailedCastRequest)
			return nil, errTmpls.WrapErrorEndpoint(err, reqID)
		}

		records, err := s.GetList(ctx, req.BarID, req.TableID, req.CityID, req.WithDeleted)
		if err != nil {
			logger.Error().Stack().Err(errTmpls.ErrorDetailFromError(err)).Msg("failed to place.GetList")
			return nil, errTmpls.WrapErrorEndpoint(err, reqID)
		}

		return &entities.GetPlaceListResponse{Places: records}, nil
	}
}

func makeCreate(s place.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		logger := s.GetLogger().With().Str("Source", "makeCreate").Logger()

		req, err := helpers.CastRequest[*entities.CreatePlaceRequest](request)
		if err != nil {
			logger.Error().Stack().Err(errTmpls.ErrorDetailFromError(err)).Msg(errors.FailedCastRequest)
			return nil, errTmpls.WrapErrorEndpoint(err, reqID)
		}

		err = s.GetValidator().Struct(req)
		if err != nil {
			logger.Error().Stack().Err(errTmpls.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
			return nil, errTmpls.WrapErrorEndpoint(err, reqID)
		}

		id, err := s.Create(ctx, *req)
		if err != nil {
			logger.Error().Stack().Err(errTmpls.ErrorDetailFromError(err)).Msg("failed to place.Create")
			return nil, errTmpls.WrapErrorEndpoint(err, reqID)
		}

		response := &struct {
			ID int64 `json:"id"`
		}{}

		response.ID = *id

		return response, nil
	}
}

func makeUpdate(s place.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		logger := s.GetLogger().With().Str("Source", "makeUpdate").Logger()

		req, err := helpers.CastRequest[*entities.UpdatePlaceRequest](request)
		if err != nil {
			logger.Error().Stack().Err(errTmpls.ErrorDetailFromError(err)).Msg(errors.FailedCastRequest)
			return nil, errTmpls.WrapErrorEndpoint(err, reqID)
		}

		err = s.GetValidator().Struct(req)
		if err != nil {
			logger.Error().Stack().Err(errTmpls.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
			return nil, errTmpls.WrapErrorEndpoint(err, reqID)
		}

		err = s.Update(ctx, *req)
		if err != nil {
			logger.Error().Stack().Err(errTmpls.ErrorDetailFromError(err)).Msg("failed to place.Update")
			return nil, errTmpls.WrapErrorEndpoint(err, reqID)
		}

		response := &struct {
			ID int64 `json:"id"`
		}{}

		response.ID = req.PlaceID

		return response, nil
	}
}

func makeDelete(s place.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		logger := s.GetLogger().With().Str("Source", "makeDelete").Logger()

		req, err := helpers.CastRequest[*entities.DeletePlaceRequest](request)
		if err != nil {
			logger.Error().Stack().Err(errTmpls.ErrorDetailFromError(err)).Msg(errors.FailedCastRequest)
			return nil, errTmpls.WrapErrorEndpoint(err, reqID)
		}

		err = s.GetValidator().Struct(req)
		if err != nil {
			logger.Error().Stack().Err(errTmpls.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
			return nil, errTmpls.WrapErrorEndpoint(err, reqID)
		}

		err = s.Delete(ctx, req.ID, req.Executor)
		if err != nil {
			logger.Error().Stack().Err(errTmpls.ErrorDetailFromError(err)).Msg("failed to place.Delete")
			return nil, errTmpls.WrapErrorEndpoint(err, reqID)
		}

		return struct {
			Message string `json:"message"`
		}{
			Message: "place " + strconv.FormatInt(req.ID, 10) + " deleted",
		}, nil
	}
}
