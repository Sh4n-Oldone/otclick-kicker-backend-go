package postgresql

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/config"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/database/postgresql/tx"
)

type RWDBOperationer interface {
	CreateCity(logger zerolog.Logger, ctx context.Context, city entities.City) (id *int64, err error)
	UpdateCity(logger zerolog.Logger, ctx context.Context, city entities.City) error
	DeleteCity(logger zerolog.Logger, ctx context.Context, id int64) error

	CreateUser(logger zerolog.Logger, ctx context.Context, user entities.User) (id *int64, err error)
	UpdateUser(logger zerolog.Logger, ctx context.Context, user entities.User) error
	CreateTournamentMaster(logger zerolog.Logger, ctx context.Context, master entities.TournamentMaster, config *config.DBConfig) (id int64, err error)
	UpdateTournamentMaster(logger zerolog.Logger, ctx context.Context, master entities.UpdateTournamentMasterRequest, cfg *config.DBConfig) error
	DeleteTournamentMaster(logger zerolog.Logger, ctx context.Context, userId int64, cfg *config.DBConfig) error

	CreateMatch(logger zerolog.Logger, ctx context.Context, match entities.Match) (id int64, err error)
	CreateGameMatch(logger zerolog.Logger, ctx context.Context, match entities.GamesMatch, gameId int64, cfg *config.DBConfig) (id int64, err error)
	UpdateMatch(logger zerolog.Logger, ctx context.Context, match entities.Match) error
	DeleteMatch(logger zerolog.Logger, ctx context.Context, id int64) (bool, error)
	DeleteOldGameMatches(logger zerolog.Logger, ctx context.Context, gameId int64, newMatchesIds []int, cfg *config.DBConfig) error
	UpdateOldMatch(logger zerolog.Logger, ctx context.Context, gameId int64, match *entities.NewMatch, cfg *config.DBConfig) error
	CreateNewMatch(logger zerolog.Logger, ctx context.Context, gameId int64, match *entities.NewMatch, cfg *config.DBConfig) error
	DeleteGameMatches(logger zerolog.Logger, ctx context.Context, gameId int64, cfg *config.DBConfig) error
	DeleteTournamentMatches(logger zerolog.Logger, ctx context.Context, tournamentId int64, tx tx.ITx) error

	CreatePlayer(logger zerolog.Logger, ctx context.Context, player entities.CreatePlayerRequest, dbConfig *config.DBConfig) (int, error)
	DeletePlayer(logger zerolog.Logger, ctx context.Context, playerID int) error
	RecoverPlayer(logger zerolog.Logger, ctx context.Context, playerID int) error
	UpdatePlayer(logger zerolog.Logger, ctx context.Context, playerData entities.UpdatePlayerRequest) error

	CreateBar(logger zerolog.Logger, ctx context.Context, entity entities.Bar) (id *int64, err error)
	UpdateBar(logger zerolog.Logger, ctx context.Context, entity entities.UpdateBarRequest) error
	DeleteBar(logger zerolog.Logger, ctx context.Context, id int64) error

	CreateTable(logger zerolog.Logger, ctx context.Context, entity entities.Table) (id *int64, err error)
	UpdateTable(logger zerolog.Logger, ctx context.Context, entity entities.Table) error
	DeleteTable(logger zerolog.Logger, ctx context.Context, id int64) error

	CreatePlace(logger zerolog.Logger, ctx context.Context, entity entities.CreatePlaceRequest, cfg *config.DBConfig) (id *int64, err error)
	UpdatePlace(logger zerolog.Logger, ctx context.Context, entity entities.UpdatePlaceRequest, cfg *config.DBConfig) error
	DeletePlace(logger zerolog.Logger, ctx context.Context, id int64, cfg *config.DBConfig) error

	CreateTeam(logger zerolog.Logger, ctx context.Context, team entities.CreateTeamRequest, tx tx.ITx) (id int64, err error)
	UpdateTeam(logger zerolog.Logger, ctx context.Context, team entities.UpdateTeamRequest, cfg *config.DBConfig) (bool, error)
	DeleteTeam(logger zerolog.Logger, ctx context.Context, id int64, cfg *config.DBConfig) (bool, error)
	AddPlayerIntoTeam(logger zerolog.Logger, ctx context.Context, playerID, teamID int64, tx tx.ITx) (bool, error)
	RemovePlayerFromTeam(logger zerolog.Logger, ctx context.Context, playerID, teamID int64, cfg *config.DBConfig) (bool, error)

	CreateLeague(logger zerolog.Logger, ctx context.Context, league entities.League, teams []int64) (id int64, err error)
	UpdateLeague(logger zerolog.Logger, ctx context.Context, league entities.League, teams []int64) error
	DeleteLeague(logger zerolog.Logger, ctx context.Context, id int64) error

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

	CreateRating(logger zerolog.Logger, ctx context.Context, entity entities.Rating, operator *string) error
	UpdateRating(logger zerolog.Logger, ctx context.Context, entity entities.Rating) error
	CreateTournamentRating(logger zerolog.Logger, ctx context.Context, playerId, value, tournamentId int64, cfg *config.DBConfig) error
	DeleteTournamentRating(logger zerolog.Logger, ctx context.Context, tournamentId int64, tx tx.ITx) error

	RewriteMatchesAndPlayerRatings(logger zerolog.Logger, ctx context.Context, matches []entities.Match, ratings map[int64]entities.Rating) error

	CreateExtraPoints(logger zerolog.Logger, ctx context.Context, req *entities.CreateExtraPointsRequest) (int64, error)
	UpdateExtraPoints(logger zerolog.Logger, ctx context.Context, req *entities.UpdateExtraPointsRequest) (bool, error)
	DeleteExtraPoints(logger zerolog.Logger, ctx context.Context, extraPointsId int64) (bool, error)

	CreateSeason(logger zerolog.Logger, ctx context.Context, entity entities.Season) (id *int64, err error)
	UpdateSeason(logger zerolog.Logger, ctx context.Context, entity entities.UpdateSeasonRequest) error
	DeleteSeason(logger zerolog.Logger, ctx context.Context, id int64) (bool, error)

	CreateTournament(logger zerolog.Logger, ctx context.Context, request entities.CreateTournamentRequest, tx tx.ITx) (id int64, err error)
	CreateTournamentTx(logger zerolog.Logger, ctx context.Context, request entities.CreateTournamentRequest, tx tx.ITx) (int64, error)
	UpdateTournament(logger zerolog.Logger, ctx context.Context, request entities.UpdateTournamentRequest, tx tx.ITx) error
	UpdateTournamentTeamsLinks(logger zerolog.Logger, ctx context.Context, teamIDs []int64, tournamentId int64, tx tx.ITx) error
	AddTeamsToTournamentTx(logger zerolog.Logger, ctx context.Context, teams []int64, tournamentId int64, tx tx.ITx) error
	DeleteTournament(logger zerolog.Logger, ctx context.Context, tournamentId int64, tx tx.ITx) error
	DeleteTournamentGamesTeamLinks(logger zerolog.Logger, ctx context.Context, tournamentId int64, tx tx.ITx) error
	CreateTournamentStage(logger zerolog.Logger, ctx context.Context, tournamentID int64, tx tx.ITx) (int64, error)
	CreateTournamentStageTx(logger zerolog.Logger, ctx context.Context, tournamentID int64, tx tx.ITx) (int64, error)
	UpdateTournamentStage(logger zerolog.Logger, ctx context.Context, stage entities.NullableStage) error
	UnlinkTeamsFromTournament(logger zerolog.Logger, ctx context.Context, tournamentID int64, tx tx.ITx) error
	DeleteTournamentStages(logger zerolog.Logger, ctx context.Context, tournamentID int64, tx tx.ITx) error

	BeginTx(ctx context.Context, logger zerolog.Logger) (tx.ITx, error)
}

// RDBOperationer is the interface that implemented by the RDBOperation structure.
type RDBOperationer interface {
	GetRoleList(logger zerolog.Logger, ctx context.Context) ([]entities.Role, error)
	GetRole(logger zerolog.Logger, ctx context.Context, id *int64, name *string) (*entities.Role, error)

	GetCityList(logger zerolog.Logger, ctx context.Context, withDelete bool) ([]entities.City, error)
	GetCityByBarId(logger zerolog.Logger, ctx context.Context, barID int64, cfg *config.DBConfig) (entities.City, error)
	GetCityById(logger zerolog.Logger, ctx context.Context, cityID int64) (entities.City, error)
	GetUserCity(logger zerolog.Logger, ctx context.Context, cityID int64) (entities.City, error)

	GetUser(logger zerolog.Logger, ctx context.Context, id *int64, email *string, cfg *config.DBConfig) (*entities.User, error)
	GetTournamentMasterByUserId(logger zerolog.Logger, ctx context.Context, userId int64, cfg *config.DBConfig) (entities.TournamentMaster, error)
	GetTournamentMasterListByCityId(logger zerolog.Logger, ctx context.Context, cityID int64, cfg *config.DBConfig) ([]entities.TournamentMaster, error)
	GetTournamentMasterList(logger zerolog.Logger, ctx context.Context, cfg *config.DBConfig) ([]entities.TournamentMaster, error)

	GetPlayerByID(logger zerolog.Logger, ctx context.Context, playerID int, tx tx.ITx) (entities.Player, error)
	GetPlayersByTeamID(logger zerolog.Logger, ctx context.Context, teamID int, tx tx.ITx) ([]entities.Player, error)

	GetPastMatchesByPlayerID(logger zerolog.Logger, ctx context.Context, playerID int) ([]entities.MatchV2, error)
	GetLeaguesByPlayerID(logger zerolog.Logger, ctx context.Context, playerID int) ([]entities.PlayersLeague, error)

	GetTeamsByPlayerID(logger zerolog.Logger, ctx context.Context, playerID int, tx tx.ITx) ([]entities.TeamItem, error)

	GetPastGamesByPlayersTeam(logger zerolog.Logger, ctx context.Context, teamID int) ([]entities.GameShort, error)
	GetPastGamesByTeamAndLeague(logger zerolog.Logger, ctx context.Context, teamId, leagueId int) ([]entities.GameShort, error)

	GetPastGamesByPlayersTeams(logger zerolog.Logger, ctx context.Context, teamIDs []int) ([]entities.GameShort, error)

	FindPlayers(logger zerolog.Logger, ctx context.Context, player entities.FindPlayersRequest) ([]entities.Player, error)

	GetTableList(logger zerolog.Logger, ctx context.Context, withDelete bool) ([]entities.Table, error)

	GetBarList(logger zerolog.Logger, ctx context.Context, cityID *int64, withDelete bool) ([]entities.Bar, error)
	GetBarByID(logger zerolog.Logger, ctx context.Context, id int64, cfg *config.DBConfig) (*entities.Bar, error)

	GetPlaceList(logger zerolog.Logger, ctx context.Context, barID, tableID, cityID *int64, withDelete bool) ([]entities.Place, error)
	GetPlaceByID(logger zerolog.Logger, ctx context.Context, id int64, cfg *config.DBConfig) (*entities.Place, error)
	GetLeagueList(logger zerolog.Logger, ctx context.Context, cityID int64) ([]entities.League, error)

	GetGame(logger zerolog.Logger, ctx context.Context, gameID int) (entities.GetGameResponse, error)
	FindGames(logger zerolog.Logger, ctx context.Context, request entities.FindGameRequest) ([]entities.FindGame, error)
	GetGamesYears(logger zerolog.Logger, ctx context.Context) (entities.GetGamesYearsResponse, error)
	GetGameList(logger zerolog.Logger, ctx context.Context, request entities.GetGameListRequest) ([]entities.GameV2, error)
	GetComingGames(logger zerolog.Logger, ctx context.Context) ([]entities.ComingGame, error)
	GetFutureGames(logger zerolog.Logger, ctx context.Context, cityID int) ([]entities.ShortGame, error)
	GetTeamGames(logger zerolog.Logger, ctx context.Context, teamID int) ([]entities.TeamGame, error)
	GetTournamentGameList(logger zerolog.Logger, ctx context.Context, tournamentID int64, cfg *config.DBConfig) ([]entities.TournamentGame, error)
	GetTournamentGame(logger zerolog.Logger, ctx context.Context, gameID int64, cfg *config.DBConfig) (entities.TournamentGame, error)
	GetTournamentStageGames(logger zerolog.Logger, ctx context.Context, stageId int64, cfg *config.DBConfig) ([]entities.TournamentGame, error)

	GetTeam(logger zerolog.Logger, ctx context.Context, teamID int64, cfg *config.DBConfig) (entities.GetTeamResponse, error)
	GetTeams(logger zerolog.Logger, ctx context.Context, cityId int64, onlyFree bool) ([]entities.TeamShort, error)
	GetTeamsByCity(logger zerolog.Logger, ctx context.Context, onlyFree bool, cityID int64) ([]entities.TeamShort, error)
	GetTeamsByLeague(logger zerolog.Logger, ctx context.Context, leagueID int64) ([]entities.TeamByLeague, error)
	GetTournamentTeamList(logger zerolog.Logger, ctx context.Context, tournamentID int64, cfg *config.DBConfig) ([]entities.TournamentTeam, error)

	FetchLeagues(logger zerolog.Logger, ctx context.Context, cityID int64, seasonID int64) ([]entities.League, error)
	FetchTeams(logger zerolog.Logger, ctx context.Context, leagueID int64) ([]entities.Team, error)
	FetchPastGames(logger zerolog.Logger, ctx context.Context, teamID1, teamID2, cityID, leagueID int64, tiebreak *bool) ([]entities.GameFetch, error)
	FetchPastGamesTiebreak(logger zerolog.Logger, ctx context.Context, leagueId int64) ([]entities.GameTiebreak, error)
	FetchMatches(logger zerolog.Logger, ctx context.Context, gameID int64, cfg *config.DBConfig) ([]entities.ShortMatch, error)
	TeamsHaveNoGames(logger zerolog.Logger, ctx context.Context, teams []entities.Team, seasonID int64) (bool, error)

	GetRatingByPlayerIDAndByLeagueID(logger zerolog.Logger, ctx context.Context, playerID, leagueID int64) (int64, error)
	GetPlayerRatingByTournamentId(logger zerolog.Logger, ctx context.Context, playerID, tournamentID int64) (int64, error)
	GetTournamentPlayers(ctx context.Context, logger zerolog.Logger, tournamentId int64, withDeleted bool, tx tx.ITx) ([]entities.TournamentPlayer, error)

	GetMatchListByGameID(ctx context.Context, logger zerolog.Logger, gameID int64, cfg *config.DBConfig) ([]entities.MatchV2, error)

	GetPlayerIDsByLeagueID(logger zerolog.Logger, ctx context.Context, leagueID int64) ([]int64, error)
	GetMatchListByLeagueID(logger zerolog.Logger, ctx context.Context, leagueID int64) ([]entities.Match, error)

	GetTeamExtraPointsCount(logger zerolog.Logger, ctx context.Context, teamID, leagueID int64) (int64, error)
	GetExtraPointsListByTeamAndLeagueId(logger zerolog.Logger, ctx context.Context, teamId, leagueId int64) ([]entities.ExtraPoints, error)
	GetExtraPointsById(logger zerolog.Logger, ctx context.Context, extraPointsId int64) (entities.ExtraPoints, error)

	GetSeasonList(logger zerolog.Logger, ctx context.Context, req *entities.GetSeasonListRequest) ([]entities.Season, error)

	GetTeamById(logger zerolog.Logger, ctx context.Context, teamID int64, tx tx.ITx) (entities.TeamV2, error)
	GetLeagueListByTeamId(logger zerolog.Logger, ctx context.Context, teamID int64) ([]entities.LeagueShort, error)
	GetCaptainByTeamId(logger zerolog.Logger, ctx context.Context, teamID int64) (entities.User, error)
	GetTeamGamesInLeague(logger zerolog.Logger, ctx context.Context, teamId, leagueId int64, tiebreak *bool) ([]entities.Game, error)

	GetTournamentTypeList(logger zerolog.Logger, ctx context.Context, withDeleted bool, cfg *config.DBConfig) ([]entities.TournamentType, error)
	GetTournamentById(logger zerolog.Logger, ctx context.Context, tournamentId int64, cfg *config.DBConfig) (entities.Tournament, error)
	GetTournamentStage(logger zerolog.Logger, ctx context.Context, stageId int64, cfg *config.DBConfig) (entities.TournamentStage, error)
	GetTournamentStageList(logger zerolog.Logger, ctx context.Context, tournamentId int64) ([]entities.TournamentStage, error)
	GetTournamentList(logger zerolog.Logger, ctx context.Context, req entities.GetTournamentListRequest) ([]entities.TournamentShort, int64, error)
	GetTournamentListByPlayerID(logger zerolog.Logger, ctx context.Context, playerId int64) ([]entities.PlayersTournament, error)

	GetSuffix(logger zerolog.Logger, ctx context.Context, tx tx.ITx) (string, error)
}

type dbp struct {
	db  *pgxpool.Pool
	cfg *config.DBConfig
}

// RWDBOperation is a structure that implements the RWDBOperationer interface.
type RWDBOperation dbp

// RDBOperation is a structure that implements the RDBOperationer interface.
type RDBOperation dbp

func NewOperationer(rwConn *pgxpool.Pool, rConn *pgxpool.Pool, cfg *config.Configuration) (RWDBOperationer, RDBOperationer) {
	return &RWDBOperation{rwConn, &cfg.RWDB}, &RDBOperation{rConn, &cfg.RDB}
}

func poolOrTx(pg tx.IExecutor, tx tx.ITx) tx.IExecutor {
	if tx != nil {
		if txExec := tx.Executor(); txExec != nil {
			return txExec
		}
	}

	return pg
}

type Tx struct {
	tx     pgx.Tx
	pg     *pgxpool.Pool
	logger zerolog.Logger
}

func (db *RWDBOperation) BeginTx(ctx context.Context, logger zerolog.Logger) (tx.ITx, error) {
	tX, err := db.db.Begin(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("BeginRW")
		return nil, DecodeDatabaseError(err)
	}

	return &Tx{
		tx:     tX,
		pg:     db.db,
		logger: logger,
	}, nil
}

//func (db *RDBOperation) BeginR(ctx context.Context, logger zerolog.Logger) (tx.ITx, error) {
//	tx, err := db.db.Begin(ctx)
//	if err != nil {
//		logger.Error().Err(err).Msg("BeginR")
//		return nil, DecodeDatabaseError(err)
//	}
//
//	return &Tx{
//		tx: tx,
//		pg: db.db,
//	}, nil
//}

func (t *Tx) Commit(ctx context.Context) error {
	return t.tx.Commit(ctx)
}

func (t *Tx) Rollback(ctx context.Context) {
	logger := t.logger.With().Str("transaction", "Rollback").Logger()
	err := t.tx.Rollback(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("Failed tx.Rollback")
	}
}

func (t *Tx) Executor() tx.IExecutor {
	return t.tx
}
