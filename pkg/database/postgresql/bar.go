package postgresql

import (
	"context"
	stderr "errors"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/errors"
)

func (db *RDBOperation) GetBarList(logger zerolog.Logger, ctx context.Context, cityID *int64, withDeleted bool) ([]entities.Bar, error) {

	queries := map[string]string{
		"queryGetBarList":                    queryGetBarList,
		"queryGetBarListByCityID":            queryGetBarListByCityID,
		"queryGetBarListWithDeleted":         queryGetBarListWithDeleted,
		"queryGetBarListByCityIDWithDeleted": queryGetBarListByCityIDWithDeleted,
	}

	queryName := "queryGetBarList"
	if cityID != nil {
		queryName += "ByCityID"
	}
	if withDeleted {
		queryName += "WithDeleted"
	}

	var err error
	var rows pgx.Rows
	if cityID != nil {
		rows, err = db.db.Query(ctx, queries[queryName], cityID)
	} else {
		rows, err = db.db.Query(ctx, queries[queryName])
	}

	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to postgresql.GetBarList")
		return nil, DecodeDatabaseError(err)
	}
	defer rows.Close()

	var bars []entities.Bar

	for rows.Next() {
		var city entities.City
		var bar entities.Bar

		err = rows.Scan(&bar.ID, &city.ID, &bar.Name, &bar.Description, &bar.UpdatedAt, &bar.DeletedAt)
		if err != nil {
			logger.Error().Stack().Err(err).Msg("failed scan to postgresql.GetBarList")
			return nil, DecodeDatabaseError(err)
		}

		bar.City = city

		bars = append(bars, bar)
	}

	return bars, nil
}

func (db *RDBOperation) GetBarByID(logger zerolog.Logger, ctx context.Context, id int64) (*entities.Bar, error) {
	var bar entities.Bar
	bar.City = entities.City{}

	err := db.db.QueryRow(ctx, queryGetBarByID, id).
		Scan(&bar.ID, &bar.City.ID, &bar.Name, &bar.Description, &bar.UpdatedAt, &bar.DeletedAt)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to postgresql.GetBarByID")
		return nil, DecodeDatabaseError(stderr.New(errors.ErrGetPlayer))
	}

	return &bar, nil
}

func (db *RWDBOperation) CreateBar(logger zerolog.Logger, ctx context.Context, bar entities.Bar) (*int64, error) {
	var id int64

	err := db.db.QueryRow(ctx, queryCreateBar, bar.City.ID, bar.Name, bar.Description).Scan(&id)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to postgresql.CreateBar")
		return nil, DecodeDatabaseError(err)
	}

	return &id, nil
}

func (db *RWDBOperation) UpdateBar(logger zerolog.Logger, ctx context.Context, bar entities.UpdateBarRequest) error {
	res, err := db.db.Exec(ctx, queryUpdateBar, bar.ID, bar.Name, bar.Description)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to postgresql.UpdateBar")
		return DecodeDatabaseError(err)
	}

	if res.RowsAffected() == 0 {
		err = pgx.ErrNoRows
		logger.Error().Stack().Err(err).Msg("failed find to postgresql.UpdateBar")
		return DecodeDatabaseError(err)
	}

	return nil
}

func (db *RWDBOperation) DeleteBar(logger zerolog.Logger, ctx context.Context, id int64) error {
	res, err := db.db.Exec(ctx, queryDeleteBar, id)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to postgresql.DeleteBar")
		return DecodeDatabaseError(err)
	}

	if res.RowsAffected() == 0 {
		err = pgx.ErrNoRows
		logger.Error().Stack().Err(err).Msg("failed find to postgresql.DeleteBar")
		return DecodeDatabaseError(err)
	}

	return nil
}
