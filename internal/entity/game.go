package entity

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
	LeagueID   int         `json:"leagueId"`
	LeagueName string      `json:"leagueName"`
	Teams      []TeamShort `json:"teams"`
}

type CreateFutureGameRequest struct {
	CityID     int        `json:"cityId" validate:"required,gt=0"`
	LeagueID   int        `json:"leagueId" validate:"required,gt=0"`
	Date       *time.Time `json:"date"`
	PlaceID    *int       `json:"placeId"`
	Team1ID    int        `json:"team1Id" validate:"required,gt=0"`
	Team2ID    int        `json:"team2Id" validate:"required,gt=0"`
	IsTiebreak bool       `json:"isTiebreak"`
}

type CreateFutureGameResponse struct {
	ID int `json:"id"`
}

type UpdateFutureGameRequest struct {
	ID       int        `json:"id" validate:"required,gt=0"`
	LeagueID int        `json:"leagueId" validate:"required,gt=0"`
	Date     *time.Time `json:"date"`
	PlaceID  *int       `json:"placeId"`
	Team1ID  int        `json:"team1Id" validate:"required,gt=0"`
	Team2ID  int        `json:"team2Id" validate:"required,gt=0"`
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
	ID      int64 `json:"id" validate:"required,gt=0"`
	Team1ID int64 `json:"team1Id" validate:"required,gt=0"`
	Team2ID int64 `json:"team2Id" validate:"required,gt=0"`
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
	TechLooseTeamId *int       `json:"techLooseTeamId,omitempty"`
}
