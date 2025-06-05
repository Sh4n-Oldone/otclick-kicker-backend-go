package postgresql

import (
	"context"
	stderr "errors"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/entity"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/errors"
)

func (db *RDBOperation) GetBarList(logger zerolog.Logger, ctx context.Context, cityID *int64, withDeleted bool) ([]entity.Bar, error) {

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

	var entities []entity.Bar

	for rows.Next() {
		var rel entity.City
		var entity entity.Bar

		err = rows.Scan(&entity.ID, &rel.ID, &entity.Name, &entity.Description, &entity.UpdatedAt, &entity.DeletedAt)
		if err != nil {
			logger.Error().Stack().Err(err).Msg("failed scan to postgresql.GetBarList")
			return nil, DecodeDatabaseError(err)
		}

		entity.City = rel

		entities = append(entities, entity)
	}

	return entities, nil
}

func (db *RDBOperation) GetBarByID(logger zerolog.Logger, ctx context.Context, id int64) (*entity.Bar, error) {
	var rel entity.City
	var entity entity.Bar

	entity.City = rel

	err := db.db.QueryRow(ctx, queryGetBarByID, id).
		Scan(&entity.ID, &entity.City.ID, &entity.Name, &entity.Description, &entity.UpdatedAt, &entity.DeletedAt)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to postgresql.GetBarByID")
		return nil, DecodeDatabaseError(stderr.New(errors.ErrGetPlayer))
	}

	return &entity, nil
}

func (db *RWDBOperation) CreateBar(logger zerolog.Logger, ctx context.Context, entity entity.Bar) (*int64, error) {
	var id int64

	err := db.db.QueryRow(ctx, queryCreateBar, entity.City.ID, entity.Name, entity.Description).Scan(&id)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to postgresql.CreateBar")
		return nil, DecodeDatabaseError(err)
	}

	return &id, nil
}

func (db *RWDBOperation) UpdateBar(logger zerolog.Logger, ctx context.Context, entity entity.UpdateBarRequest) error {
	res, err := db.db.Exec(ctx, queryUpdateBar, entity.ID, entity.Name, entity.Description)
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
