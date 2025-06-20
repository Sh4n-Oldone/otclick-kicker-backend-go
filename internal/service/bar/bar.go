package bar

import (
	"context"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
)

func (s *Service) GetList(ctx context.Context, cityID *int64, withDeleted bool) ([]entities.Bar, error) {
	logger := s.logger.With().Interface("service", "GetList").Logger()

	list, err := s.rdbOperations.GetBarList(logger, ctx, cityID, withDeleted)
	if err != nil {
		return nil, err
	}

	return list, nil
}

func (s *Service) Create(ctx context.Context, entity entities.Bar) (*int64, error) {
	logger := s.logger.With().Interface("service", "Create").Logger()

	id, err := s.rwdbOperations.CreateBar(logger, ctx, entity)
	if err != nil {
		return nil, err
	}

	return id, nil
}

func (s *Service) Update(ctx context.Context, entity entities.UpdateBarRequest) error {
	logger := s.logger.With().Interface("service", "Update").Logger()

	err := s.rwdbOperations.UpdateBar(logger, ctx, entity)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	logger := s.logger.With().Interface("service", "DeleteBar").Logger()

	err := s.rwdbOperations.DeleteBar(logger, ctx, id)
	if err != nil {
		return err
	}

	return nil
}
