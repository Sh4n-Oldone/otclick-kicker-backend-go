package abstract

import (
	"context"

	"github.com/rs/zerolog"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/database/postgresql/tx"
)

type IRatingR interface {
	GetRatingsByLeagueIdToMigrate(ctx context.Context, logger zerolog.Logger, leagueId int64) ([]entities.Rating, error)
}

type IRatingRW interface {
	CreateRatingUpdateOnConflict(logger zerolog.Logger, ctx context.Context, entity entities.Rating, tx tx.ITx) error
	UpdateRating(logger zerolog.Logger, ctx context.Context, entity entities.Rating) error
	CreateTournamentRating(ctx context.Context, logger zerolog.Logger, playerId, value, tournamentId int64, tx tx.ITx) error
	DeleteTournamentRating(logger zerolog.Logger, ctx context.Context, tournamentId int64, tx tx.ITx) error
}
