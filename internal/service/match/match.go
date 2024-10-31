package match

import (
	"context"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/entity"
)

func (s *Service) Create(ctx context.Context, match entity.Match) (*int64, error) {
	logger := s.logger.With().Interface("service", "Create").Logger()

	id, err := s.rwdbOperations.CreateMatch(logger, ctx, match)
	if err != nil {
		return nil, err
	}

	return &id, nil
}

func (s *Service) Update(ctx context.Context, match entity.Match) error {
	logger := s.logger.With().Interface("service", "Update").Logger()

	err := s.rwdbOperations.UpdateMatch(logger, ctx, match)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Delete(ctx context.Context, id int64) (bool, error) {
	logger := s.logger.With().Interface("service", "Delete").Logger()

	res, err := s.rwdbOperations.DeleteMatch(logger, ctx, id)
	if err != nil {
		return res, err
	}

	return res, nil
}
