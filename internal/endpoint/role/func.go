package role

import (
	"context"

	"github.com/go-kit/kit/endpoint"

	// "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/middleware"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/entity"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/role"
)

func makeGetList(s role.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		// reqID, ctx := middleware.GetRequestID(ctx)
		// serviceLogger := s.GetLogger().With().Str("Source", "makeCreate").Logger()

		roles, err := s.GetList(ctx)
		if err != nil {
			return nil, err
		}

		response := &entity.GetRoleListResponse{}
		response.Roles = roles

		return response, nil
	}
}
