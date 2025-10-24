package season

import (
	"context"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
)

func (s *Service) GetList(ctx context.Context, req *entities.GetSeasonListRequest) ([]entities.Season, error) {
	logger := s.logger.With().Str("service", "GetList").Logger()

	seasons, err := s.rdbOperations.GetSeasonList(logger, ctx, req)
	if err != nil {
		return nil, err
	}

	return seasons, nil
}

func (s *Service) Create(ctx context.Context, entity entities.Season) (*int64, error) {
	logger := s.logger.With().Str("service", "Create").Logger()

	id, err := s.rwdbOperations.CreateSeason(logger, ctx, entity)
	if err != nil {
		return nil, err
	}

	return id, nil
}

func (s *Service) Update(ctx context.Context, entity entities.UpdateSeasonRequest) error {
	logger := s.logger.With().Str("service", "Update").Logger()

	err := s.rwdbOperations.UpdateSeason(logger, ctx, entity)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Delete(ctx context.Context, id int64) (bool, error) {
	logger := s.logger.With().Str("service", "Delete").Logger()

	res, err := s.rwdbOperations.DeleteSeason(logger, ctx, id)
	if err != nil {
		return false, err
	}

	return res, nil
}
