package role

import (
	"context"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/entity"
)

func (s *Service) GetList(ctx context.Context) ([]entity.Role, error) {
	logger := s.logger.With().Interface("service", "GetList").Logger()

	roles, err := s.rdbOperations.GetRoleList(logger, ctx)
	if err != nil {
		return nil, err
	}

	return roles, nil
}
