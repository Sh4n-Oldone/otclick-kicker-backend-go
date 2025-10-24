package abstract

import (
	"context"

	"github.com/rs/zerolog"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/config"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/database/postgresql/tx"
)

type ITeamRW interface {
	CreateTeam(logger zerolog.Logger, ctx context.Context, team entities.CreateTeamRequest, tx tx.ITx) (id int64, err error)
	UpdateTeam(logger zerolog.Logger, ctx context.Context, team entities.UpdateTeamRequest, cfg *config.DBConfig) (bool, error)
	DeleteTeam(logger zerolog.Logger, ctx context.Context, id int64, cfg *config.DBConfig) (bool, error)
	AddPlayerIntoTeam(logger zerolog.Logger, ctx context.Context, playerID, teamID int64, tx tx.ITx) (bool, error)
	RemovePlayerFromTeam(logger zerolog.Logger, ctx context.Context, playerID, teamID int64, cfg *config.DBConfig) (bool, error)

	CreateExtraPoints(logger zerolog.Logger, ctx context.Context, req *entities.CreateExtraPointsRequest) (int64, error)
	UpdateExtraPoints(logger zerolog.Logger, ctx context.Context, req *entities.UpdateExtraPointsRequest) (bool, error)
	DeleteExtraPoints(logger zerolog.Logger, ctx context.Context, extraPointsId int64) (bool, error)
}

type ITeamR interface {
	GetTeam(logger zerolog.Logger, ctx context.Context, teamID int64, cfg *config.DBConfig) (entities.GetTeamResponse, error)
	GetTeams(logger zerolog.Logger, ctx context.Context, cityId int64, onlyFree bool) ([]entities.TeamShort, error)
	GetTeamsByCity(logger zerolog.Logger, ctx context.Context, onlyFree bool, cityID int64) ([]entities.TeamShort, error)
	GetTeamsByLeague(logger zerolog.Logger, ctx context.Context, leagueID int64) ([]entities.TeamByLeague, error)
	GetTournamentTeamList(logger zerolog.Logger, ctx context.Context, tournamentID int64, cfg *config.DBConfig) ([]entities.TournamentTeam, error)
	FetchTeams(logger zerolog.Logger, ctx context.Context, leagueID int64) ([]entities.Team, error)
	TeamsHaveNoGames(logger zerolog.Logger, ctx context.Context, teams []entities.Team, seasonID int64) (bool, error)
	GetTeamsByPlayerID(logger zerolog.Logger, ctx context.Context, playerID int, tx tx.ITx) ([]entities.TeamItem, error)
	GetTeamById(logger zerolog.Logger, ctx context.Context, teamID int64, tx tx.ITx) (entities.TeamV2, error)

	GetTeamExtraPointsCount(logger zerolog.Logger, ctx context.Context, teamID, leagueID int64) (int64, error)
	GetExtraPointsListByTeamAndLeagueId(logger zerolog.Logger, ctx context.Context, teamId, leagueId int64) ([]entities.ExtraPoints, error)
	GetExtraPointsById(logger zerolog.Logger, ctx context.Context, extraPointsId int64) (entities.ExtraPoints, error)

	FetchTeamsByTournament(logger zerolog.Logger, ctx context.Context, tournamentId, tournamentType *int64) ([]entities.Team, error)
	TeamsHaveNoGamesTournament(logger zerolog.Logger, ctx context.Context, teams []entities.Team, seasonId int64) (bool, error)
	GetTournamentTeamExtraPointsCount(logger zerolog.Logger, ctx context.Context, teamId, tournamentId int64) (int64, error)
}
