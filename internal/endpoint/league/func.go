package league

import (
	"context"

	"github.com/go-kit/kit/endpoint"

	// "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/middleware"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/error_templates"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/errors"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/helpers"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/entity"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/league"
)

func makeGetList(s league.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		// reqID, ctx := middleware.GetRequestID(ctx)
		// serviceLogger := s.GetLogger().With().Str("Source", "makeCreate").Logger()

		leagues, err := s.GetList(ctx, request.(*entity.GetLeagueListRequest).CityID)
		if err != nil {
			return nil, err
		}

		response := &entity.GetLeagueListResponse{}
		response.Leagues = leagues

		return response, nil
	}
}

func makeCreate(s league.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		// reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeCreate").Logger()

		err := helpers.ValidateCreateLeagueRequest(request.(*entity.CreateLeagueRequest))
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
			return nil, err
		}

		league := helpers.ConvertCreateLeagueRequestToLeague(request.(*entity.CreateLeagueRequest))

		id, err := s.Create(ctx, *league, request.(*entity.CreateLeagueRequest).Teams)
		if err != nil {
			return nil, err
		}

		response := &entity.CreateLeagueResponse{}
		response.ID = *id

		return response, nil
	}
}

func makeUpdate(s league.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		// reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeUpdate").Logger()

		err := helpers.ValidateUpdateLeagueRequest(request.(*entity.UpdateLeagueRequest))
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
			return nil, err
		}

		league := helpers.ConvertUpdateLeagueRequestToLeague(request.(*entity.UpdateLeagueRequest))

		err = s.Update(ctx, *league, request.(*entity.UpdateLeagueRequest).Teams)
		if err != nil {
			return nil, err
		}

		response := &entity.UpdateLeagueResponse{}
		response.ID = league.ID

		return response, nil
	}
}

func makeDelete(s league.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		// reqID, ctx := middleware.GetRequestID(ctx)
		// serviceLogger := s.GetLogger().With().Str("Source", "makeCreate").Logger()

		err := s.Delete(ctx, request.(*entity.DeleteLeagueRequest).ID)
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

		err := s.Recalc(ctx, request.(*entity.RecalcLeagueRequest).ID)
		if err != nil {
			return nil, err
		}

		return nil, nil
	}
}
