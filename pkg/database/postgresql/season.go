package postgresql

import (
	"context"
	stderr "errors"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/entity"
)

func (db *RDBOperation) GetSeasonList(logger zerolog.Logger, ctx context.Context) ([]entity.Season, error) {
	rows, err := db.db.Query(ctx, queryGetSeasonList)
	if err != nil {
		logger.Error().Err(err).Msg("failed to GetSeasonList")
		return nil, DecodeDatabaseError(err)
	}
	defer rows.Close()

	var seasons []entity.Season

	for rows.Next() {
		var season entity.Season

		err = rows.Scan(&season.ID, &season.Name, &season.Description)
		if err != nil {
			logger.Error().Err(err).Msg("failed to scan season record")
			return nil, DecodeDatabaseError(err)
		}

		seasons = append(seasons, season)
	}

	return seasons, nil
}

func (db *RWDBOperation) CreateSeason(logger zerolog.Logger, ctx context.Context, entity entity.Season) (*int64, error) {
	var id int64

	err := db.db.QueryRow(ctx, queryCreateSeason, entity.Name, entity.Description).Scan(&id)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to postgresql.CreateSeason")
		return nil, DecodeDatabaseError(err)
	}

	return &id, nil
}

func (db *RWDBOperation) UpdateSeason(logger zerolog.Logger, ctx context.Context, entity entity.UpdateSeasonRequest) error {
	res, err := db.db.Exec(ctx, queryUpdateSeason, entity.ID, entity.Name, entity.Description)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to postgresql.UpdateSeason")
		return DecodeDatabaseError(err)
	}

	if res.RowsAffected() == 0 {
		err = pgx.ErrNoRows
		logger.Error().Stack().Err(err).Msg("failed find to postgresql.UpdateSeason")
		return DecodeDatabaseError(err)
	}

	return nil
}

func (db *RWDBOperation) DeleteSeason(logger zerolog.Logger, ctx context.Context, id int64) (bool, error) {
	tx, err := db.db.Begin(ctx)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to delete season record")
		return false, DecodeDatabaseError(err)
	}

	// убираем сезон из таблицы лиг
	_, err = db.db.Exec(ctx, queryDeleteLeagueSeasonId, id)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to delete season id from leagues")
		_ = tx.Rollback(ctx)
		return false, DecodeDatabaseError(err)
	}

	result, err := db.db.Exec(ctx, queryDeleteSeason, id)
	if err != nil {
		logger.Error().Err(err).Msg("failed to delete season record")
		return false, DecodeDatabaseError(err)
	}

	if result.RowsAffected() == 0 {
		logger.Error().Err(err).Msg("failed to get affected rows")
		return false, stderr.New("Failed to delete season, it does not exist")
	}

	err = tx.Commit(ctx)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to delete season record")
		_ = tx.Rollback(ctx)
		return false, DecodeDatabaseError(err)
	}

	return true, nil
}
