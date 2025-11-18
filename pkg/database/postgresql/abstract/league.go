package abstract

import (
	"context"
	"github.com/rs/zerolog"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/database/postgresql/tx"
)

type ILeagueR interface {
	FetchLeagues(logger zerolog.Logger, ctx context.Context, cityID int64, seasonID int64) ([]entities.League, error)
	GetLeagueList(logger zerolog.Logger, ctx context.Context, cityID int64) ([]entities.League, error)
	GetLeagueListByTeamId(logger zerolog.Logger, ctx context.Context, teamID int64) ([]entities.LeagueShort, error)
	GetLeaguesByPlayerID(logger zerolog.Logger, ctx context.Context, playerID int) ([]entities.PlayersLeague, error)
	GetLeagueById(logger zerolog.Logger, ctx context.Context, leagueID int64, tx tx.ITx) (entities.League, error)

	GetLeagueListToMigrate(ctx context.Context, logger zerolog.Logger) ([]entities.League, error)
	GetLeagueGamesToMigrate(ctx context.Context, logger zerolog.Logger, leagueId int64) ([]entities.TournamentGame, error)
	GetExtraPointsByLeagueIdToMigrate(ctx context.Context, logger zerolog.Logger, leagueId int64) ([]entities.ExtraPoints, error)
	GetMigratedLeagueIds(ctx context.Context, logger zerolog.Logger) ([]int64, error)
}

type ILeagueRW interface {
	CreateLeague(logger zerolog.Logger, ctx context.Context, request *entities.CreateLeagueRequest) (id int64, err error)
	UpdateLeague(logger zerolog.Logger, ctx context.Context, league entities.League, teams []int64) error
	DeleteLeague(logger zerolog.Logger, ctx context.Context, id int64) error
}
