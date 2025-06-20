package postgresql

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
)

type RWDBOperationer interface {
	CreateCity(logger zerolog.Logger, ctx context.Context, city entities.City) (id *int64, err error)
	UpdateCity(logger zerolog.Logger, ctx context.Context, city entities.City) error
	DeleteCity(logger zerolog.Logger, ctx context.Context, id int64) error

	CreateUser(logger zerolog.Logger, ctx context.Context, user entities.User) (id *int64, err error)
	UpdateUser(logger zerolog.Logger, ctx context.Context, user entities.User) error

	CreateMatch(logger zerolog.Logger, ctx context.Context, match entities.Match) (id int64, err error)
	UpdateMatch(logger zerolog.Logger, ctx context.Context, match entities.Match) error
	DeleteMatch(logger zerolog.Logger, ctx context.Context, id int64) (bool, error)

	CreatePlayer(logger zerolog.Logger, ctx context.Context, player entities.CreatePlayerRequest) (int, error)
	DeletePlayer(logger zerolog.Logger, ctx context.Context, playerID int) error
	RecoverPlayer(logger zerolog.Logger, ctx context.Context, playerID int) error
	UpdatePlayer(logger zerolog.Logger, ctx context.Context, playerData entities.UpdatePlayerRequest) error

	CreateBar(logger zerolog.Logger, ctx context.Context, entity entities.Bar) (id *int64, err error)
	UpdateBar(logger zerolog.Logger, ctx context.Context, entity entities.UpdateBarRequest) error
	DeleteBar(logger zerolog.Logger, ctx context.Context, id int64) error

	CreateTable(logger zerolog.Logger, ctx context.Context, entity entities.Table) (id *int64, err error)
	UpdateTable(logger zerolog.Logger, ctx context.Context, entity entities.Table) error
	DeleteTable(logger zerolog.Logger, ctx context.Context, id int64) error

	CreatePlace(logger zerolog.Logger, ctx context.Context, entity entities.Place) (id *int64, err error)
	UpdatePlace(logger zerolog.Logger, ctx context.Context, entity entities.Place) error
	DeletePlace(logger zerolog.Logger, ctx context.Context, id int64) error
	CreateTeam(logger zerolog.Logger, ctx context.Context, team entities.CreateTeamRequest) (id int64, err error)
	UpdateTeam(logger zerolog.Logger, ctx context.Context, team entities.UpdateTeamRequest) (bool, error)
	DeleteTeam(logger zerolog.Logger, ctx context.Context, id int64) (bool, error)
	AddPlayerIntoTeam(logger zerolog.Logger, ctx context.Context, playerID, teamID int64) (bool, error)
	RemovePlayerFromTeam(logger zerolog.Logger, ctx context.Context, playerID, teamID int64) (bool, error)
	CreateLeague(logger zerolog.Logger, ctx context.Context, league entities.League, teams []int64) (id int64, err error)
	UpdateLeague(logger zerolog.Logger, ctx context.Context, league entities.League, teams []int64) error
	DeleteLeague(logger zerolog.Logger, ctx context.Context, id int64) error

	DeleteGame(logger zerolog.Logger, ctx context.Context, gameID int) error
	UpdateGame(logger zerolog.Logger, ctx context.Context, request entities.UpdateGameRequest, rates []entities.Rating) error
	UpdateFutureGame(logger zerolog.Logger, ctx context.Context, request entities.UpdateFutureGameRequest) error
	CreateFutureGame(logger zerolog.Logger, ctx context.Context, request entities.CreateFutureGameRequest) (int, error)
	CreateGameWithRating(logger zerolog.Logger, ctx context.Context, request entities.CreateGameRequest, rates map[int]int, operator *string, leagueID int64) (entities.CreateGameResponse, error)
	DeleteFutureGame(logger zerolog.Logger, ctx context.Context, gameID int64) error

	CreateRating(logger zerolog.Logger, ctx context.Context, entity entities.Rating, operatior *string) error
	UpdateRating(logger zerolog.Logger, ctx context.Context, entity entities.Rating) error

	RewriteMatchesAndPlayerRatings(logger zerolog.Logger, ctx context.Context, matches []entities.Match, ratings map[int64]entities.Rating) error

	CreateExtraPoints(logger zerolog.Logger, ctx context.Context, req *entities.CreateExtraPointsRequest) (int64, error)
	UpdateExtraPoints(logger zerolog.Logger, ctx context.Context, req *entities.UpdateExtraPointsRequest) (bool, error)
	DeleteExtraPoints(logger zerolog.Logger, ctx context.Context, extraPointsId int64) (bool, error)

	CreateSeason(logger zerolog.Logger, ctx context.Context, entity entities.Season) (id *int64, err error)
	UpdateSeason(logger zerolog.Logger, ctx context.Context, entity entities.UpdateSeasonRequest) error
	DeleteSeason(logger zerolog.Logger, ctx context.Context, id int64) (bool, error)
}

// RDBOperationer is the interface that implemented by the RDBOperation structure.
type RDBOperationer interface {
	GetRoleList(logger zerolog.Logger, ctx context.Context) ([]entities.Role, error)
	GetRole(logger zerolog.Logger, ctx context.Context, id *int64, name *string) (*entities.Role, error)

	GetCityList(logger zerolog.Logger, ctx context.Context, withDelete bool) ([]entities.City, error)

	GetUser(logger zerolog.Logger, ctx context.Context, id *int64, email *string) (*entities.User, error)

	GetPlayerByID(logger zerolog.Logger, ctx context.Context, playerID int) (entities.Player, error)
	GetPlayersByTeamID(logger zerolog.Logger, ctx context.Context, teamID int) ([]entities.Player, error)

	GetPastMatchesByPlayerID(logger zerolog.Logger, ctx context.Context, playerID int) ([]entities.MatchV2, error)
	GetLeaguesByPlayerID(logger zerolog.Logger, ctx context.Context, playerID int) ([]entities.PlayersLeague, error)

	GetTeamsByPlayerID(logger zerolog.Logger, ctx context.Context, playerID int) ([]entities.TeamItem, error)

	GetPastGamesByPlayersTeam(logger zerolog.Logger, ctx context.Context, teamID int) ([]entities.GameShort, error)

	GetPastGamesByPlayersTeams(logger zerolog.Logger, ctx context.Context, teamIDs []int) ([]entities.GameShort, error)

	FindPlayers(logger zerolog.Logger, ctx context.Context, player entities.FindPlayersRequest) ([]entities.Player, error)

	GetTableList(logger zerolog.Logger, ctx context.Context, withDelete bool) ([]entities.Table, error)

	GetBarList(logger zerolog.Logger, ctx context.Context, cityID *int64, withDelete bool) ([]entities.Bar, error)
	GetBarByID(logger zerolog.Logger, ctx context.Context, id int64) (*entities.Bar, error)

	GetPlaceList(logger zerolog.Logger, ctx context.Context, barID, tableID, cityID *int64, withDelete bool) ([]entities.Place, error)
	GetPlaceByID(logger zerolog.Logger, ctx context.Context, id int64) (*entities.Place, error)
	GetLeagueList(logger zerolog.Logger, ctx context.Context, cityID int64) ([]entities.League, error)

	GetGame(logger zerolog.Logger, ctx context.Context, gameID int) (entities.GetGameResponse, error)
	FindGames(logger zerolog.Logger, ctx context.Context, request entities.FindGameRequest) ([]entities.FindGame, error)
	GetGamesYears(logger zerolog.Logger, ctx context.Context) (entities.GetGamesYearsResponse, error)
	GetGameList(logger zerolog.Logger, ctx context.Context, request entities.GetGameListRequest) ([]entities.GameV2, error)
	GetComingGames(logger zerolog.Logger, ctx context.Context) ([]entities.ComingGame, error)
	GetFutureGames(logger zerolog.Logger, ctx context.Context, cityID int) ([]entities.ShortGame, error)
	GetTeamGames(logger zerolog.Logger, ctx context.Context, teamID int) ([]entities.TeamGame, error)
	GetTeam(logger zerolog.Logger, ctx context.Context, teamID int64) (entities.GetTeamResponse, error)
	GetTeams(logger zerolog.Logger, ctx context.Context, cityId int64, onlyFree bool) ([]entities.TeamShort, error)
	GetTeamsByCity(logger zerolog.Logger, ctx context.Context, onlyFree bool, cityID int64) ([]entities.TeamShort, error)
	GetTeamsByLeague(logger zerolog.Logger, ctx context.Context, leagueID int64) ([]entities.TeamByLeague, error)

	FetchLeagues(logger zerolog.Logger, ctx context.Context, cityID int64, seasonID int64) ([]entities.League, error)
	FetchTeams(logger zerolog.Logger, ctx context.Context, leagueID int64) ([]entities.Team, error)
	FetchPastGames(logger zerolog.Logger, ctx context.Context, teamID1, teamID2, cityID, leagueID int64, tiebreak *bool) ([]entities.GameFetch, error)
	FetchPastGamesTiebreak(logger zerolog.Logger, ctx context.Context, leagueId int64) ([]entities.GameTiebreak, error)
	FetchMatches(logger zerolog.Logger, ctx context.Context, gameID int64) ([]entities.ShortMatch, error)
	TeamsHaveNoGames(logger zerolog.Logger, ctx context.Context, teams []entities.Team, seasonID int64) (bool, error)

	GetRatingByPlayerIDAndByLeagueID(logger zerolog.Logger, ctx context.Context, playerID, leagueID int64) (int64, error)

	GetMatchListByGameID(ctx context.Context, logger zerolog.Logger, gameID int) ([]entities.MatchV2, error)

	GetPlayerIDsByLeagueID(logger zerolog.Logger, ctx context.Context, leagueID int64) ([]int64, error)
	GetMatchListByLeagueID(logger zerolog.Logger, ctx context.Context, leagueID int64) ([]entities.Match, error)

	GetTeamExtraPointsCount(logger zerolog.Logger, ctx context.Context, teamID, leagueID int64) (int64, error)
	GetExtraPointsListByTeamAndLeagueId(logger zerolog.Logger, ctx context.Context, teamId, leagueId int64) ([]entities.ExtraPoints, error)
	GetExtraPointsById(logger zerolog.Logger, ctx context.Context, extraPointsId int64) (entities.ExtraPoints, error)

	GetSeasonList(logger zerolog.Logger, ctx context.Context) ([]entities.Season, error)

	GetTeamById(logger zerolog.Logger, ctx context.Context, teamID int64) (entities.TeamV2, error)
	GetLeagueListByTeamId(logger zerolog.Logger, ctx context.Context, teamID int64) ([]entities.LeagueShort, error)
	GetCaptainByTeamId(logger zerolog.Logger, ctx context.Context, teamID int64) (entities.User, error)
	GetTeamGamesInLeague(logger zerolog.Logger, ctx context.Context, teamId, leagueId int64, tiebreak *bool) ([]entities.Game, error)
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
