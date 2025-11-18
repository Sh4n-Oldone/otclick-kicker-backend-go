package entities

import "time"

type Year struct {
	Year uint16 `json:"year"`
}

type GetGamesYearsResponse struct {
	Years []Year `json:"years"`
}

type GetFutureGamesResponse struct {
	Games []ShortGame `json:"games"`
}

type GetGameListRequest struct {
	CityId     *int       `json:"cityId" validate:"omitempty,gt=0"`
	LeagueId   *int       `json:"leagueId" validate:"omitempty,gt=0"`
	SeasonId   *int       `json:"seasonId" validate:"omitempty,gt=0"`
	DateFrom   *time.Time `json:"dateFrom"`
	DateTo     *time.Time `json:"dateTo"`
	PlaceId    *int       `json:"placeId" validate:"omitempty,gt=0"`
	Team1Id    *int       `json:"team1Id" validate:"omitempty,gt=0"`
	Team2Id    *int       `json:"team2Id" validate:"omitempty,gt=0"`
	IsTiebreak *bool      `json:"isTiebreak,omitempty"`
	SortField  *int       `json:"sortField,omitempty"`
	SortType   *int       `json:"sortType,omitempty"`
	Limit      *int       `json:"limit,omitempty"`
	Offset     *int       `json:"offset,omitempty"`
}

type GameV2 struct {
	Id              int        `json:"id"`
	Date            *time.Time `json:"date"`
	CityId          *int       `json:"cityId"`
	LeagueId        *int       `json:"leagueId"`
	Season          Season     `json:"season"`
	Place           PlaceShort `json:"place"`
	Team1           TeamShort  `json:"team1"`
	Team2           TeamShort  `json:"team2"`
	ScoreTeam1      *int       `json:"scoreTeam1"`
	ScoreTeam2      *int       `json:"scoreTeam2"`
	TechLooseTeamId *int       `json:"techLooseTeamId"`
	IsTiebreak      *bool      `json:"isTiebreak"`
}

type GetGameListResponse struct {
	Games []GameV2 `json:"games"`
}

type ShortGame struct {
	ID         int         `json:"id"`
	Date       time.Time   `json:"date"`
	CityID     int         `json:"cityId"`
	Place      PlaceShort  `json:"place"`
	LeagueID   *int        `json:"leagueId"`
	LeagueName *string     `json:"leagueName"`
	Teams      []TeamShort `json:"teams"`
}

type CreateFutureGameRequest struct {
	CityID     int        `json:"cityId" validate:"required,gt=0"`
	LeagueID   int        `json:"leagueId" validate:"required,gt=0"`
	Date       *time.Time `json:"date" validate:"omitempty,valid-date"`
	PlaceID    *int       `json:"placeId" validate:"omitempty,gt=0"`
	Team1ID    int        `json:"team1Id" validate:"required,gt=0"`
	Team2ID    int        `json:"team2Id" validate:"required,gt=0"`
	IsTiebreak bool       `json:"isTiebreak"`
	Creator    User
}

type CreateFutureGameResponse struct {
	ID int `json:"id"`
}

type UpdateFutureGameRequest struct {
	ID       int        `json:"id" validate:"required,gt=0"`
	LeagueID int        `json:"leagueId" validate:"required,gt=0"`
	Date     *time.Time `json:"date" validate:"omitempty,valid-date"`
	PlaceID  *int       `json:"placeId" validate:"omitempty,gt=0"`
	Team1ID  int        `json:"team1Id" validate:"required,gt=0"`
	Team2ID  int        `json:"team2Id" validate:"required,gt=0"`
	Executor User
}

type GetTeamGamesResponse struct {
	Games []TeamGame `json:"games"`
}

type TeamGame struct {
	ID         *int       `json:"id,omitempty"`
	IsHomeGame *bool      `json:"isHomeGame,omitempty"`
	Team       TeamShort  `json:"team"`
	Place      PlaceShort `json:"place"`
	Date       *time.Time `json:"date,omitempty"`
}

type DeleteFutureGameRequest struct {
	ID       int64 `json:"id" validate:"required,gt=0"`
	Executor User
}

type Game struct {
	Id              int        `json:"id"`
	CityId          int        `json:"cityId"`
	PlaceId         int        `json:"placeId"`
	Date            *time.Time `json:"date,omitempty"`
	Team1Id         int        `json:"team1Id"`
	Team2Id         int        `json:"team2Id"`
	LeagueId        int        `json:"leagueId"`
	TechLooseTeamId *int       `json:"techLooseTeamId,omitempty"`
	IsHomeGame      bool       `json:"isHomeGame"`
}

type GameTiebreak struct {
	Id              int        `json:"id"`
	CityId          int        `json:"cityId"`
	PlaceId         int        `json:"placeId"`
	Date            *time.Time `json:"date,omitempty"`
	Team1Id         int        `json:"team1Id"`
	Team2Id         int        `json:"team2Id"`
	ScoreTeam1      *int       `json:"scoreTeam1"`
	ScoreTeam2      *int       `json:"scoreTeam2"`
	LeagueId        int        `json:"leagueId"`
	TournamentId    int        `json:"tournamentId"`
	TechLooseTeamId *int       `json:"techLooseTeamId,omitempty"`
}

type GameShort struct {
	ID           int
	CityID       int
	Date         *time.Time
	Team1ID      int
	Team2ID      int
	LeagueID     *int
	TournamentID *int
}

type CreateGameRequest struct {
	CityID          int          `json:"cityId" validate:"required,gt=0"`
	PlaceID         int          `json:"placeId" validate:"required,gt=0"`
	LeagueID        int          `json:"leagueId" validate:"required,gt=0"`
	Date            time.Time    `json:"date" validate:"required,valid-date"`
	Team1ID         int          `json:"team1Id" validate:"required,gt=0"`
	Team2ID         int          `json:"team2Id" validate:"required,gt=0"`
	TechLooseTeamID *int         `json:"techLooseTeamId" validate:"omitempty,gt=0"`
	Matches         []GamesMatch `json:"matches" validate:"omitempty,dive"`
	IsTiebreak      bool         `json:"isTiebreak"`
	Creator         User
}

type CreateGameResponse struct {
	GameID   int64   `json:"gameId"`
	MatchIDs []int64 `json:"matchIds"`
}

type DeleteGameRequest struct {
	ID       int `json:"id"`
	Executor User
}

type GetGameResponse struct {
	ID              int         `json:"id"`
	CityID          int         `json:"cityId"`
	Date            *time.Time  `json:"date"`
	Place           PlaceShort  `json:"place"`
	LeagueID        *int        `json:"leagueId"`
	Team1ID         int         `json:"team1Id"`
	Team1Name       string      `json:"team1Name"`
	Team2ID         int         `json:"team2Id"`
	Team2Name       string      `json:"team2Name"`
	TechLooseTeamID *int        `json:"techLooseTeamId"`
	Matches         []FullMatch `json:"matches"`
}

type UpdateGameRequest struct {
	ID              int        `json:"id" validate:"required,gt=0"`
	Date            time.Time  `json:"date" validate:"required,valid-date"`
	PlaceID         int        `json:"placeId" validate:"required,gt=0"`
	LeagueID        int        `json:"leagueId" validate:"required,gt=0"`
	Team1ID         int        `json:"team1Id" validate:"required,gt=0"`
	Team2ID         int        `json:"team2Id" validate:"required,gt=0"`
	TechLooseTeamID *int       `json:"techLooseTeamId" validate:"omitempty,gt=0"`
	Matches         []NewMatch `json:"matches" validate:"omitempty,dive"`
	Executor        User
}

type FindGameRequest struct {
	Team1ID int `json:"team1Id"`
	Team2ID int `json:"team2Id"`
}

type FindGame struct {
	ID              int       `json:"gameId"`
	Date            time.Time `json:"dateOfGame"`
	Team1ID         int       `json:"team1Id"`
	Team1ShortName  string    `json:"team1ShortName"`
	Team2ID         int       `json:"team2Id"`
	Team2ShortName  string    `json:"team2ShortName"`
	TotalScoreTeam1 int       `json:"totalScoreTeam1"`
	TotalScoreTeam2 int       `json:"totalScoreTeam2"`
}

type FindGameResponse struct {
	Games []FindGame `json:"games"`
}

type GetComingGamesResponse struct {
	ComingGames []ComingGame `json:"games"`
}

type ComingGame struct {
	ID        int       `json:"id"`
	Date      time.Time `json:"date"`
	CityID    int       `json:"cityId"`
	Bar       string    `json:"bar"`
	Table     string    `json:"table"`
	Team1ID   int       `json:"team1Id"`
	Team1Name string    `json:"team1ShortName"`
	Team2ID   int       `json:"team2Id"`
	Team2Name string    `json:"team2ShortName"`
}

type TournamentGame struct {
	ID              int64      `json:"id"`
	CityID          int64      `json:"cityId"`
	PlaceID         *int64     `json:"placeId,omitempty"`
	Date            *time.Time `json:"date,omitempty"`
	Team1ID         int64      `json:"team1Id"`
	Team2ID         int64      `json:"team2Id"`
	TechLooseTeamID *int64     `json:"techLooseTeamId,omitempty"`
	IsTiebreak      bool       `json:"isTiebreak"`
	StageID         int64      `json:"stageId"`
	UpdatedAt       *time.Time `json:"updatedAt"`
}

type CreateFutureTournamentGameRequest struct {
	TournamentID int64      `json:"tournamentId" validate:"required,gt=0"`
	StageID      int64      `json:"stageId" validate:"required,gt=0"`
	PlaceID      *int64     `json:"placeId" validate:"omitempty,gt=0"`
	Date         *time.Time `json:"date" validate:"omitempty,valid-date"`
	Team1ID      int64      `json:"team1Id" validate:"required,gt=0"`
	Team2ID      int64      `json:"team2Id" validate:"required,gt=0"`
	IsTiebreak   bool       `json:"isTiebreak"`
	CityID       int64
	Creator      User
}

type UpdateFutureTournamentGameRequest struct {
	GameID   int64      `validate:"required,gt=0"`
	PlaceID  *int64     `json:"placeId" validate:"omitempty,gt=0"`
	Date     *time.Time `json:"date" validate:"omitempty,valid-date"`
	Team1ID  *int64     `json:"team1Id" validate:"omitempty,gt=0"`
	Team2ID  *int64     `json:"team2Id" validate:"omitempty,gt=0"`
	Executor User
}

type DeleteFutureTournamentGameRequest struct {
	GameID   int64 `validate:"required,gt=0"`
	Executor User
}

type CreatePlayedTournamentGameRequest struct {
	TournamentID    int64        `json:"tournamentId" validate:"required,gt=0"`
	StageID         int64        `json:"stageId" validate:"required,gt=0"`
	PlaceID         int64        `json:"placeId" validate:"required,gt=0"`
	Date            time.Time    `json:"date" validate:"required,valid-date"`
	Team1ID         int64        `json:"team1Id" validate:"required,gt=0"`
	Team2ID         int64        `json:"team2Id" validate:"required,gt=0"`
	IsTiebreak      bool         `json:"isTiebreak"`
	TechLooseTeamID *int64       `json:"techLooseTeamId" validate:"omitempty,gt=0"`
	Matches         []GamesMatch `json:"matches" validate:"omitempty,dive"`
	CityID          int64
	Creator         User
}

type CreatePlayedTournamentGameResponse struct {
	GameId     int64  `json:"gameId"`
	StageId    int64  `json:"stageId"`
	StageState string `json:"stageState"`
}

type UpdatePlayedTournamentGameRequest struct {
	GameID          int64      `validate:"required,gt=0"`
	Date            time.Time  `json:"date" validate:"required,valid-date"`
	PlaceID         int64      `json:"placeId" validate:"required,gt=0"`
	Team1ID         int64      `json:"team1Id" validate:"required,gt=0"`
	Team2ID         int64      `json:"team2Id" validate:"required,gt=0"`
	TechLooseTeamID *int64     `json:"techLooseTeamId" validate:"omitempty,gt=0"`
	Matches         []NewMatch `json:"matches" validate:"omitempty,dive"`
	StageID         int64
	Executor        User
}

type UpdatePlayedTournamentGameResponse struct {
	GameID     int64  `json:"updatedGameId"`
	StageID    int64  `json:"stageId"`
	StageState string `json:"stageState"`
}

type DeletePlayedTournamentGameRequest struct {
	GameID   int64 `validate:"required,gt=0"`
	Executor User
}

type FullTournamentGame struct {
	Game    TournamentGame      `json:"game"`
	Team1   *FullTournamentTeam `json:"team1,omitempty"`
	Team2   *FullTournamentTeam `json:"team2,omitempty"`
	Matches []ShortMatch        `json:"matches,omitempty"`
}

type GetTournamentGameList struct {
	TournamentId *int64 `json:"tournamentId" validate:"omitempty,gt=0"`
	CityId       *int64 `json:"cityId" validate:"omitempty,gt=0"`
}

type GameLeagueTournament struct {
	ID              int64      `json:"id"`
	CityID          int64      `json:"cityId"`
	PlaceID         *int64     `json:"placeId,omitempty"`
	Date            *time.Time `json:"date,omitempty"`
	Team1ID         int64      `json:"team1Id"`
	Team2ID         int64      `json:"team2Id"`
	LeagueID        *int64     `json:"leagueId,omitempty"`
	TechLooseTeamID *int64     `json:"techLooseTeamId,omitempty"`
	IsTiebreak      bool       `json:"isTiebreak"`
	StageID         *int64     `json:"stageId,omitempty"`
}

type GameStat struct {
	Game     TournamentGame
	Matches  []ShortMatch
	WinnerId int64
}
