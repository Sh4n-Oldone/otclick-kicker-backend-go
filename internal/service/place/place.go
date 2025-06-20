package place

import (
	"context"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
)

func (s *Service) GetList(ctx context.Context, barID, tableID, cityID *int64, withDeleted bool) ([]entities.Place, error) {
	logger := s.logger.With().Interface("service", "GetPlaceList").Logger()

	entities, err := s.rdbOperations.GetPlaceList(logger, ctx, barID, tableID, cityID, withDeleted)
	if err != nil {
		return nil, err
	}

	return entities, nil
}

func (s *Service) Get(ctx context.Context, id int64) (*entities.Place, error) {
	logger := s.logger.With().Interface("service", "GetPlace").Logger()

	entity, err := s.rdbOperations.GetPlaceByID(logger, ctx, id)
	if err != nil {
		return nil, err
	}

	return entity, nil
}

func (s *Service) Create(ctx context.Context, entity entities.Place) (*int64, error) {
	logger := s.logger.With().Interface("service", "CreatePlace").Logger()

	id, err := s.rwdbOperations.CreatePlace(logger, ctx, entity)
	if err != nil {
		return nil, err
	}

	return id, nil
}

func (s *Service) Update(ctx context.Context, entity entities.Place) error {
	logger := s.logger.With().Interface("service", "UpdatePlace").Logger()

	err := s.rwdbOperations.UpdatePlace(logger, ctx, entity)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	logger := s.logger.With().Interface("service", "DeletePlace").Logger()

	err := s.rwdbOperations.DeletePlace(logger, ctx, id)
	if err != nil {
		return err
	}

	return nil
}
