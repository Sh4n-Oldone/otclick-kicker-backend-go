package season

import (
	"context"
	"github.com/go-kit/kit/endpoint"

	// "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/middleware"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/season"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/error_templates"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/errors"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/helpers"
)

func makeGetList(s season.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {

		seasons, err := s.GetList(ctx)
		if err != nil {
			return nil, err
		}

		response := &entities.GetSeasonResponse{}
		response.Season = seasons

		return response, nil
	}
}

func makeCreate(s season.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		// reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeCreate").Logger()

		req, err := helpers.CastRequest[*entities.CreateSeasonRequest](request)
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(errors.FailedCastRequest)
			return nil, err
		}

		err = helpers.ValidateCreateSeasonRequest(req)
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
			return nil, err
		}

		entityReq := &entities.Season{
			Name:        req.Name,
			Description: req.Description,
		}

		id, err := s.Create(ctx, *entityReq)
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(errors.ErrCreateSeason)
			return nil, err
		}

		response := &entities.CreateSeasonResponse{}
		response.Id = *id

		return response, nil
	}
}

func makeUpdate(s season.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		// reqID, ctx := middleware.GetRequestID(ctx)
		logger := s.GetLogger().With().Str("Source", "makeUpdate").Logger()

		req, err := helpers.CastRequest[*entities.UpdateSeasonRequest](request)
		if err != nil {
			logger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(errors.FailedCastRequest)
			return nil, err
		}

		err = helpers.ValidateUpdateSeasonRequest(req)
		if err != nil {
			logger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
			return nil, err
		}

		entityReq := entities.UpdateSeasonRequest{
			ID:          req.ID,
			Name:        req.Name,
			Description: req.Description,
		}

		err = s.Update(ctx, entityReq)
		if err != nil {
			logger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(errors.ErrUpdateSeason)
			return nil, err
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
		// reqID, ctx := middleware.GetRequestID(ctx)
		logger := s.GetLogger().With().Str("Source", "makeDelete").Logger()

		req, err := helpers.CastRequest[*entities.DeleteSeasonRequest](request)
		if err != nil {
			logger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(errors.FailedCastRequest)
			return nil, err
		}

		err = helpers.ValidateDeleteSeasonRequest(req)
		if err != nil {
			logger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
			return nil, err
		}

		res, err := s.Delete(ctx, req.ID)
		if err != nil {
			return res, err
		}

		response := &struct {
			Success bool `json:"success"`
		}{}

		response.Success = res

		return response, nil
	}
}
