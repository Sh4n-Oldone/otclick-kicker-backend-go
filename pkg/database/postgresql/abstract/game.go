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
	FindGames(logger zerolog.Logger, ctx context.Context, request entities.FindGameRequest) ([]entities.FindGame, error)
	GetGamesYears(logger zerolog.Logger, ctx context.Context) (entities.GetGamesYearsResponse, error)
	GetGameList(logger zerolog.Logger, ctx context.Context, request entities.GetGameListRequest) ([]entities.GameV2, error)
	GetComingGames(logger zerolog.Logger, ctx context.Context) ([]entities.ComingGame, error)
	GetFutureGames(logger zerolog.Logger, ctx context.Context, cityID int) ([]entities.ShortGame, error)
	GetTeamGames(logger zerolog.Logger, ctx context.Context, teamID int) ([]entities.TeamGame, error)
	GetTournamentGameList(logger zerolog.Logger, ctx context.Context, request *entities.GetTournamentGameList) ([]entities.TournamentGame, error)
	GetTournamentGame(logger zerolog.Logger, ctx context.Context, gameID int64, cfg *config.DBConfig) (entities.TournamentGame, error)
	GetTournamentStageGames(logger zerolog.Logger, ctx context.Context, stageId int64, cfg *config.DBConfig) ([]entities.TournamentGame, error)
	FetchPastGames(logger zerolog.Logger, ctx context.Context, teamID1, teamID2, cityID, leagueID int64, tiebreak *bool) ([]entities.GameFetch, error)
	FetchPastGamesTiebreak(logger zerolog.Logger, ctx context.Context, leagueId int64) ([]entities.GameTiebreak, error)
	GetPastGamesByPlayersTeam(logger zerolog.Logger, ctx context.Context, teamID int) ([]entities.GameShort, error)
	GetPastGamesByTeamAndLeague(logger zerolog.Logger, ctx context.Context, teamId, leagueId int) ([]entities.GameShort, error)
	GetPastGamesByPlayersTeams(logger zerolog.Logger, ctx context.Context, teamIDs []int) ([]entities.GameShort, error)
	GetTeamGamesInLeague(logger zerolog.Logger, ctx context.Context, teamId, leagueId int64, tiebreak *bool) ([]entities.Game, error)
}

type IGameRW interface {
	DeleteGame(logger zerolog.Logger, ctx context.Context, gameID int64, cfg *config.DBConfig) error
	UpdateGame(logger zerolog.Logger, ctx context.Context, request entities.UpdateGameRequest, rates []entities.Rating) error
	UpdateFutureGame(logger zerolog.Logger, ctx context.Context, request entities.UpdateFutureGameRequest) error
	CreateFutureGame(logger zerolog.Logger, ctx context.Context, request entities.CreateFutureGameRequest) (int, error)
	CreateGameWithRating(logger zerolog.Logger, ctx context.Context, request entities.CreateGameRequest, rates map[int]int, operator *string, leagueID int64) (entities.CreateGameResponse, error)
	DeleteFutureGame(logger zerolog.Logger, ctx context.Context, gameID int64) error
	CreateFutureTournamentStageGames(logger zerolog.Logger, ctx context.Context, stageID, cityID int64, team1IDs, team2IDs []int64, tx tx.ITx) error
	CreateFutureTournamentGame(logger zerolog.Logger, ctx context.Context, request *entities.CreateFutureTournamentGameRequest, cfg *config.DBConfig) (int64, error)
	UpdateFutureTournamentGame(logger zerolog.Logger, ctx context.Context, request *entities.UpdateFutureTournamentGameRequest, cfg *config.DBConfig) error
	CreatePlayedTournamentGame(logger zerolog.Logger, ctx context.Context, request *entities.CreatePlayedTournamentGameRequest, cfg *config.DBConfig) (int64, error)
	UpdatePlayedTournamentGame(logger zerolog.Logger, ctx context.Context, req *entities.UpdatePlayedTournamentGameRequest, cfg *config.DBConfig) error
	DeleteTournamentGames(logger zerolog.Logger, ctx context.Context, tournamentId int64, tx tx.ITx) error
}
