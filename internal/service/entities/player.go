package entities

import "time"

type CreatePlayerRequest struct {
	Name         *string `json:"name,omitempty"`
	SecondName   *string `json:"secondName,omitempty"`
	LastName     string  `json:"lastName"`
	Avatar       []byte  `json:"avatar,omitempty"`
	ActivePlayer *bool   `json:"activePlayer,omitempty"`
	CityID       int     `json:"cityId"`
}

type CreatePlayerResponse struct {
	ID int `json:"id"`
}

type UpdatePlayerRequest struct {
	ID           int     `json:"id"`
	Name         *string `json:"name,omitempty"`
	SecondName   *string `json:"secondName,omitempty"`
	LastName     *string `json:"lastName,omitempty"`
	ActivePlayer *bool   `json:"activePlayer,omitempty"`
	Avatar       *string `json:"avatar,omitempty"`
	CityID       *int32  `json:"cityId,omitempty"`
}

type DeletePlayerRequest struct {
	ID int `json:"id"`
}

type GetPlayerRequest struct {
	ID int `json:"id"`
}

type GetPlayersByTeamIDRequest struct {
	TeamID int `json:"teamId"`
}

type FindPlayersRequest struct {
	LeagueID          *int    `json:"league,omitempty"`
	FindAny           *string `json:"findAny,omitempty"`
	GamesPlayedNumber *int    `json:"gamesPlayedNumber,omitempty"`
	Rating            *int    `json:"rating,omitempty"`
	CityID            *int    `json:"cityId,omitempty"`
	WithDeleted       *bool   `json:"withDeleted,omitempty"`
	OnlyFree          *bool   `json:"onlyFree,omitempty"`
	KeepSimple        *bool   `json:"keepSimple,omitempty"`
}

type FindPlayersResponse struct {
	Players     []Player     `json:"players,omitempty"`
	FullPlayers []FullPlayer `json:"fullPlayers,omitempty"`
}

type Player struct {
	ID            int        `json:"id"`
	Name          *string    `json:"name,omitempty"`
	SecondName    *string    `json:"secondName,omitempty"`
	LastName      string     `json:"lastName"`
	Avatar        []byte     `json:"avatar,omitempty"`
	ActivePlayer  *bool      `json:"activePlayer,omitempty"`
	DeletedAt     *time.Time `json:"deletedAt,omitempty"`
	CityID        *int       `json:"cityId,omitempty"`
	TeamID        *int       `json:"teamId,omitempty"`
	TeamName      *string    `json:"teamName,omitempty"`
	TeamShortName *string    `json:"teamShortName,omitempty"`
	Rating        *int       `json:"rating,omitempty"`
}

type FullPlayer struct {
	ID                        int          `json:"id"`
	Name                      *string      `json:"name,omitempty"`
	SecondName                *string      `json:"secondName,omitempty"`
	LastName                  string       `json:"lastName,omitempty"`
	Avatar                    []byte       `json:"avatar,omitempty"`
	MatchesPlayed             int          `json:"matchesPlayed"`
	GoalsScoredNumber         int          `json:"goalsScoredNumber"`
	GoalsConcededNumber       int          `json:"goalsConcededNumber"`
	GamesPlayedNumber         int          `json:"gamesPlayedNumber"`
	PercentageOfParticipation float32      `json:"percentageOfParticipation"`
	ActivePlayer              *bool        `json:"activePlayer"`
	Deleted                   bool         `json:"deleted"`
	TeamName                  *string      `json:"teamName"`
	TeamShortName             *string      `json:"teamShortName"`
	CityID                    *int         `json:"cityId"`
	Leagues                   []LeagueItem `json:"leagues,omitempty"`
	Rating                    *int         `json:"rating,omitempty"`
}

type LeagueItem struct {
	ID     int `json:"id"`
	Rating int `json:"rating"`
}

type OkResponse struct {
	Message string `json:"message"`
}
