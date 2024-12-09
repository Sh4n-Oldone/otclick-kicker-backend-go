package entities

import (
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/entity"
	"time"
)

type Game struct {
	ID      int
	CityID  int
	Date    *time.Time
	Team1ID int
	Team2ID int
	LeagueID *int
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
}

type CreateGameResponse struct {
	GameID   int   `json:"gameId"`
	MatchIDs []int `json:"matchIds"`
}

type DeleteGameRequest struct {
	ID int `json:"id"`
}

type GetGameResponse struct {
	ID              int               `json:"id"`
	CityID          int               `json:"cityId"`
	Date            *time.Time        `json:"date"`
	Place           entity.PlaceShort `json:"place"`
	LeagueID        int               `json:"leagueId"`
	Team1ID         int               `json:"team1Id"`
	Team1Name       string            `json:"team1Name"`
	Team2ID         int               `json:"team2Id"`
	Team2Name       string            `json:"team2Name"`
	TechLooseTeamID *int              `json:"techLooseTeamId"`
	Matches         []FullMatch       `json:"matches"`
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
