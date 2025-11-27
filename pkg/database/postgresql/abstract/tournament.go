package abstract

import (
	"context"

	"github.com/rs/zerolog"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/config"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/database/postgresql/tx"
)

type ITournamentRW interface {
	CreateTournament(logger zerolog.Logger, ctx context.Context, request entities.CreateTournamentRequest, tx tx.ITx) (id int64, err error)
	CreateTournamentTx(logger zerolog.Logger, ctx context.Context, request entities.CreateTournamentRequest, tx tx.ITx) (int64, error)
	UpdateTournament(logger zerolog.Logger, ctx context.Context, request entities.UpdateTournamentRequest, tx tx.ITx) error
	UpdateTournamentTeamsLinks(logger zerolog.Logger, ctx context.Context, teamIDs []int64, tournamentId int64, tx tx.ITx) error
	AddTeamsToTournament(ctx context.Context, logger zerolog.Logger, teams []int64, tournamentId int64, tx tx.ITx) error
	DeleteTournament(logger zerolog.Logger, ctx context.Context, tournamentId int64, tx tx.ITx) error
	DeleteTournamentGamesTeamLinks(logger zerolog.Logger, ctx context.Context, tournamentId int64, tx tx.ITx) error
	CreateTournamentStage(logger zerolog.Logger, ctx context.Context, tournamentID int64, number string, tx tx.ITx) (int64, error)
	UpdateTournamentStage(ctx context.Context, logger zerolog.Logger, stage entities.NullableStage, tx tx.ITx) error
	UnlinkTeamsFromTournament(logger zerolog.Logger, ctx context.Context, tournamentID int64, tx tx.ITx) error
	DeleteTournamentStages(logger zerolog.Logger, ctx context.Context, tournamentID int64, tx tx.ITx) error

	MarkLeagueAsMigrated(ctx context.Context, logger zerolog.Logger, leagueId, tournamentId int64, tx tx.ITx) error
	UnmarkMigratedLeague(ctx context.Context, logger zerolog.Logger, tournamentId int64, tx tx.ITx) error
}

type ITournamentR interface {
	GetTournamentTypeList(logger zerolog.Logger, ctx context.Context, withDeleted bool, cfg *config.DBConfig) ([]entities.TournamentType, error)
	GetTournamentById(logger zerolog.Logger, ctx context.Context, tournamentId int64) (entities.Tournament, error)
	GetTournamentStage(logger zerolog.Logger, ctx context.Context, stageId int64) (entities.TournamentStage, error)
	GetTournamentStageList(ctx context.Context, logger zerolog.Logger, tournamentId int64) ([]entities.TournamentStage, error)
	GetTournamentList(logger zerolog.Logger, ctx context.Context, req entities.GetTournamentListRequest) ([]entities.TournamentShort, int64, error)
	GetTournamentListByPlayerID(logger zerolog.Logger, ctx context.Context, playerId int64) ([]entities.PlayersTournament, error)

	GetMigratedTournamentIds(ctx context.Context, logger zerolog.Logger) ([]int64, error)
}
