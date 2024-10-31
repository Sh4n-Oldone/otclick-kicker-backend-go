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
	CityID  int       `json:"cityId" validate:"required,gt=0"`
	Date    time.Time `json:"date" validate:"valid-date"`
	PlaceID int       `json:"placeId" validate:"gt=0"`
	Team1ID int       `json:"team1Id" validate:"required,gt=0"`
	Team2ID int       `json:"team2Id" validate:"required,gt=0"`
}

type CreateFutureGameResponse struct {
	ID int `json:"id"`
}

type GetTeamGamesResponse struct {
	Games []TeamGame `json:"games"`
}

type TeamGame struct {
	ID    *int       `json:"id,omitempty"`
	Team  TeamShort  `json:"team"`
	Place PlaceShort `json:"place"`
	Date  *time.Time `json:"date,omitempty"`
}
