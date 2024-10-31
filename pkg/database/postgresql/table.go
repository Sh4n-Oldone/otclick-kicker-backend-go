package postgresql

import (
	"context"
	stderr "errors"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/entity"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/errors"
)

func (db *RDBOperation) GetTableList(logger zerolog.Logger, ctx context.Context, withDeleted bool) ([]entity.Table, error) {
	query := queryGetTableList
	if withDeleted {
		query = queryGetTableListWithDeleted
	}

	rows, err := db.db.Query(ctx, query)	
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to postgresql.GetTableList")
		return nil, DecodeDatabaseError(err)
	}
	defer rows.Close()

	var entities []entity.Table

	for rows.Next() {
		var entity entity.Table

		err = rows.Scan(&entity.ID, &entity.Name, &entity.UpdatedAt, &entity.DeletedAt)
		if err != nil {
			logger.Error().Stack().Err(err).Msg("failed scan to postgresql.GetTableList")
			return nil, DecodeDatabaseError(err)
		}

		entities = append(entities, entity)
	}

	return entities, nil
}

func (db *RDBOperation) GetTableByID(logger zerolog.Logger, ctx context.Context, id int64) (*entity.Table, error) {
	var entity entity.Table

	err := db.db.QueryRow(ctx, queryGetTableByID, id).
		Scan(&entity.ID, &entity.Name, &entity.UpdatedAt, &entity.DeletedAt)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to postgresql.GetTableByID")
		return nil, DecodeDatabaseError(stderr.New(errors.ErrGetPlayer))
	}

	return &entity, nil
}


func (db *RWDBOperation) CreateTable(logger zerolog.Logger, ctx context.Context, entity entity.Table) (*int64, error) {
	var id int64

	err := db.db.QueryRow(ctx, queryCreateTable, entity.Name).Scan(&id)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to postgresql.CreateTable")
		return nil, DecodeDatabaseError(err)
	}

	return &id, nil
}

func (db *RWDBOperation) UpdateTable(logger zerolog.Logger, ctx context.Context, entity entity.Table) error {
	res, err := db.db.Exec(ctx, queryUpdateTable, entity.ID, entity.Name)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to postgresql.UpdateTable")
		return DecodeDatabaseError(err)
	}

	if res.RowsAffected() == 0 {
		err = pgx.ErrNoRows
		logger.Error().Stack().Err(err).Msg("failed find to postgresql.UpdateTable")
		return DecodeDatabaseError(err)
	}

	return nil
}

func (db *RWDBOperation) DeleteTable(logger zerolog.Logger, ctx context.Context, id int64) error {
	res, err := db.db.Exec(ctx, queryDeleteTable, id)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to postgresql.DeleteTable")
		return DecodeDatabaseError(err)
	}

	if res.RowsAffected() == 0 {
		err = pgx.ErrNoRows
		logger.Error().Stack().Err(err).Msg("failed find to postgresql.DeleteTable")
		return DecodeDatabaseError(err)
	}

	return nil
}
