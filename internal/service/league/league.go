package league

import (
	"context"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/entity"
)

func (s *Service) GetList(ctx context.Context, cityID int64) ([]entity.League, error) {
	logger := s.logger.With().Interface("service", "GetList").Logger()

	leagues, err := s.rdbOperations.GetLeagueList(logger, ctx, cityID)
	if err != nil {
		return nil, err
	}

	return leagues, nil
}

func (s *Service) Create(ctx context.Context, league entity.League, teams []int64) (*int64, error) {
	logger := s.logger.With().Interface("service", "Create").Logger()

	id, err := s.rwdbOperations.CreateLeague(logger, ctx, league, teams)
	if err != nil {
		return nil, err
	}

	return &id, nil
}

func (s *Service) Update(ctx context.Context, league entity.League, teams []int64) error {
	logger := s.logger.With().Interface("service", "Update").Logger()

	err := s.rwdbOperations.UpdateLeague(logger, ctx, league, teams)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	logger := s.logger.With().Interface("service", "Create").Logger()

	err := s.rwdbOperations.DeleteLeague(logger, ctx, id)
	if err != nil {
		return err
	}

	return nil
}
