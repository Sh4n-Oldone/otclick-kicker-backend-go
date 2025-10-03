package postgresql

import (
	"context"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/database/postgresql/tx"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/config"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
)

func (db *RDBOperation) GetRatingList(logger zerolog.Logger, ctx context.Context, leagueID *int64, playerID *int64) ([]entities.Rating, error) {

	queries := map[string]string{
		"queryGetRatingList":           queryGetRatingList,
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

	var ratings []entities.Rating

	for rows.Next() {
		var rating entities.Rating

		err = rows.Scan(&rating.PlayerID, &rating.LeagueID, &rating.Value)
		if err != nil {
			logger.Error().Stack().Err(err).Msg("failed scan to postgresql.GetRatingList")
			return nil, DecodeDatabaseError(err)
		}

		ratings = append(ratings, rating)
	}

	return ratings, nil
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

func (db *RWDBOperation) CreateRating(logger zerolog.Logger, ctx context.Context, rating entities.Rating, operator *string) error {
	query := queryCreateRating
	if operator != nil && *operator == "insertIgnore" {
		query = queryCreateRatingInsertIgnore
	}
	_, err := db.db.Exec(ctx, query, rating.PlayerID, rating.LeagueID, rating.Value)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to postgresql.CreateRating")
		return DecodeDatabaseError(err)
	}

	return nil
}

func (db *RWDBOperation) CreateTournamentRating(logger zerolog.Logger, ctx context.Context, playerID, value, tournamentID int64, cfg *config.DBConfig) error {
	timeout, cancel := context.WithTimeout(ctx, cfg.MaxIdleConnectionTimeout)
	defer cancel()

	const query string = `
		INSERT INTO tournament_rating (player_id, tournament_id, value) 
		VALUES ($1, $2, $3)  ON CONFLICT (player_id, tournament_id) DO UPDATE SET value = $3;`

	_, err := db.db.Exec(timeout, query, playerID, tournamentID, value)
	if err != nil {
		logger.Error().Err(err).Msg("failed to postgresql.CreateTournamentRating")
		return DecodeDatabaseError(err)
	}

	return nil
}

func (db *RWDBOperation) UpdateRating(logger zerolog.Logger, ctx context.Context, rating entities.Rating) error {
	res, err := db.db.Exec(ctx, queryUpdateRating, rating.PlayerID, rating.LeagueID, rating.Value)
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

func (db *RDBOperation) GetPlayerRatingByTournamentId(logger zerolog.Logger, ctx context.Context, playerID, tournamentID int64, cfg *config.DBConfig) (int64, error) {
	timeout, cancel := context.WithTimeout(ctx, cfg.MaxIdleConnectionTimeout)
	defer cancel()

	var rate int64

	const query string = `SELECT value FROM tournament_rating WHERE player_id = $1 AND tournament_id = $2;`

	err := db.db.QueryRow(timeout, query, playerID, tournamentID).Scan(&rate)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to postgresql.GetPlayerRatingByTournamentId")
		return 0, DecodeDatabaseError(err)
	}

	return rate, nil
}

func (db *RWDBOperation) DeleteTournamentRating(logger zerolog.Logger, ctx context.Context, tournamentId int64, tx tx.ITx) error {
	timeout, cancel := context.WithTimeout(ctx, db.cfg.MaxIdleConnectionTimeout)
	defer cancel()

	const query string = `DELETE FROM tournament_rating WHERE tournament_id = $1;`

	_, err := poolOrTx(db.db, tx).Exec(timeout, query, tournamentId)
	if err != nil {
		logger.Error().Err(err).Msg("failed to postgresql.DeleteTournamentRating")
		return DecodeDatabaseError(err)
	}

	return nil
}
