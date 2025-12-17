package abstract

import (
	"context"

	"github.com/rs/zerolog"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
)

type ITableRW interface {
	CreateTable(logger zerolog.Logger, ctx context.Context, entity entities.Table) (id *int64, err error)
	UpdateTable(logger zerolog.Logger, ctx context.Context, entity entities.Table) error
	DeleteTable(logger zerolog.Logger, ctx context.Context, id int64) error
}

type ITableR interface {
	GetTableList(logger zerolog.Logger, ctx context.Context, withDelete bool) ([]entities.Table, error)
	GetTableByID(logger zerolog.Logger, ctx context.Context, id int64) (*entities.Table, error)
}
