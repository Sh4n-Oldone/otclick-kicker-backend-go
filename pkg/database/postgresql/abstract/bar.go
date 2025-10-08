package abstract

import (
	"context"

	"github.com/rs/zerolog"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/config"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
)

type IBarR interface {
	GetBarList(logger zerolog.Logger, ctx context.Context, cityID *int64, withDelete bool) ([]entities.Bar, error)
	GetBarByID(logger zerolog.Logger, ctx context.Context, id int64, cfg *config.DBConfig) (*entities.Bar, error)
}

type IBarRW interface {
	CreateBar(logger zerolog.Logger, ctx context.Context, entity entities.Bar) (id *int64, err error)
	UpdateBar(logger zerolog.Logger, ctx context.Context, entity entities.UpdateBarRequest) error
	DeleteBar(logger zerolog.Logger, ctx context.Context, id int64) error
}
