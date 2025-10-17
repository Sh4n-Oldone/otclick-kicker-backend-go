package role

import (
	"context"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
)

func (s *Service) GetList(ctx context.Context) ([]entities.Role, error) {
	logger := s.logger.With().Interface("service", "GetList").Logger()

	roles, err := s.rdbOperations.GetRoleList(logger, ctx)
	if err != nil {
		return nil, err
	}

	return roles, nil
}
