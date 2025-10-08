package abstract

import (
	"context"

	"github.com/rs/zerolog"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/config"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
)

type IPlaceR interface {
	GetPlaceList(logger zerolog.Logger, ctx context.Context, barID, tableID, cityID *int64, withDelete bool) ([]entities.Place, error)
	GetPlaceByID(logger zerolog.Logger, ctx context.Context, id int64, cfg *config.DBConfig) (*entities.Place, error)
}

type IPlaceRW interface {
	CreatePlace(logger zerolog.Logger, ctx context.Context, entity entities.CreatePlaceRequest, cfg *config.DBConfig) (id *int64, err error)
	UpdatePlace(logger zerolog.Logger, ctx context.Context, entity entities.UpdatePlaceRequest, cfg *config.DBConfig) error
	DeletePlace(logger zerolog.Logger, ctx context.Context, id int64, cfg *config.DBConfig) error
}
