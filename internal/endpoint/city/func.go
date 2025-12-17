package city

import (
	"context"

	"github.com/go-kit/kit/endpoint"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/city"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/middleware"
	errTmpls "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/error_templates"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/errors"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/helpers"
)

func makeGetList(s city.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeCreate").Str("request_id", reqID).Logger()

		cities, err := s.GetList(ctx, request.(*entities.GetCityListRequest).WithDeleted)
		if err != nil {
			serviceLogger.Error().Stack().Err(errTmpls.ErrorDetailFromError(err)).Msg("failed to city.GetList")
			return nil, errTmpls.WrapErrorEndpoint(err, reqID)
		}

		response := &entities.GetCityListResponse{}
		response.Cities = cities

		return response, nil
	}
}

func makeCreate(s city.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeCreate").Str("request_id", reqID).Logger()

		err := helpers.ValidateCreateCityRequest(request.(*entities.CreateCityRequest))
		if err != nil {
			serviceLogger.Error().Stack().Err(errTmpls.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
			return nil, err
		}

		city := helpers.ConvertCreateCityRequestToCity(request.(*entities.CreateCityRequest))

		id, err := s.Create(ctx, *city)
		if err != nil {
			serviceLogger.Error().Stack().Err(errTmpls.ErrorDetailFromError(err)).Msg("failed to city.Create")
			return nil, errTmpls.WrapErrorEndpoint(err, reqID)
		}

		response := &struct {
			Id int64 `json:"id"`
		}{}

		response.Id = *id

		return response, nil
	}
}

func makeUpdate(s city.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeUpdate").Str("request_id", reqID).Logger()

		err := helpers.ValidateUpdateCityRequest(request.(*entities.UpdateCityRequest))
		if err != nil {
			serviceLogger.Error().Stack().Err(errTmpls.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
			return nil, err
		}

		city := helpers.ConvertUpdateCityRequestToCity(request.(*entities.UpdateCityRequest))

		err = s.Update(ctx, *city)
		if err != nil {
			serviceLogger.Error().Stack().Err(errTmpls.ErrorDetailFromError(err)).Msg("failed to city.Update")
			return nil, errTmpls.WrapErrorEndpoint(err, reqID)
		}

		response := &struct {
			Id int64 `json:"id"`
		}{}

		response.Id = city.ID

		return response, nil
	}
}

func makeDelete(s city.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeCreate").Str("request_id", reqID).Logger()

		err := s.Delete(ctx, request.(*entities.DeleteCityRequest).ID)
		if err != nil {
			serviceLogger.Error().Stack().Err(errTmpls.ErrorDetailFromError(err)).Msg("failed to city.Delete")
			return nil, errTmpls.WrapErrorEndpoint(err, reqID)
		}

		return nil, nil
	}
}
