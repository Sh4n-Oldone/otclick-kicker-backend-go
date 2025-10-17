package abstract

import (
	"context"
	"github.com/rs/zerolog"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
)

type ILeagueR interface {
	FetchLeagues(logger zerolog.Logger, ctx context.Context, cityID int64, seasonID int64) ([]entities.League, error)
	GetLeagueList(logger zerolog.Logger, ctx context.Context, cityID int64) ([]entities.League, error)
	GetLeagueListByTeamId(logger zerolog.Logger, ctx context.Context, teamID int64) ([]entities.LeagueShort, error)
	GetLeaguesByPlayerID(logger zerolog.Logger, ctx context.Context, playerID int) ([]entities.PlayersLeague, error)
}

type ILeagueRW interface {
	CreateLeague(logger zerolog.Logger, ctx context.Context, request *entities.CreateLeagueRequest) (id int64, err error)
	UpdateLeague(logger zerolog.Logger, ctx context.Context, league entities.League, teams []int64) error
	DeleteLeague(logger zerolog.Logger, ctx context.Context, id int64) error
}
