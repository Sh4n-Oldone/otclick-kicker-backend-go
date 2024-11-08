package place

import (
	"context"

	"github.com/go-kit/kit/endpoint"

	// "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/middleware"

	errTmpls "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/error_templates"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/errors"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/helpers"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/entity"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/place"
)

func makeGetList(s place.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		// reqID, ctx := middleware.GetRequestID(ctx)
		logger := s.GetLogger().With().Str("Source", "makeGetList Place").Logger()

		req, err := helpers.CastRequest[*entity.GetPlaceListRequest](request)
		if err != nil {
			logger.Error().Stack().Err(errTmpls.ErrorDetailFromError(err)).Msg(errors.FailedCastRequest)
			return nil, err
		}

		records, err := s.GetList(ctx, req.BarID, req.TableID, req.CityID, req.WithDeleted)
		if err != nil {
			logger.Error().Stack().Err(errTmpls.ErrorDetailFromError(err)).Msg(errors.ErrGetPlaceList)
			return nil, err
		}

		return &entity.GetPlaceListResponse{Places: records}, nil
	}
}

func makeCreate(s place.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		// reqID, ctx := middleware.GetRequestID(ctx)
		logger := s.GetLogger().With().Str("Source", "makeCreate Place").Logger()

		req, err := helpers.CastRequest[*entity.CreatePlaceRequest](request)
		if err != nil {
			logger.Error().Stack().Err(errTmpls.ErrorDetailFromError(err)).Msg(errors.FailedCastRequest)
			return nil, err
		}

		err = helpers.ValidateCreatePlaceRequest(req)
		if err != nil {
			logger.Error().Stack().Err(errTmpls.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
			return nil, err
		}

		rel1 := &entity.Bar{
			ID: req.BarID,
		}
		rel2 := &entity.Table{
			ID: req.TableID,
		}
		entity := &entity.Place{
			Bar:   *rel1,
			Table: *rel2,
		}

		id, err := s.Create(ctx, *entity)
		if err != nil {
			logger.Error().Stack().Err(errTmpls.ErrorDetailFromError(err)).Msg(errors.ErrCreatePlace)
			return nil, err
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
		// reqID, ctx := middleware.GetRequestID(ctx)
		logger := s.GetLogger().With().Str("Source", "makeUpdate Place").Logger()

		req, err := helpers.CastRequest[*entity.UpdatePlaceRequest](request)
		if err != nil {
			logger.Error().Stack().Err(errTmpls.ErrorDetailFromError(err)).Msg(errors.FailedCastRequest)
			return nil, err
		}

		err = helpers.ValidateUpdatePlaceRequest(req)
		if err != nil {
			logger.Error().Stack().Err(errTmpls.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
			return nil, err
		}

		rel1 := &entity.Bar{
			ID: req.BarID,
		}
		rel2 := &entity.Table{
			ID: req.TableID,
		}
		entity := &entity.Place{
			ID:    req.ID,
			Bar:   *rel1,
			Table: *rel2,
		}

		err = s.Update(ctx, *entity)
		if err != nil {
			logger.Error().Stack().Err(errTmpls.ErrorDetailFromError(err)).Msg(errors.ErrDeletePlace)
			return nil, err
		}

		response := &struct {
			ID int64 `json:"id"`
		}{}

		response.ID = req.ID

		return response, nil
	}
}

func makeDelete(s place.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		// reqID, ctx := middleware.GetRequestID(ctx)
		logger := s.GetLogger().With().Str("Source", "makeDelete Place").Logger()

		req, err := helpers.CastRequest[*entity.DeletePlaceRequest](request)
		if err != nil {
			logger.Error().Stack().Err(errTmpls.ErrorDetailFromError(err)).Msg(errors.FailedCastRequest)
			return nil, err
		}

		err = helpers.ValidateDeletePlaceRequest(req)
		if err != nil {
			logger.Error().Stack().Err(errTmpls.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
			return nil, err
		}

		err = s.Delete(ctx, req.ID)
		if err != nil {
			logger.Error().Stack().Err(errTmpls.ErrorDetailFromError(err)).Msg(errors.ErrDeletePlace)
			return nil, err
		}

		return nil, nil
	}
}
