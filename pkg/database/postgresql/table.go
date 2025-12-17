package postgresql

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
)

func (db *RDBOperation) GetTableList(logger zerolog.Logger, ctx context.Context, withDeleted bool) ([]entities.Table, error) {
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

	var tables []entities.Table

	for rows.Next() {
		var table entities.Table

		err = rows.Scan(&table.ID, &table.Name, &table.UpdatedAt, &table.DeletedAt)
		if err != nil {
			logger.Error().Stack().Err(err).Msg("failed scan to postgresql.GetTableList")
			return nil, DecodeDatabaseError(err)
		}

		tables = append(tables, table)
	}

	return tables, nil
}

func (db *RDBOperation) GetTableByID(logger zerolog.Logger, ctx context.Context, id int64) (*entities.Table, error) {
	var table entities.Table

	err := db.db.QueryRow(ctx, queryGetTableByID, id).
		Scan(&table.ID, &table.Name, &table.UpdatedAt, &table.DeletedAt)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to postgresql.GetTableByID")
		return nil, DecodeDatabaseError(err)
	}

	return &table, nil
}

func (db *RWDBOperation) CreateTable(logger zerolog.Logger, ctx context.Context, table entities.Table) (*int64, error) {
	var id int64

	err := db.db.QueryRow(ctx, queryCreateTable, table.Name).Scan(&id)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to postgresql.CreateTable")
		return nil, DecodeDatabaseError(err)
	}

	return &id, nil
}

func (db *RWDBOperation) UpdateTable(logger zerolog.Logger, ctx context.Context, table entities.Table) error {
	res, err := db.db.Exec(ctx, queryUpdateTable, table.ID, table.Name)
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
