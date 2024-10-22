package match

import (
	"context"

	"github.com/go-kit/kit/endpoint"

	// "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/middleware"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/error_templates"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/errors"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/helpers"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/entity"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/match"
)

func makeCreate(s match.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		// reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeCreate").Logger()

		err := helpers.ValidateCreateMatchRequest(request.(*entity.CreateMatchRequest))
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
			return nil, err
		}

		match := helpers.ConvertCreateMatchRequestToMatch(request.(*entity.CreateMatchRequest))

		id, err := s.Create(ctx, *match)
		if err != nil {
			return nil, err
		}

		response := &struct{
			Id int64 `json:"id"`
		}{}

		response.Id = *id

		return response, nil
	}
}

func makeUpdate(s match.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		// reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeUpdate").Logger()

		err := helpers.ValidateUpdateMatchRequest(request.(*entity.UpdateMatchRequest))
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
			return nil, err
		}

		match := helpers.ConvertUpdateMatchRequestToMatch(request.(*entity.UpdateMatchRequest))

		err = s.Update(ctx, *match)
		if err != nil {
			return nil, err
		}

		response := &struct{
			Id int64 `json:"id"`
		}{}

		response.Id = match.ID

		return response, nil
	}
}

func makeDelete(s match.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		// reqID, ctx := middleware.GetRequestID(ctx)
		// serviceLogger := s.GetLogger().With().Str("Source", "makeCreate").Logger()

		err := s.Delete(ctx, request.(*entity.DeleteMatchRequest).ID)
		if err != nil {
			return nil, err
		}

		return nil, nil
	}
}

