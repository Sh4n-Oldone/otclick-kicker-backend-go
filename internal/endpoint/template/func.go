package template

import (
	"context"
	"github.com/go-kit/kit/endpoint"
	pb "node71.otclick.ru/backend/rhumb-go-genproto/rhumb/rhumb_api/template/v1"
	"node71.otclick.ru/backend/template/internal/transport/http/middleware"
	"node71.otclick.ru/backend/template/pkg/error_templates"
	"node71.otclick.ru/backend/template/pkg/errors"
	"node71.otclick.ru/backend/template/pkg/helpers"

	"node71.otclick.ru/backend/template/internal/service/template"
)

func makeCreate(s template.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeCreate").Logger()

		req, err := helpers.CastValidateRequest[*pb.CreateRequest](s.GetValidator(), request)
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(errors.FailedCastValidateRequest)
			return nil, err
		}

		serviceLogger.Error().Msg("Not implemented")

		return nil, nil
	}
}
