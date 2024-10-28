package entity

type PlayerTeam struct {
	PlayerID int64 `db:"player_id" json:"playerId"`
	TeamID   int64 `db:"team_id" json:"teamId"`
}

type PlayerGetTeam struct {
	ID                  int     `json:"id"`
	Name                *string `json:"name,omitempty"`
	SecondName          *string `json:"secondName,omitempty"`
	LastName            string  `json:"lastName,omitempty"`
	TeamName            *string `json:"teamName"`
	TeamShortName       *string `json:"teamShortName"`
	Avatar              []byte  `json:"avatar,omitempty"`
	MatchesPlayed       int     `json:"matchesPlayed"`
	GoalsScoredNumber   int     `json:"goalsScoredNumber"`
	GoalsConcededNumber int     `json:"goalsConcededNumber"`
	GamesPlayedNumber   int     `json:"gamesPlayedNumber"`
	RatingNumber        *int    `json:"ratingNumber"`
	ActivePlayer        bool    `json:"activePlayer"`
	Deleted             bool    `json:"deleted"`
	CityID              *int64  `json:"cityId"`
}

// - 'percentageOfParticipation'
// - 'leagues'
// - 'leaguesCount'

type TeamShort struct { // ByCity
	ID        int64  `db:"id" json:"id"`
	Name      string `db:"name" json:"name"`
	ShortName string `db:"short_name" json:"shortName"`
}

type TeamByLeague struct {
	Name      string  `db:"name" json:"name"`
	ShortName string  `db:"short_name" json:"shortName"`
	Avatar    []byte  `json:"avatar,omitempty"`
	CityId    *int64  `db:"city_id" json:"cityId"`
	Players   []int64 `json:"players"`
}

// /////////////////////////////////////////////////////////////////////////
type GetTeamResponse struct { //TeamFilledWithFullPlayers
	ID        int64           `db:"id" json:"id"`
	Name      string          `db:"name" json:"name"`
	ShortName string          `db:"short_name" json:"shortName"`
	Avatar    []byte          `json:"avatar,omitempty"`
	CityId    *int64          `db:"city_id" json:"cityId"`
	Players   []PlayerGetTeam `db:"players" json:"players"`
}

// /////////////////////////////////////////////////////////////////////////

type CreateTeamRequest struct {
	Name      string `db:"name" json:"name"`
	ShortName string `db:"short_name" json:"shortName"`
	Avatar    []byte `json:"avatar,omitempty"`
	CityId    *int64 `db:"city_id" json:"cityId"`
	LeagueID  *int64 `db:"league_id" json:"leagueId"`
}

type UpdateTeamRequest struct {
	ID        int64   `db:"id" json:"id"`
	Name      *string `db:"name" json:"name"`
	ShortName *string `db:"short_name" json:"shortName"`
	Avatar    []byte  `json:"avatar,omitempty"`
	CityId    *int64  `db:"city_id" json:"cityId"`
	LeagueID  *int64  `db:"league_id" json:"leagueId"`
}

type DeleteTeamRequest struct {
	ID int64
}

type GetTeamRequest struct {
	ID int64
}

type GetTeamsRequest struct {
	CityId   int64 `json:"cityId"`
	OnlyFree bool  `json:"onlyFree"`
}

type GetTeamsByCityRequest struct {
	OnlyFree bool  `json:"onlyFree"`
	CityID   int64 `db:"city_id" json:"city_id" validate:"required,gt=0"`
}
type GetTeamsByLeagueRequest struct {
	OnlyFree bool  `json:"onlyFree"`
	LeagueID int64 `db:"league_id" json:"league_id" validate:"required,gt=0"`
}

// //////////////////////////////////

type GetTeamVsTeamTableRequest struct {
	CityID int64 `json:"cityId"`
	Year   int64 `json:"year"`
	// WithoutEmpty bool  `json:" withoutEmpty"`
}

type GetTeamVsTeamTableResponse struct {
	// CityID int64 `json:"cityId"`
	// Year   int64 `json:"year"`
}

// //////////////////////////////////
// используется в entity/user.go
type Team struct {
	ID        int64  `json:"id" db:"id"`
	Name      string `json:"name" db:"name"`
	ShortName string `json:"short_name" db:"shortName"`
	Avatar    []byte `json:"avatar,omitempty"`
	League    League
	City      City
}

// /////////////////////////////////////////////////////////////////////
