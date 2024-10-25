package postgresql

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/entity"
)

func (db *RDBOperation) GetCityList(logger zerolog.Logger, ctx context.Context, withDeleted bool) ([]entity.City, error) {
	query := queryGetCityList
	if withDeleted {
		query = queryGetCityListWithDeleted
	}

	rows, err := db.db.Query(ctx, query)
	if err != nil {
		logger.Error().Err(err).Msg("failed to get city list")
		return nil, DecodeDatabaseError(err)
	}
	defer rows.Close()

	var cities []entity.City

	for rows.Next() {
		var city entity.City

		err = rows.Scan(&city.ID, &city.Name, &city.Ru, &city.DeletedAt)
		if err != nil {
			logger.Error().Err(err).Msg("failed to scan city list")
			return nil, DecodeDatabaseError(err)
		}

		cities = append(cities, city)
	}

	return cities, nil
}

func (db *RWDBOperation) CreateCity(logger zerolog.Logger, ctx context.Context, city entity.City) (*int64, error) {
	var id int64

	err := db.db.QueryRow(ctx, queryCreateCity, city.Name, city.Ru).Scan(&id)
	if err != nil {
		logger.Error().Err(err).Msg("failed to create City record")
		return nil, DecodeDatabaseError(err)
	}

	return &id, nil
}

func (db *RWDBOperation) UpdateCity(logger zerolog.Logger, ctx context.Context, city entity.City) error {
	res, err := db.db.Exec(ctx, queryUpdateCity, city.Name, city.Ru, city.DeletedAt, city.ID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to update City record")
		return DecodeDatabaseError(err)
	}

	if res.RowsAffected() == 0 {
		err = pgx.ErrNoRows
		logger.Error().Err(err).Msg("not found City record to update")
		return DecodeDatabaseError(err)
	}

	return nil
}

func (db *RWDBOperation) DeleteCity(logger zerolog.Logger, ctx context.Context, id int64) error {
	res, err := db.db.Exec(ctx, queryDeleteCity, id)
	if err != nil {
		logger.Error().Err(err).Msg("failed to delete City record")
		return DecodeDatabaseError(err)
	}

	if res.RowsAffected() == 0 {
		err = pgx.ErrNoRows
		logger.Error().Err(err).Msg("not found City record to delete")
		return DecodeDatabaseError(err)
	}

	return nil
}
