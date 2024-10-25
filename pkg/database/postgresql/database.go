package postgresql

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/entity"
)

type RWDBOperationer interface {
	CreateCity(logger zerolog.Logger, ctx context.Context, city entity.City) (id *int64, err error)
	UpdateCity(logger zerolog.Logger, ctx context.Context, city entity.City) error
	DeleteCity(logger zerolog.Logger, ctx context.Context, id int64) error
	CreateUser(logger zerolog.Logger, ctx context.Context, user entity.User) (id *int64, err error)
	UpdateUser(logger zerolog.Logger, ctx context.Context, user entity.User) error
	CreateMatch(logger zerolog.Logger, ctx context.Context, match entity.Match) (id int64, err error)
	UpdateMatch(logger zerolog.Logger, ctx context.Context, match entity.Match) error
	DeleteMatch(logger zerolog.Logger, ctx context.Context, id int64) error
	CreatePlayer(logger zerolog.Logger, ctx context.Context, player entities.CreatePlayerRequest) (int, error)
	DeletePlayer(logger zerolog.Logger, ctx context.Context, playerID int) error
	UpdatePlayer(logger zerolog.Logger, ctx context.Context, playerData entities.UpdatePlayerRequest) error
	CreateTeam(logger zerolog.Logger, ctx context.Context, team entity.CreateTeamRequest) (id int64, err error)
	UpdateTeam(logger zerolog.Logger, ctx context.Context, team entity.UpdateTeamRequest) (bool, error)
	DeleteTeam(logger zerolog.Logger, ctx context.Context, id int64) (bool, error)
	AddPlayerIntoTeam(logger zerolog.Logger, ctx context.Context, playerID, teamID int64) (bool, error)
	RemovePlayerFromTeam(logger zerolog.Logger, ctx context.Context, playerID, teamID int64) (bool, error)
	CreateLeague(logger zerolog.Logger, ctx context.Context, league entity.League, teams []int64) (id int64, err error)
	UpdateLeague(logger zerolog.Logger, ctx context.Context, league entity.League, teams []int64) error
	DeleteLeague(logger zerolog.Logger, ctx context.Context, id int64) error
	CreatePlayedGame(logger zerolog.Logger, ctx context.Context, request entities.CreateGameRequest) (entities.CreateGameResponse, error)
	DeleteGame(logger zerolog.Logger, ctx context.Context, gameID int) error
	UpdateGame(logger zerolog.Logger, ctx context.Context, request entities.UpdateGameRequest) error
	UpdateFutureGame(logger zerolog.Logger, ctx context.Context, request entities.UpdateFutureGameRequest) error
}

// RDBOperationer is the interface that implemented by the RDBOperation structure.
type RDBOperationer interface {
	GetRoleList(logger zerolog.Logger, ctx context.Context) ([]entity.Role, error)
	GetRole(logger zerolog.Logger, ctx context.Context, id *int64, name *string) (*entity.Role, error)
	GetCityList(logger zerolog.Logger, ctx context.Context, withDelete bool) ([]entity.City, error)
	GetUser(logger zerolog.Logger, ctx context.Context, id *int64, email *string) (*entity.User, error)
	GetPlayerByID(logger zerolog.Logger, ctx context.Context, playerID int) (entities.Player, error)
	GetPlayersByTeamID(logger zerolog.Logger, ctx context.Context, teamID int) ([]entities.Player, error)
	GetMatchesByPlayerID(logger zerolog.Logger, ctx context.Context, playerID int) ([]entities.Match, error)
	GetLeaguesByPlayerID(logger zerolog.Logger, ctx context.Context, playerID int) ([]entities.PlayersLeague, error)
	GetGamesByPlayersTeam(logger zerolog.Logger, ctx context.Context, teamID int) ([]entities.Game, error)
	FindPlayers(logger zerolog.Logger, ctx context.Context, player entities.FindPlayersRequest) ([]entities.Player, error)
	GetLeagueList(logger zerolog.Logger, ctx context.Context, cityID int64) ([]entity.League, error)
	GetGame(logger zerolog.Logger, ctx context.Context, gameID int) (entities.GetGameResponse, error)
	FindGames(logger zerolog.Logger, ctx context.Context, request entities.FindGameRequest) ([]entities.FindGame, error)
	GetGamesYears(logger zerolog.Logger, ctx context.Context) (entity.GetGamesYearsResponse, error)
}

type dbp struct {
	db *pgxpool.Pool
}

// RWDBOperation is a structure that implements the RWDBOperationer interface.
type RWDBOperation dbp

// RDBOperation is a structure that implements the RDBOperationer interface.
type RDBOperation dbp

func NewOperationer(rwConn *pgxpool.Pool, rConn *pgxpool.Pool) (RWDBOperationer, RDBOperationer) {
	return &RWDBOperation{rwConn}, &RDBOperation{rConn}
}
