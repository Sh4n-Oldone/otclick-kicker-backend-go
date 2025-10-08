package postgresql

import (
	"context"
	"github.com/rs/zerolog"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/database/postgresql/tx"
)

// GetSuffix deprecated
func (db *RDBOperation) GetSuffix(logger zerolog.Logger, ctx context.Context, tx tx.ITx) (string, error) {
	timeout, cancel := context.WithTimeout(ctx, db.cfg.MaxIdleConnectionTimeout)
	defer cancel()

	var suffix string

	const query string = "SELECT * FROM get_next_name_suffix()"

	err := poolOrTx(db.db, tx).QueryRow(timeout, query).Scan(&suffix)
	if err != nil {
		logger.Error().Err(err).Msg("failed GetSuffix")
		return suffix, DecodeDatabaseError(err)
	}

	return suffix, nil
}
