package table

import (
	"context"

	"github.com/go-kit/kit/endpoint"

	// "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/middleware"

	errTmpls "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/error_templates"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/errors"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/helpers"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/entity"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/table"
)

func makeGetList(s table.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		// reqID, ctx := middleware.GetRequestID(ctx)
		logger := s.GetLogger().With().Str("Source", "makeGetList Table").Logger()

		req, err := helpers.CastRequest[*entity.GetTableListRequest](request)
		if err != nil {
			logger.Error().Stack().Err(errTmpls.ErrorDetailFromError(err)).Msg(errors.FailedCastRequest)
			return nil, err
		}

		records, err := s.GetList(ctx, req.WithDeleted)
		if err != nil {
			logger.Error().Stack().Err(errTmpls.ErrorDetailFromError(err)).Msg(errors.ErrGetTableList)
			return nil, err
		}

		return &entity.GetTableListResponse{Tables: records}, nil
	}
}

func makeCreate(s table.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		// reqID, ctx := middleware.GetRequestID(ctx)
		logger := s.GetLogger().With().Str("Source", "makeCreate Table").Logger()

		req, err := helpers.CastRequest[*entity.CreateTableRequest](request)
		if err != nil {
			logger.Error().Stack().Err(errTmpls.ErrorDetailFromError(err)).Msg(errors.FailedCastRequest)
			return nil, err
		}

		err = helpers.ValidateCreateTableRequest(req)
		if err != nil {
			logger.Error().Stack().Err(errTmpls.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
			return nil, err
		}

		entity := &entity.Table{
			Name: req.Name,
		}

		id, err := s.Create(ctx, *entity)
		if err != nil {
			logger.Error().Stack().Err(errTmpls.ErrorDetailFromError(err)).Msg(errors.ErrCreateTable)
			return nil, err
		}

		response := &struct {
			ID int64 `json:"id"`
		}{}

		response.ID = *id

		return response, nil
	}
}

func makeUpdate(s table.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		// reqID, ctx := middleware.GetRequestID(ctx)
		logger := s.GetLogger().With().Str("Source", "makeUpdate Table").Logger()

		req, err := helpers.CastRequest[*entity.UpdateTableRequest](request)
		if err != nil {
			logger.Error().Stack().Err(errTmpls.ErrorDetailFromError(err)).Msg(errors.FailedCastRequest)
			return nil, err
		}

		err = helpers.ValidateUpdateTableRequest(req)
		if err != nil {
			logger.Error().Stack().Err(errTmpls.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
			return nil, err
		}

		entity := &entity.Table{
			ID:   req.ID,
			Name: req.Name,
		}

		err = s.Update(ctx, *entity)
		if err != nil {
			logger.Error().Stack().Err(errTmpls.ErrorDetailFromError(err)).Msg(errors.ErrUpdateTable)
			return nil, err
		}

		response := &struct {
			ID int64 `json:"id"`
		}{}

		response.ID = req.ID

		return response, nil
	}
}

func makeDelete(s table.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		// reqID, ctx := middleware.GetRequestID(ctx)
		logger := s.GetLogger().With().Str("Source", "makeDelete Table").Logger()

		req, err := helpers.CastRequest[*entity.DeleteTableRequest](request)
		if err != nil {
			logger.Error().Stack().Err(errTmpls.ErrorDetailFromError(err)).Msg(errors.FailedCastRequest)
			return nil, err
		}

		err = helpers.ValidateDeleteTableRequest(req)
		if err != nil {
			logger.Error().Stack().Err(errTmpls.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
			return nil, err
		}

		err = s.Delete(ctx, req.ID)
		if err != nil {
			logger.Error().Stack().Err(errTmpls.ErrorDetailFromError(err)).Msg(errors.ErrDeleteTable)
			return nil, err
		}

		return nil, nil
	}
}
