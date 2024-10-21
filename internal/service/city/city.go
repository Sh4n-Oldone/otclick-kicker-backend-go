package city

import (
	"context"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/entity"
)

func (s *Service) GetList(ctx context.Context, withDeleted bool) ([]entity.City, error) {
	logger := s.logger.With().Interface("service", "GetList").Logger()

	cities, err := s.rdbOperations.GetCityList(logger, ctx, withDeleted)
	if err != nil {
		return nil, err
	}

	return cities, nil
}

func (s *Service) Create(ctx context.Context, city entity.City) (*int64, error) {
	logger := s.logger.With().Interface("service", "Create").Logger()

	id, err := s.rwdbOperations.CreateCity(logger, ctx, city)
	if err != nil {
		return nil, err
	}

	return &id, nil
}

func (s *Service) Update(ctx context.Context, city entity.City) error {
	logger := s.logger.With().Interface("service", "Update").Logger()

	err := s.rwdbOperations.UpdateCity(logger, ctx, city)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	logger := s.logger.With().Interface("service", "Create").Logger()

	err := s.rwdbOperations.DeleteCity(logger, ctx, id)
	if err != nil {
		return err
	}

	return nil
}
