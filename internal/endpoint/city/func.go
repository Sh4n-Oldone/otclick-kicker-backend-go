package city

import (
	"context"

	"github.com/go-kit/kit/endpoint"

	// "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/middleware"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/error_templates"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/errors"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/helpers"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/city"
)

func makeGetList(s city.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		// reqID, ctx := middleware.GetRequestID(ctx)
		// serviceLogger := s.GetLogger().With().Str("Source", "makeCreate").Logger()

		cities, err := s.GetList(ctx, request.(*entities.GetCityListRequest).WithDeleted)
		if err != nil {
			return nil, err
		}

		response := &entities.GetCityListResponse{}
		response.Cities = cities

		return response, nil
	}
}

func makeCreate(s city.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		// reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeCreate").Logger()

		err := helpers.ValidateCreateCityRequest(request.(*entities.CreateCityRequest))
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
			return nil, err
		}

		city := helpers.ConvertCreateCityRequestToCity(request.(*entities.CreateCityRequest))

		id, err := s.Create(ctx, *city)
		if err != nil {
			return nil, err
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
		// reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeUpdate").Logger()

		err := helpers.ValidateUpdateCityRequest(request.(*entities.UpdateCityRequest))
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
			return nil, err
		}

		city := helpers.ConvertUpdateCityRequestToCity(request.(*entities.UpdateCityRequest))

		err = s.Update(ctx, *city)
		if err != nil {
			return nil, err
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
		// reqID, ctx := middleware.GetRequestID(ctx)
		// serviceLogger := s.GetLogger().With().Str("Source", "makeCreate").Logger()

		err := s.Delete(ctx, request.(*entities.DeleteCityRequest).ID)
		if err != nil {
			return nil, err
		}

		return nil, nil
	}
}
