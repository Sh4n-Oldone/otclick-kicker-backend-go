package abstract

import (
	"context"

	"github.com/rs/zerolog"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/config"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
)

type ICityR interface {
	GetCityList(logger zerolog.Logger, ctx context.Context, withDelete bool) ([]entities.City, error)
	GetCityByBarId(logger zerolog.Logger, ctx context.Context, barID int64, cfg *config.DBConfig) (entities.City, error)
	GetCityById(logger zerolog.Logger, ctx context.Context, cityID int64) (entities.City, error)
	GetUserCity(logger zerolog.Logger, ctx context.Context, cityID int64) (entities.City, error)
}

type ICityRW interface {
	CreateCity(logger zerolog.Logger, ctx context.Context, city entities.City) (id *int64, err error)
	UpdateCity(logger zerolog.Logger, ctx context.Context, city entities.City) error
	DeleteCity(logger zerolog.Logger, ctx context.Context, id int64) error
}
