package entity

type PlayerTeam struct {
	PlayerID int64 `db:"player_id" json:"player_id"`
	TeamID   int64 `db:"team_id" json:"team_id"`
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
	RatingNumber        int     `json:"raitingNumber"`
	ActivePlayer        bool    `json:"activePlayer"`
	Deleted             bool    `json:"deleted"`
	CityID              int     `json:"cityId"`
}

// - 'percentageOfParticipation'
// - 'leagues'
// - 'leaguesCount'

type TeamShort struct {
	ID        int64  `db:"id" json:"id"`
	Name      string `db:"name" json:"name"`
	ShortName string `db:"short_name" json:"short_name"`
}

// /////////////////////////////////////////////////////////////////////////
type GetTeamResponse struct { //TeamFilledWithFullPlayers
	ID         int64   `db:"id" json:"id"`
	Name       string  `db:"name" json:"name"`
	ShortName  string  `db:"short_name" json:"short_name"`
	Avatar     *string `db:"avatar" json:"avatar"`
	CityId     *int64  `db:"city_id" json:"city_id"`
	LeagueID   *int64  `db:"league_id" json:"league_id"`
	Players    []PlayerGetTeam
	PlayersIDs []int64
}

// /////////////////////////////////////////////////////////////////////////

type CreateTeamRequest struct {
	Name      string  `db:"name" json:"name"`
	ShortName string  `db:"short_name" json:"short_name"`
	Avatar    *string `db:"avatar" json:"avatar"`
	CityId    *int64  `db:"city_id" json:"city_id"`
	LeagueID  *int64  `db:"league_id" json:"league_id"`
}

type UpdateTeamRequest struct {
	ID        int64   `db:"id" json:"id"`
	Name      *string `db:"name" json:"name"`
	ShortName *string `db:"short_name" json:"short_name"`
	Avatar    *string `db:"avatar" json:"avatar"`
	CityId    *int64  `db:"city_id" json:"city_id"`
	LeagueID  *int64  `db:"league_id" json:"league_id"`
}

type DeleteTeamRequest struct {
	ID int64
}
type GetTeamRequest struct {
	ID int64
}

////////////////////////////////////

type Team struct {
	ID        int64  `json:"id" db:"id"`
	Name      string `json:"name" db:"name"`
	ShortName string `json:"short_name" db:"short_name"`
	Avatar    string `json:"avatar" db:"avatar"`
	League    League
	City      City
}
