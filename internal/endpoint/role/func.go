package role

import (
	"context"

	"github.com/go-kit/kit/endpoint"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/role"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/middleware"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/error_templates"
)

func makeGetList(s role.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeCreate").Str("request_id", reqID).Logger()

		roles, err := s.GetList(ctx)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("failed to role.GetList")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		response := &entities.GetRoleListResponse{}
		response.Roles = roles

		return response, nil
	}
}
