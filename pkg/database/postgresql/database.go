package postgresql

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

type RWDBOperationer interface {
}

// RDBOperationer is the interface that implemented by the RDBOperation structure.
type RDBOperationer interface {
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
