package table

import (
	"context"

	"github.com/go-kit/kit/endpoint"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/table"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/middleware"
	errTmpls "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/error_templates"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/errors"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/helpers"
)

func makeGetList(s table.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeGetList Table").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[*entities.GetTableListRequest](request)
		if err != nil {
			serviceLogger.Error().Stack().Err(errTmpls.ErrorDetailFromError(err)).Msg(errors.FailedCastRequest)
			return nil, err
		}

		records, err := s.GetList(ctx, req.WithDeleted)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("failed to table.GetList")
			return nil, errTmpls.WrapErrorEndpoint(err, reqID)
		}

		return &entities.GetTableListResponse{Tables: records}, nil
	}
}

func makeCreate(s table.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeCreate Table").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[*entities.CreateTableRequest](request)
		if err != nil {
			serviceLogger.Error().Stack().Err(errTmpls.ErrorDetailFromError(err)).Msg(errors.FailedCastRequest)
			return nil, err
		}

		err = helpers.ValidateCreateTableRequest(req)
		if err != nil {
			serviceLogger.Error().Stack().Err(errTmpls.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
			return nil, err
		}

		entity := &entities.Table{
			Name: req.Name,
		}

		id, err := s.Create(ctx, *entity)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("failed to table.Create")
			return nil, errTmpls.WrapErrorEndpoint(err, reqID)
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
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeUpdate Table").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[*entities.UpdateTableRequest](request)
		if err != nil {
			serviceLogger.Error().Stack().Err(errTmpls.ErrorDetailFromError(err)).Msg(errors.FailedCastRequest)
			return nil, err
		}

		err = helpers.ValidateUpdateTableRequest(req)
		if err != nil {
			serviceLogger.Error().Stack().Err(errTmpls.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
			return nil, err
		}

		entity := &entities.Table{
			ID:   req.ID,
			Name: req.Name,
		}

		err = s.Update(ctx, *entity)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("failed to table.Update")
			return nil, errTmpls.WrapErrorEndpoint(err, reqID)
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
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeDelete Table").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[*entities.DeleteTableRequest](request)
		if err != nil {
			serviceLogger.Error().Stack().Err(errTmpls.ErrorDetailFromError(err)).Msg(errors.FailedCastRequest)
			return nil, err
		}

		err = helpers.ValidateDeleteTableRequest(req)
		if err != nil {
			serviceLogger.Error().Stack().Err(errTmpls.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
			return nil, err
		}

		err = s.Delete(ctx, req.ID)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("failed to table.Delete")
			return nil, errTmpls.WrapErrorEndpoint(err, reqID)
		}

		return nil, nil
	}
}

func makeGetTableByID(s table.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeGetTableByID Table").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[*entities.GetTableRequest](request)
		if err != nil {
			serviceLogger.Error().Stack().Err(errTmpls.ErrorDetailFromError(err)).Msg(errors.FailedCastRequest)
			return nil, err
		}

		resp, err := s.GetTableByID(ctx, req.ID)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("failed to table.GetTableByID")
			return nil, errTmpls.WrapErrorEndpoint(err, reqID)
		}

		return &resp, nil
	}
}
