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
	DeleteMatch(logger zerolog.Logger, ctx context.Context, id int64) (bool, error)

	CreatePlayer(logger zerolog.Logger, ctx context.Context, player entities.CreatePlayerRequest) (int, error)
	DeletePlayer(logger zerolog.Logger, ctx context.Context, playerID int) error
	RecoverPlayer(logger zerolog.Logger, ctx context.Context, playerID int) error
	UpdatePlayer(logger zerolog.Logger, ctx context.Context, playerData entities.UpdatePlayerRequest) error

	CreateBar(logger zerolog.Logger, ctx context.Context, entity entity.Bar) (id *int64, err error)
	UpdateBar(logger zerolog.Logger, ctx context.Context, entity entity.UpdateBarRequest) error
	DeleteBar(logger zerolog.Logger, ctx context.Context, id int64) error

	CreateTable(logger zerolog.Logger, ctx context.Context, entity entity.Table) (id *int64, err error)
	UpdateTable(logger zerolog.Logger, ctx context.Context, entity entity.Table) error
	DeleteTable(logger zerolog.Logger, ctx context.Context, id int64) error

	CreatePlace(logger zerolog.Logger, ctx context.Context, entity entity.Place) (id *int64, err error)
	UpdatePlace(logger zerolog.Logger, ctx context.Context, entity entity.Place) error
	DeletePlace(logger zerolog.Logger, ctx context.Context, id int64) error
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
	UpdateFutureGame(logger zerolog.Logger, ctx context.Context, request entity.UpdateFutureGameRequest) error
	CreateFutureGame(logger zerolog.Logger, ctx context.Context, request entity.CreateFutureGameRequest) (int, error)

	CreateRating(logger zerolog.Logger, ctx context.Context, entity entity.Rating, operatior *string) error
	UpdateRating(logger zerolog.Logger, ctx context.Context, entity entity.Rating) error

	RewriteMatchesAndPlayerRatings(logger zerolog.Logger, ctx context.Context, matches []entity.Match, ratings map[int64]entity.Rating) error
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

	GetTableList(logger zerolog.Logger, ctx context.Context, withDelete bool) ([]entity.Table, error)

	GetBarList(logger zerolog.Logger, ctx context.Context, cityID *int64, withDelete bool) ([]entity.Bar, error)
	GetBarByID(logger zerolog.Logger, ctx context.Context, id int64) (*entity.Bar, error)

	GetPlaceList(logger zerolog.Logger, ctx context.Context, barID, tableID *int64, withDelete bool) ([]entity.Place, error)
	GetPlaceByID(logger zerolog.Logger, ctx context.Context, id int64) (*entity.Place, error)
	GetLeagueList(logger zerolog.Logger, ctx context.Context, cityID int64) ([]entity.League, error)

	GetGame(logger zerolog.Logger, ctx context.Context, gameID int) (entities.GetGameResponse, error)
	FindGames(logger zerolog.Logger, ctx context.Context, request entities.FindGameRequest) ([]entities.FindGame, error)
	GetGamesYears(logger zerolog.Logger, ctx context.Context) (entity.GetGamesYearsResponse, error)
	GetComingGames(logger zerolog.Logger, ctx context.Context) ([]entities.ComingGame, error)
	GetFutureGames(logger zerolog.Logger, ctx context.Context, cityID int) ([]entity.ShortGame, error)
	GetTeamGames(logger zerolog.Logger, ctx context.Context, teamID int) ([]entity.TeamGame, error)
	GetTeam(logger zerolog.Logger, ctx context.Context, teamID int64) (entity.GetTeamResponse, error)
	GetTeams(logger zerolog.Logger, ctx context.Context, cityId int64, onlyFree bool) ([]entity.TeamShort, error)
	GetTeamsByCity(logger zerolog.Logger, ctx context.Context, onlyFree bool, cityID int64) ([]entity.TeamShort, error)
	GetTeamsByLeague(logger zerolog.Logger, ctx context.Context, leagueID int64) ([]entity.TeamByLeague, error)

	FetchLeagues(logger zerolog.Logger, ctx context.Context, cityID int64) ([]entity.League, error)
	FetchTeams(logger zerolog.Logger, ctx context.Context, leagueID int64) ([]entity.Team, error)
	FetchGames(logger zerolog.Logger, ctx context.Context, teamID1, teamID2, cityID, year int64) ([]entities.ComingGame, error)
	FetchMatches(logger zerolog.Logger, ctx context.Context, gameID int64) ([]entities.Match, error)
	TeamsHaveNoGames(logger zerolog.Logger, ctx context.Context, teams []entity.Team, year int64) (bool, error)

	GetRatingByPlayerIDAndByLeagueID(logger zerolog.Logger, ctx context.Context, playerID, leagueID int64) (int64, error)

	GetMatchListByGameID(ctx context.Context, logger zerolog.Logger, gameID int) ([]entities.Match, error)

	GetPlayerIDsByLeagueID(logger zerolog.Logger, ctx context.Context, leagueID int64) ([]int64, error)
	GetMatchListByLeagueID(logger zerolog.Logger, ctx context.Context, leagueID int64) ([]entity.Match, error)
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
