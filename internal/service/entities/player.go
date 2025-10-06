package entities

import "time"

type FullPlayer struct {
	ID           int          `json:"id"`
	Name         *string      `json:"name,omitempty"`
	SecondName   *string      `json:"secondName,omitempty"`
	LastName     string       `json:"lastName"`
	Avatar       []byte       `json:"avatar,omitempty"`
	ActivePlayer *bool        `json:"activePlayer,omitempty"`
	Deleted      bool         `json:"deleted"`
	CityID       *int         `json:"cityId,omitempty"`
	CityName     *string      `json:"cityName,omitempty"`
	Leagues      []LeagueItem `json:"leagues,omitempty"`
}

type PlayersLeague struct {
	ID     int
	Name   string
	CityID int
	Rating *int
}

type PlayersTournament struct {
	ID     int64
	Name   string
	CityID int64
	Rating *int64
}

type CreatePlayerRequest struct {
	Name         *string `json:"name,omitempty" validate:"omitempty,min=1,max=64"`
	SecondName   *string `json:"secondName,omitempty" validate:"omitempty,min=1,max=64"`
	LastName     string  `json:"lastName" validate:"required,min=1,max=64"`
	Avatar       []byte  `json:"avatar,omitempty"`
	ActivePlayer *bool   `json:"activePlayer,omitempty"`
	CityID       int64   `json:"cityId"`
	CityIdParam  string  `json:"cityIdParam"`
	Creator      User    `json:"creator" validate:"required"`
}

type CreatePlayerResponse struct {
	ID int `json:"id"`
}

type UpdatePlayerRequest struct {
	ID           int64   `json:"id" validate:"required,gt=0"`
	Name         *string `json:"name,omitempty" validate:"omitempty,min=1,max=64"`
	SecondName   *string `json:"secondName,omitempty" validate:"omitempty,min=1,max=64"`
	LastName     *string `json:"lastName,omitempty" validate:"omitempty,min=1,max=64"`
	ActivePlayer *bool   `json:"activePlayer,omitempty"`
	Avatar       []byte  `json:"avatar,omitempty"`
	CityID       *int32  `json:"cityId,omitempty"`
}

type DeletePlayerRequest struct {
	ID int `json:"id"`
}

type RecoverPlayerRequest struct {
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
	Players     []Player       `json:"players,omitempty"`
	FullPlayers []FullPlayerV2 `json:"fullPlayers,omitempty"`
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
	CityName      *string    `json:"cityName,omitempty"`
	TeamID        *int       `json:"teamId,omitempty"`
	TeamName      *string    `json:"teamName,omitempty"`
	TeamShortName *string    `json:"teamShortName,omitempty"`
	Rating        *int       `json:"rating,omitempty"`
}

type FullPlayerV2 struct {
	ID           int     `json:"id"`
	Name         *string `json:"name,omitempty"`
	SecondName   *string `json:"secondName,omitempty"`
	LastName     string  `json:"lastName,omitempty"`
	Avatar       []byte  `json:"avatar,omitempty"`
	ActivePlayer *bool   `json:"activePlayer"`
	Deleted      bool    `json:"deleted"`
	CityID       *int    `json:"cityId"`
	CityName     *string `json:"cityName"`

	Leagues []LeagueItem `json:"leagues,omitempty"`

	// total stats
	MatchesPlayed             int     `json:"matchesPlayed"`
	GoalsScoredNumber         int     `json:"goalsScoredNumber"`
	GoalsConcededNumber       int     `json:"goalsConcededNumber"`
	GamesPlayedNumber         int     `json:"gamesPlayedNumber"`
	PercentageOfParticipation float32 `json:"percentageOfParticipation"`

	// deprecated
	// TeamName                  *string      `json:"teamName"`
	// TeamShortName             *string      `json:"teamShortName"`
	// Rating                    *int         `json:"rating,omitempty"`
}

type LeagueItemV2 struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Rating int    `json:"rating"`

	// league stats
	MatchesPlayed             int     `json:"matchesPlayed"`
	GoalsScoredNumber         int     `json:"goalsScoredNumber"`
	GoalsConcededNumber       int     `json:"goalsConcededNumber"`
	GamesPlayedNumber         int     `json:"gamesPlayedNumber"`
	PercentageOfParticipation float32 `json:"percentageOfParticipation"`

	Teams []TeamItem `json:"teams,omitempty"`
}

type OkResponse struct {
	Message string `json:"message"`
}

type TeamItemV2 struct {
	ID        int    `json:"id"`
	Name      string `json:"name,omitempty"`
	ShortName string `json:"shortName,omitempty"`
	Avatar    []byte `json:"avatar,omitempty"`
	CityID    int    `json:"city_id"`
	Leagues   []int  `json:"leagues"`
}
