package postgresql

import (
	"context"
	stderr "errors"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/errors"
)

func (db *RDBOperation) GetPlaceList(logger zerolog.Logger, ctx context.Context, barID, tableID, cityID *int64, withDeleted bool) ([]entities.Place, error) {
	queries := map[string]string{
		"queryGetPlaceList":                             queryGetPlaceList,
		"queryGetPlaceListWithDeleted":                  queryGetPlaceListWithDeleted,
		"queryGetPlaceListByBarID":                      queryGetPlaceListByBarID,
		"queryGetPlaceListByBarIDWithDeleted":           queryGetPlaceListByBarIDWithDeleted,
		"queryGetPlaceListByTableID":                    queryGetPlaceListByTableID,
		"queryGetPlaceListByTableIDWithDeleted":         queryGetPlaceListByTableIDWithDeleted,
		"queryGetPlaceListByBarIDByTableID":             queryGetPlaceListByBarIDByTableID,
		"queryGetPlaceListByBarIDByTableIDWithDeleted":  queryGetPlaceListByBarIDByTableIDWithDeleted,
		"queryGetPlaceListByCityID":                     queryGetPlaceListByCityID,
		"queryGetPlaceListByCityIDWithDeleted":          queryGetPlaceListByCityIDWithDeleted,
		"queryGetPlaceListByCityIDByTableID":            queryGetPlaceListByCityIDByTableID,
		"queryGetPlaceListByCityIDByTableIDWithDeleted": queryGetPlaceListByCityIDByTableIDWithDeleted,
	}

	var queryAttr *int64
	queryName := "queryGetPlaceList"
	if barID != nil {
		queryAttr = barID
		queryName += "ByBarID"
	}
	if cityID != nil {
		queryAttr = cityID
		queryName += "ByCityID"
	}
	if tableID != nil {
		queryAttr = tableID
		queryName += "ByTableID"
	}
	if withDeleted {
		queryName += "WithDeleted"
	}

	var err error
	var rows pgx.Rows
	if barID != nil && tableID != nil {
		rows, err = db.db.Query(ctx, queries[queryName], barID, tableID)
	} else if cityID != nil && tableID != nil {
		rows, err = db.db.Query(ctx, queries[queryName], cityID, tableID)
	} else if barID != nil || cityID != nil || tableID != nil {
		rows, err = db.db.Query(ctx, queries[queryName], queryAttr)
	} else {
		rows, err = db.db.Query(ctx, queries[queryName])
	}
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to postgresql.GetPlaceList")
		return nil, stderr.New(errors.ErrGetPlaceList)
	}
	defer rows.Close()

	var places []entities.Place

	for rows.Next() {
		var place entities.Place
		place.Bar = entities.Bar{}
		place.Table = entities.Table{}

		err = rows.Scan(&place.ID, &place.Bar.ID, &place.Bar.Name, &place.Table.ID, &place.Table.Name, &place.UpdatedAt, &place.DeletedAt)
		if err != nil {
			logger.Error().Stack().Err(err).Msg("failed scan to postgresql.GetPlaceList")
			return nil, stderr.New(errors.ErrGetPlaceList)
		}

		places = append(places, place)
	}

	return places, nil
}

func (db *RDBOperation) GetPlaceByID(logger zerolog.Logger, ctx context.Context, id int64) (*entities.Place, error) {
	var place entities.Place
	place.Bar = entities.Bar{}
	place.Table = entities.Table{}

	err := db.db.QueryRow(ctx, queryGetPlaceByID, id).
		Scan(&place.ID, &place.Bar.ID, &place.Table.ID, &place.UpdatedAt, &place.DeletedAt)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to postgresql.GetBarByID")
		return nil, DecodeDatabaseError(stderr.New(errors.ErrGetPlayer))
	}

	return &place, nil
}

func (db *RWDBOperation) CreatePlace(logger zerolog.Logger, ctx context.Context, place entities.Place) (*int64, error) {
	var id int64

	err := db.db.QueryRow(ctx, queryCreatePlace, place.Bar.ID, place.Table.ID).Scan(&id)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to postgresql.CreatePlace")
		return nil, DecodeDatabaseError(err)
	}

	return &id, nil
}

func (db *RWDBOperation) UpdatePlace(logger zerolog.Logger, ctx context.Context, place entities.Place) error {
	res, err := db.db.Exec(ctx, queryUpdatePlace, place.ID, place.Bar.ID, place.Table.ID)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to postgresql.UpdatePlace")
		return DecodeDatabaseError(err)
	}

	if res.RowsAffected() == 0 {
		err = pgx.ErrNoRows
		logger.Error().Stack().Err(err).Msg("failed find to postgresql.UpdatePlace")
		return DecodeDatabaseError(err)
	}

	return nil
}

func (db *RWDBOperation) DeletePlace(logger zerolog.Logger, ctx context.Context, id int64) error {
	res, err := db.db.Exec(ctx, queryDeletePlace, id)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to postgresql.DeletePlace")
		return DecodeDatabaseError(err)
	}

	if res.RowsAffected() == 0 {
		err = pgx.ErrNoRows
		logger.Error().Stack().Err(err).Msg("failed find to postgresql.DeletePlace")
		return DecodeDatabaseError(err)
	}

	return nil
}
