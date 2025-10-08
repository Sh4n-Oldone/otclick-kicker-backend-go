package abstract

import (
	"context"

	"github.com/rs/zerolog"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
)

type ISeasonR interface {
	GetSeasonList(logger zerolog.Logger, ctx context.Context, req *entities.GetSeasonListRequest) ([]entities.Season, error)
}

type ISeasonRW interface {
	CreateSeason(logger zerolog.Logger, ctx context.Context, entity entities.Season) (id *int64, err error)
	UpdateSeason(logger zerolog.Logger, ctx context.Context, entity entities.UpdateSeasonRequest) error
	DeleteSeason(logger zerolog.Logger, ctx context.Context, id int64) (bool, error)
}
