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
	AddTeamsToTournamentTx(logger zerolog.Logger, ctx context.Context, teams []int64, tournamentId int64, tx tx.ITx) error
	DeleteTournament(logger zerolog.Logger, ctx context.Context, tournamentId int64, tx tx.ITx) error
	DeleteTournamentGamesTeamLinks(logger zerolog.Logger, ctx context.Context, tournamentId int64, tx tx.ITx) error
	CreateTournamentStage(logger zerolog.Logger, ctx context.Context, tournamentID int64, number string, tx tx.ITx) (int64, error)
	UpdateTournamentStage(logger zerolog.Logger, ctx context.Context, stage entities.NullableStage) error
	UnlinkTeamsFromTournament(logger zerolog.Logger, ctx context.Context, tournamentID int64, tx tx.ITx) error
	DeleteTournamentStages(logger zerolog.Logger, ctx context.Context, tournamentID int64, tx tx.ITx) error
}

type ITournamentR interface {
	GetTournamentTypeList(logger zerolog.Logger, ctx context.Context, withDeleted bool, cfg *config.DBConfig) ([]entities.TournamentType, error)
	GetTournamentById(logger zerolog.Logger, ctx context.Context, tournamentId int64, cfg *config.DBConfig) (entities.Tournament, error)
	GetTournamentStage(logger zerolog.Logger, ctx context.Context, stageId int64, cfg *config.DBConfig) (entities.TournamentStage, error)
	GetTournamentStageList(logger zerolog.Logger, ctx context.Context, tournamentId int64) ([]entities.TournamentStage, error)
	GetTournamentList(logger zerolog.Logger, ctx context.Context, req entities.GetTournamentListRequest) ([]entities.TournamentShort, int64, error)
	GetTournamentListByPlayerID(logger zerolog.Logger, ctx context.Context, playerId int64) ([]entities.PlayersTournament, error)
}
