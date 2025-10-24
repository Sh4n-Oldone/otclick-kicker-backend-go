package postgresql

import (
	"context"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/config"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
)

func (db *RDBOperation) GetCityList(logger zerolog.Logger, ctx context.Context, withDeleted bool) ([]entities.City, error) {
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

	var cities []entities.City

	for rows.Next() {
		var city entities.City

		err = rows.Scan(&city.ID, &city.Name, &city.Ru, &city.DeletedAt)
		if err != nil {
			logger.Error().Err(err).Msg("failed to scan city list")
			return nil, DecodeDatabaseError(err)
		}

		cities = append(cities, city)
	}

	return cities, nil
}

func (db *RDBOperation) GetCityByBarId(logger zerolog.Logger, ctx context.Context, barID int64, cfg *config.DBConfig) (entities.City, error) {
	timeout, cancel := context.WithTimeout(ctx, cfg.MaxIdleConnectionTimeout)
	defer cancel()

	var city entities.City

	const query string = `SELECT c.id, c.name, c.ru
		FROM bars b
		JOIN cities c ON c.id = b.city_id
		WHERE b.id = $1 AND c.deleted_at IS NULL AND b.deleted_at IS NULL;`

	err := db.db.QueryRow(timeout, query, barID).Scan(&city.ID, &city.Name, &city.Ru)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to postgresql.GetCityByBarId")
		return entities.City{}, DecodeDatabaseError(err)
	}

	return city, nil
}

func (db *RWDBOperation) CreateCity(logger zerolog.Logger, ctx context.Context, city entities.City) (*int64, error) {
	var id int64

	err := db.db.QueryRow(ctx, queryCreateCity, city.Name, city.Ru).Scan(&id)
	if err != nil {
		logger.Error().Err(err).Msg("failed to create City record")
		return nil, DecodeDatabaseError(err)
	}

	return &id, nil
}

func (db *RWDBOperation) UpdateCity(logger zerolog.Logger, ctx context.Context, city entities.City) error {
	res, err := db.db.Exec(ctx, queryUpdateCity, city.Name, city.Ru, city.ID)
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

func (db *RDBOperation) GetCityById(logger zerolog.Logger, ctx context.Context, cityID int64) (entities.City, error) {
	timeout, cancel := context.WithTimeout(ctx, db.cfg.MaxIdleConnectionTimeout)
	defer cancel()

	var city entities.City

	const query string = `SELECT id, name, ru, deleted_at FROM cities WHERE id = $1;`

	err := db.db.QueryRow(timeout, query, cityID).Scan(&city.ID, &city.Name, &city.Ru, &city.DeletedAt)
	if err != nil {
		logger.Error().Err(err).Msg("failed GetCityById")
		return entities.City{}, DecodeDatabaseError(err)
	}

	return city, nil
}

func (db *RDBOperation) GetUserCity(logger zerolog.Logger, ctx context.Context, userId int64) (entities.City, error) {
	timeout, cancel := context.WithTimeout(ctx, db.cfg.MaxIdleConnectionTimeout)
	defer cancel()

	var city entities.City

	const query string = `
		SELECT c.id, c.name, c.ru, c.deleted_at
		FROM cities c
		LEFT JOIN users_cities_links ucl ON c.id = ucl.city_id
		WHERE ucl.user_id = $1;`

	err := db.db.QueryRow(timeout, query, userId).Scan(&city.ID, &city.Name, &city.Ru, &city.DeletedAt)
	if err != nil {
		logger.Error().Err(err).Msg("failed GetCityById")
		return entities.City{}, DecodeDatabaseError(err)
	}

	return city, nil
}
