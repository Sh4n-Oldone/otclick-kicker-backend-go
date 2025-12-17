package match

import (
	"context"

	"github.com/go-kit/kit/endpoint"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/match"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/middleware"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/error_templates"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/errors"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/helpers"
)

func makeCreate(s match.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeCreate").Str("request_id", reqID).Logger()

		err := helpers.ValidateCreateMatchRequest(request.(*entities.CreateMatchRequest))
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
			return nil, err
		}

		match := helpers.ConvertCreateMatchRequestToMatch(request.(*entities.CreateMatchRequest))

		id, err := s.Create(ctx, *match)
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).
				Msg("failed to match.Create")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		response := &struct {
			Id int64 `json:"matchId"`
		}{}

		response.Id = *id

		return response, nil
	}
}

func makeUpdate(s match.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeUpdate").Str("request_id", reqID).Logger()

		err := helpers.ValidateUpdateMatchRequest(request.(*entities.UpdateMatchRequest))
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
			return nil, err
		}

		match := helpers.ConvertUpdateMatchRequestToMatch(request.(*entities.UpdateMatchRequest))

		err = s.Update(ctx, *match)
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).
				Msg("failed to match.Update")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		response := &struct {
			Success bool `json:"success"`
		}{}

		response.Success = true

		return response, nil
	}
}

func makeDelete(s match.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeCreate").Str("request_id", reqID).Logger()

		res, err := s.Delete(ctx, request.(*entities.DeleteMatchRequest).ID)
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).
				Msg("failed to match.Delete")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		response := &struct {
			Success bool `json:"success"`
		}{}

		response.Success = res

		return response, nil
	}
}
