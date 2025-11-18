package abstract

import (
	"context"

	"github.com/rs/zerolog"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/config"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/database/postgresql/tx"
)

type IGameR interface {
	GetGame(logger zerolog.Logger, ctx context.Context, gameID int) (entities.GetGameResponse, error)
	GetGameById(logger zerolog.Logger, ctx context.Context, gameID int, tx tx.ITx) (entities.GameLeagueTournament, error)
	FindGames(logger zerolog.Logger, ctx context.Context, request entities.FindGameRequest) ([]entities.FindGame, error)
	GetGamesYears(logger zerolog.Logger, ctx context.Context) (entities.GetGamesYearsResponse, error)
	GetGameList(logger zerolog.Logger, ctx context.Context, request entities.GetGameListRequest) ([]entities.GameV2, error)
	GetComingGames(logger zerolog.Logger, ctx context.Context) ([]entities.ComingGame, error)
	GetFutureGames(logger zerolog.Logger, ctx context.Context, cityID int) ([]entities.ShortGame, error)
	GetTeamGames(logger zerolog.Logger, ctx context.Context, teamID int) ([]entities.TeamGame, error)
	GetTournamentGameList(logger zerolog.Logger, ctx context.Context, request *entities.GetTournamentGameList) ([]entities.TournamentGame, error)
	GetTournamentGame(logger zerolog.Logger, ctx context.Context, gameID int64, cfg *config.DBConfig) (entities.TournamentGame, error)
	GetTournamentStageGames(logger zerolog.Logger, ctx context.Context, stageId int64) ([]entities.TournamentGame, error)
	FetchPastGames(logger zerolog.Logger, ctx context.Context, teamID1, teamID2, cityID, leagueID int64, tiebreak *bool) ([]entities.GameFetch, error)
	FetchPastGamesTiebreak(logger zerolog.Logger, ctx context.Context, leagueId int64) ([]entities.GameTiebreak, error)
	GetPastGamesByPlayersTeam(logger zerolog.Logger, ctx context.Context, teamID int) ([]entities.GameShort, error)
	GetPastGamesByTeamAndLeague(logger zerolog.Logger, ctx context.Context, teamId, leagueId int) ([]entities.GameShort, error)
	GetPastGamesByPlayersTeams(logger zerolog.Logger, ctx context.Context, teamIDs []int) ([]entities.GameShort, error)
	GetTeamGamesInLeague(logger zerolog.Logger, ctx context.Context, teamId, leagueId int64, tiebreak *bool) ([]entities.Game, error)

	FetchTournamentPastGames(logger zerolog.Logger, ctx context.Context, teamId1, teamId2, cityId, tournamentId int64, tiebreak *bool) ([]entities.GameFetch, error)
	FetchTournamentPastGamesTiebreak(logger zerolog.Logger, ctx context.Context, tournamentId int64) ([]entities.GameTiebreak, error)
	FetchTournamentTeamGames(logger zerolog.Logger, ctx context.Context, tournamentId, teamId int64) ([]entities.TournamentGame, error)
}

type IGameRW interface {
	DeleteGame(logger zerolog.Logger, ctx context.Context, gameID int64, cfg *config.DBConfig) error
	UpdateGame(logger zerolog.Logger, ctx context.Context, request entities.UpdateGameRequest, tx tx.ITx) error
	UpdateFutureGame(logger zerolog.Logger, ctx context.Context, request entities.UpdateFutureGameRequest, tx tx.ITx) error
	CreateFutureGame(logger zerolog.Logger, ctx context.Context, request entities.CreateFutureGameRequest, tx tx.ITx) (int, error)
	CreateGame(logger zerolog.Logger, ctx context.Context, req entities.CreateGameRequest, tx tx.ITx) (int64, error)
	DeleteFutureGame(logger zerolog.Logger, ctx context.Context, req entities.DeleteFutureGameRequest, tx tx.ITx) error
	CreateFutureTournamentStageGames(logger zerolog.Logger, ctx context.Context, stageID, cityID int64, team1IDs, team2IDs []int64, tx tx.ITx) error
	CreateFutureTournamentGame(ctx context.Context, logger zerolog.Logger, request *entities.CreateFutureTournamentGameRequest, tx tx.ITx) (int64, error)
	UpdateFutureTournamentGame(logger zerolog.Logger, ctx context.Context, request *entities.UpdateFutureTournamentGameRequest, cfg *config.DBConfig) error
	CreatePlayedTournamentGame(ctx context.Context, logger zerolog.Logger, request *entities.CreatePlayedTournamentGameRequest, tx tx.ITx) (int64, error)
	UpdatePlayedTournamentGame(logger zerolog.Logger, ctx context.Context, req *entities.UpdatePlayedTournamentGameRequest, tx tx.ITx) error
	DeleteTournamentGames(logger zerolog.Logger, ctx context.Context, tournamentId int64, tx tx.ITx) error

	UpdateMigratedTournamentGame(ctx context.Context, logger zerolog.Logger, game entities.TournamentGame, tournamentId int64, tx tx.ITx) (int64, error)
}
