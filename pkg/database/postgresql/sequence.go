package postgresql

import (
	"context"
	"github.com/rs/zerolog"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/config"
)

func (db *RDBOperation) GetSuffix(logger zerolog.Logger, ctx context.Context, cfg *config.DBConfig) (string, error) {
	timeout, cancel := context.WithTimeout(ctx, cfg.MaxIdleConnectionTimeout)
	defer cancel()

	var suffix string

	const query string = "SELECT * FROM get_next_name_suffix()"

	err := db.db.QueryRow(timeout, query).Scan(&suffix)
	if err != nil {
		logger.Error().Err(err).Msg("failed GetSuffix")
		return suffix, DecodeDatabaseError(err)
	}

	return suffix, nil
}
