package abstract

import (
	"context"

	"github.com/rs/zerolog"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/config"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/database/postgresql/tx"
)

type IPlayerR interface {
	GetRatingByPlayerIDAndByLeagueID(logger zerolog.Logger, ctx context.Context, playerID, leagueID int64) (int64, error)
	GetPlayerRatingByTournamentId(logger zerolog.Logger, ctx context.Context, playerID, tournamentID int64) (int64, error)
	GetTournamentPlayers(ctx context.Context, logger zerolog.Logger, tournamentId int64, withDeleted bool, tx tx.ITx) ([]entities.TournamentPlayer, error)
	GetPlayerIDsByLeagueID(logger zerolog.Logger, ctx context.Context, leagueID int64) ([]int64, error)
	GetPlayerIdsByTournamentId(logger zerolog.Logger, ctx context.Context, tournamentId int64) ([]int64, error)
	GetPlayerByID(logger zerolog.Logger, ctx context.Context, playerID int, tx tx.ITx) (entities.Player, error)
	GetPlayersByTeamID(logger zerolog.Logger, ctx context.Context, teamID int, tx tx.ITx) ([]entities.Player, error)
	FindPlayers(logger zerolog.Logger, ctx context.Context, player entities.FindPlayersRequest) ([]entities.Player, error)
	FindPlayersV2(logger zerolog.Logger, ctx context.Context, player entities.FindPlayersRequest) ([]entities.Player, error)
}

type IPlayerRW interface {
	CreatePlayer(logger zerolog.Logger, ctx context.Context, player entities.CreatePlayerRequest, dbConfig *config.DBConfig) (int, error)
	DeletePlayer(logger zerolog.Logger, ctx context.Context, playerID int) error
	RecoverPlayer(logger zerolog.Logger, ctx context.Context, playerID int) error
	UpdatePlayer(logger zerolog.Logger, ctx context.Context, playerData entities.UpdatePlayerRequest) error
}
