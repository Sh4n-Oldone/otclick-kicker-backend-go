package postgresql

import (
	"context"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/database/postgresql/abstract"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/config"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/database/postgresql/tx"
)

type RWDBOperationer interface {
	abstract.IUserRW
	abstract.ICityRW
	abstract.IPlaceRW
	abstract.IPlayerRW
	abstract.IBarRW
	abstract.ITableRW
	abstract.IPlaceRW
	abstract.IMatchRW
	abstract.IGameRW
	abstract.ITeamRW
	abstract.IRatingRW
	abstract.ISeasonRW
	abstract.ITournamentRW
	abstract.ILeagueRW

	BeginTx(ctx context.Context, logger zerolog.Logger) (tx.ITx, error)
}

// RDBOperationer is the interface that implemented by the RDBOperation structure.
type RDBOperationer interface {
	abstract.IRoleR
	abstract.ICityR
	abstract.IBarR
	abstract.IPlaceR
	abstract.IUserR
	abstract.ISeasonR
	abstract.ITableR
	abstract.IMatchR
	abstract.ILeagueR
	abstract.IGameR
	abstract.ITeamR
	abstract.IPlayerR
	abstract.ITournamentR
	abstract.IRatingR

	GetSuffix(logger zerolog.Logger, ctx context.Context, tx tx.ITx) (string, error)
}

type dbp struct {
	db  *pgxpool.Pool
	cfg *config.DBConfig
}

// RWDBOperation is a structure that implements the RWDBOperationer interface.
type RWDBOperation dbp

// RDBOperation is a structure that implements the RDBOperationer interface.
type RDBOperation dbp

func NewOperationer(rwConn *pgxpool.Pool, rConn *pgxpool.Pool, cfg *config.Configuration) (RWDBOperationer, RDBOperationer) {
	return &RWDBOperation{rwConn, &cfg.RWDB}, &RDBOperation{rConn, &cfg.RDB}
}

func poolOrTx(pg tx.IExecutor, tx tx.ITx) tx.IExecutor {
	if tx != nil {
		if txExec := tx.Executor(); txExec != nil {
			return txExec
		}
	}

	return pg
}

type Tx struct {
	tx     pgx.Tx
	pg     *pgxpool.Pool
	logger zerolog.Logger
}

func (db *RWDBOperation) BeginTx(ctx context.Context, logger zerolog.Logger) (tx.ITx, error) {
	tX, err := db.db.Begin(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("BeginRW")
		return nil, DecodeDatabaseError(err)
	}

	return &Tx{
		tx:     tX,
		pg:     db.db,
		logger: logger,
	}, nil
}

//func (db *RDBOperation) BeginR(ctx context.Context, logger zerolog.Logger) (tx.ITx, error) {
//	tx, err := db.db.Begin(ctx)
//	if err != nil {
//		logger.Error().Err(err).Msg("BeginR")
//		return nil, DecodeDatabaseError(err)
//	}
//
//	return &Tx{
//		tx: tx,
//		pg: db.db,
//	}, nil
//}

func (t *Tx) Commit(ctx context.Context) error {
	return t.tx.Commit(ctx)
}

func (t *Tx) Rollback(ctx context.Context) {
	logger := t.logger.With().Str("transaction", "Rollback").Logger()
	err := t.tx.Rollback(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("Failed tx.Rollback")
	}
}

func (t *Tx) Executor() tx.IExecutor {
	return t.tx
}
