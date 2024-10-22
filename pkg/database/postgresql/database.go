package postgresql

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/entity"
)

type RWDBOperationer interface {
	CreateCity(logger zerolog.Logger, ctx context.Context, city entity.City) (id int64, err error)
	UpdateCity(logger zerolog.Logger, ctx context.Context, city entity.City) error
	DeleteCity(logger zerolog.Logger, ctx context.Context, id int64) error

	CreateMatch(logger zerolog.Logger, ctx context.Context, match entity.Match) (id int64, err error)
	UpdateMatch(logger zerolog.Logger, ctx context.Context, match entity.Match) error
	DeleteMatch(logger zerolog.Logger, ctx context.Context, id int64) error
}

// RDBOperationer is the interface that implemented by the RDBOperation structure.
type RDBOperationer interface {
	GetCityList(logger zerolog.Logger, ctx context.Context, withDelete bool) ([]entity.City, error)
}

type dbp struct {
	db *pgxpool.Pool
}

// RWDBOperation is a structure that implements the RWDBOperationer interface.
type RWDBOperation dbp

// RDBOperation is a structure that implements the RDBOperationer interface.
type RDBOperation dbp

func NewOperationer(rwConn *pgxpool.Pool, rConn *pgxpool.Pool) (RWDBOperationer, RDBOperationer) {
	return &RWDBOperation{rwConn}, &RDBOperation{rConn}
}
