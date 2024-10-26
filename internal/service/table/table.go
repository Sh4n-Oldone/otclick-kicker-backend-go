package table

import (
	"context"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/entity"
)

func (s *Service) GetList(ctx context.Context, withDeleted bool) ([]entity.Table, error) {
	logger := s.logger.With().Interface("service", "GetTableList").Logger()

	records, err := s.rdbOperations.GetTableList(logger, ctx, withDeleted)
	if err != nil {
		return nil, err
	}

	return records, nil
}

func (s *Service) Create(ctx context.Context, entity entity.Table) (*int64, error) {
	logger := s.logger.With().Interface("service", "CreateTable").Logger()

	id, err := s.rwdbOperations.CreateTable(logger, ctx, entity)
	if err != nil {
		return nil, err
	}

	return id, nil
}

func (s *Service) Update(ctx context.Context, entity entity.Table) error {
	logger := s.logger.With().Interface("service", "UpdateTable").Logger()

	err := s.rwdbOperations.UpdateTable(logger, ctx, entity)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	logger := s.logger.With().Interface("service", "DeleteTable").Logger()

	err := s.rwdbOperations.DeleteTable(logger, ctx, id)
	if err != nil {
		return err
	}

	return nil
}
