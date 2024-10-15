package template

import (
	"context"
)

func (s *Service) Create(ctx context.Context) (*int64, error) {
	logger := s.logger.With().Interface("service", "Create").Logger()

	timeout, cancel := context.WithTimeout(ctx, s.config.RWDB.MaxIdleConnectionTimeout)
	defer cancel()

	s.logger.Error().Msg("Not implemented")
	return nil, nil
}
