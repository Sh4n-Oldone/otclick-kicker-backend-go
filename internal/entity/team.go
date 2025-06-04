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
	Id        int64   `db:"id" json:"id"`
	Name      string  `db:"name" json:"name"`
	ShortName string  `db:"short_name" json:"shortName"`
	Avatar    []byte  `json:"avatar,omitempty"`
	CityId    *int64  `db:"city_id" json:"cityId"`
	Players   []int64 `json:"players"`
}

// /////////////////////////////////////////////////////////////////////////
type LeagueShort struct {
	ID   int64  `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
}

type GetTeamResponse struct { //TeamFilledWithFullPlayers
	ID        int64           `db:"id" json:"id"`
	Name      string          `db:"name" json:"name"`
	ShortName string          `db:"short_name" json:"shortName"`
	Avatar    []byte          `json:"avatar,omitempty"`
	CityId    *int64          `db:"city_id" json:"cityId"`
	Leagues   []LeagueShort   `db:"leagues" json:"leagues"`
	Players   []PlayerGetTeam `db:"players" json:"players"`
}

// /////////////////////////////////////////////////////////////////////////

type CreateTeamRequest struct {
	Name      string `db:"name" json:"name"`
	ShortName string `db:"short_name" json:"shortName"`
	Avatar    []byte `json:"avatar,omitempty"`
	CityId    *int64 `db:"city_id" json:"cityId"`
}

type UpdateTeamRequest struct {
	ID        int64   `db:"id" json:"id"`
	Name      *string `db:"name" json:"name"`
	ShortName *string `db:"short_name" json:"shortName"`
	Avatar    []byte  `json:"avatar,omitempty"`
	CityId    *int64  `db:"city_id" json:"cityId"`
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
	CityID   int64 `json:"cityId"`
	SeasonID int64 `json:"seasonId"`
	// WithoutEmpty bool  `json:" withoutEmpty"`
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
type GetTeamVsTeamTableResponse struct {
	Data    []Data `json:"data"`
	Message string `json:"message"`
}

type Data struct {
	LeagueID   int64       `json:"id"`
	LeagueName string      `json:"name"`
	Table      TableLeague `json:"table"`
}

type TableLeague struct {
	Columns []Column `json:"columns"`
	Body    []Body   `json:"body"`
}

type Column struct {
	Uid  string `json:"uid"`
	Name string `json:"name"`
}

type TableCell struct {
	Game1ID int64  `json:"game1Id"`
	Game2ID int64  `json:"game2Id"`
	Score1  string `json:"score1"`
	Score2  string `json:"score2"`
}

type Body struct {
	Id                int64  `json:"id"`
	TeamShortName     string `json:"teamShortName"`
	Score             int64  `json:"score"`
	DifferenceInScore int64  `json:"differenceInScore"`
	GamesPlayed       int64  `json:"gamesPlayed"`
	GamesToPlay       int64  `json:"gamesToPlay"`
	TableCell         map[string]TableCell
}

type GameFetch struct {
	ID              int64  `json:"id"`
	TechLooseTeamID *int64 `json:"techLooseTeamId"`
}

// {
// 	"data": [
// 	  {
// 		"id": "",
// 		"name": "",
// 		"table": {
// 		  "columns": [
// 			{
// 			  "uid": "",
// 			  "name": ""
// 			},
// 			...
// 		  ],
// 		  "body": [
// 			{
// 			  "id": "",
// 			  "teamShortName": "",
// 			  "score": "",
// 			  "differenceInScore": "",
// 			  "gamesPlayed": "",
// 			  "gamesToPlay": "",
// 			  "FS": {
// 				"match1Id": 0,
// 				"match2Id": 0,
// 				"score1": "0:0",
// 				"score2": "0:0"
// 			  },
// 			  ...
// 			},
// 			...
// 		  ]
// 		}
// 	  },
// 	  ...
// 	]
// 	"message": "OK"
//   }
