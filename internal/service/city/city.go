package city

import (
	"context"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
)

func (s *Service) GetList(ctx context.Context, withDeleted bool) ([]entities.City, error) {
	logger := s.logger.With().Interface("service", "GetCityList").Logger()

	cities, err := s.rdbOperations.GetCityList(logger, ctx, withDeleted)
	if err != nil {
		return nil, err
	}

	return cities, nil
}

func (s *Service) Create(ctx context.Context, city entities.City) (*int64, error) {
	logger := s.logger.With().Interface("service", "CreateCity").Logger()

	id, err := s.rwdbOperations.CreateCity(logger, ctx, city)
	if err != nil {
		return nil, err
	}

	return id, nil
}

func (s *Service) Update(ctx context.Context, city entities.City) error {
	logger := s.logger.With().Interface("service", "UpdateCity").Logger()

	err := s.rwdbOperations.UpdateCity(logger, ctx, city)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	logger := s.logger.With().Interface("service", "DeleteCity").Logger()

	err := s.rwdbOperations.DeleteCity(logger, ctx, id)
	if err != nil {
		return err
	}

	return nil
}
