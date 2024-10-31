package postgresql

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/entity"
)

func (db *RDBOperation) GetRatingList(logger zerolog.Logger, ctx context.Context, leagueID *int64, playerID *int64) ([]entity.Rating, error) {

	queries := map[string]string {
		"queryGetRatingList": queryGetRatingList,
		"queryGetRatingListByPlayerID": queryGetRatingListByPlayerID,
		"queryGetRatingListByLeagueID": queryGetRatingListByLeagueID,
	}

	var queryAttr int64
	queryName := "queryGetRatingList"
	if playerID != nil {
		queryAttr = *playerID
		queryName += "ByPlayerID"
	}
	if leagueID != nil {
		queryAttr = *leagueID
		queryName += "ByLeagueID"
	}

	var err error
	var rows pgx.Rows

	if queryAttr > 0 {
		rows, err = db.db.Query(ctx, queries[queryName], queryAttr)	
	} else {
		rows, err = db.db.Query(ctx, queries[queryName])	
	}

	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to postgresql.GetRatingList")
		return nil, DecodeDatabaseError(err)
	}
	defer rows.Close()

	var entities []entity.Rating

	for rows.Next() {
		var entity entity.Rating

		err = rows.Scan(&entity.PlayerID, &entity.LeagueID, &entity.Value)
		if err != nil {
			logger.Error().Stack().Err(err).Msg("failed scan to postgresql.GetRatingList")
			return nil, DecodeDatabaseError(err)
		}

		entities = append(entities, entity)
	}

	return entities, nil
}

func (db *RDBOperation) GetRatingByPlayerIDAndByLeagueID(logger zerolog.Logger, ctx context.Context, playerID, leagueID int64) (int64, error) {
	var rate int64
	
	err := db.db.QueryRow(ctx, queryGetRatingByPlayerIDAndByLeagueID, playerID, leagueID).Scan(&rate)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to postgresql.GetRatingByPlayerIDAndByLeagueID")
		return 0, DecodeDatabaseError(err)
	}

	return rate, nil
}

func (db *RWDBOperation) CreateRating(logger zerolog.Logger, ctx context.Context, entity entity.Rating, operator *string) error {
	query := queryCreateRating
	if operator != nil && *operator == "insertIgnore" {
		query = queryCreateRatingInsertIgnore
	}
	_, err := db.db.Exec(ctx, query, entity.PlayerID, entity.LeagueID, entity.Value)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to postgresql.CreateRating")
		return DecodeDatabaseError(err)
	}

	return nil
}

func (db *RWDBOperation) UpdateRating(logger zerolog.Logger, ctx context.Context, entity entity.Rating) error {
	res, err := db.db.Exec(ctx, queryUpdateRating, entity.PlayerID, entity.LeagueID, entity.Value)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to postgresql.UpdateRating")
		return DecodeDatabaseError(err)
	}

	if res.RowsAffected() == 0 {
		err = pgx.ErrNoRows
		logger.Error().Stack().Err(err).Msg("failed find to postgresql.UpdateRating")
		return DecodeDatabaseError(err)
	}

	return nil
}
