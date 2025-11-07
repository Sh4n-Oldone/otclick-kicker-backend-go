package abstract

import (
	"context"

	"github.com/rs/zerolog"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/database/postgresql/tx"
)

type IRatingR interface {
}

type IRatingRW interface {
	CreateRatingUpdateOnConflict(logger zerolog.Logger, ctx context.Context, entity entities.Rating, tx tx.ITx) error
	UpdateRating(logger zerolog.Logger, ctx context.Context, entity entities.Rating) error
	CreateTournamentRating(logger zerolog.Logger, ctx context.Context, playerId, value, tournamentId int64, tx tx.ITx) error
	DeleteTournamentRating(logger zerolog.Logger, ctx context.Context, tournamentId int64, tx tx.ITx) error
}
