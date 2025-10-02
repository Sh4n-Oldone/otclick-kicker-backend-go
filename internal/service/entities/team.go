package entities

type MovingPlayerTeam struct {
	PlayerID int64 `json:"playerId" validate:"required,gt=0"`
	TeamID   int64 `json:"teamId" validate:"required,gt=0"`
	Executor User  `validate:"required"`
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

type GetTeamResponseV2 struct {
	ID           int64            `db:"id" json:"id"`
	Name         string           `db:"name" json:"name"`
	ShortName    string           `db:"short_name" json:"shortName"`
	Avatar       []byte           `json:"avatar,omitempty"`
	CityId       *int64           `db:"city_id" json:"cityId"`
	Captain      User             `json:"captain"`
	LeaguesStats []TeamLeagueStat `json:"leaguesStats"`
	Players      []FullPlayer     `db:"players" json:"players"`
}

type TeamV2 struct {
	Id         int64   `db:"id" json:"id"`
	Name       string  `db:"name" json:"name"`
	ShortName  string  `db:"short_name" json:"shortName"`
	CityId     *int64  `db:"city_id" json:"cityId"`
	Avatar     []byte  `json:"avatar,omitempty"`
	PlayersIds []int64 `json:"playersIds"`
}

type TeamLeagueStat struct {
	League          LeagueShort `json:"league"`
	Points          int         `json:"points"`
	ScoreDifference int         `json:"scoreDifference"`
	GamesCount      int         `json:"gamesCount"`
	BestPlayer      FullPlayer  `json:"bestPlayer,omitempty"`
}

type CreateTeamRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=64"`
	ShortName   string `json:"shortName" validate:"required,min=1,max=32"`
	Avatar      []byte `json:"avatar,omitempty"`
	CityId      int64  `json:"cityId"`
	CityIdParam string `json:"cityIdParam"`
	Creator     User   `json:"creator" validate:"required"`
}

type CreateTeamByMasterRequest struct {
	Name      string           `json:"name" validate:"required,min=1,max=64"`
	ShortName string           `json:"shortName" validate:"required,min=1,max=32"`
	Avatar    []byte           `json:"avatar,omitempty"`
	Master    TournamentMaster `validate:"required"`
}

type UpdateTeamRequest struct {
	ID        int64   `json:"id" validate:"required,gt=0"`
	Name      *string `json:"name" validate:"omitempty,min=1,max=64"`
	ShortName *string `json:"shortName" validate:"omitempty,min=1,max=32"`
	Avatar    []byte  `json:"avatar,omitempty"`
	CityId    *int64  `json:"cityId" validate:"omitempty,gt=0"`
	Updater   User    `validate:"required"`
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

type GetTeamVsTeamTableRequest struct {
	CityID   int64 `json:"cityId"`
	SeasonID int64 `json:"seasonId"`
	// WithoutEmpty bool  `json:" withoutEmpty"`
}

type Team struct {
	ID        int64   `json:"id" db:"id"`
	Name      string  `json:"name" db:"name"`
	ShortName string  `json:"shortName" db:"short_name"`
	Avatar    []byte  `json:"avatar,omitempty"`
	League    *League `json:"league,omitempty"`
	City      *City   `json:"city,omitempty"`
}

type GetTeamVsTeamTableResponse struct {
	Data    []Data `json:"data"`
	Message string `json:"message"`
}

type Data struct {
	LeagueID      int64          `json:"id"`
	LeagueName    string         `json:"name"`
	Table         TableLeague    `json:"table"`
	GamesTiebreak []GameTiebreak `json:"gamesTiebreak"`
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

type TournamentTeam struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	ShortName string `json:"shortName"`
	CityID    int64  `json:"cityId"`
	Avatar    []byte `json:"avatar,omitempty"`
}

type FullTournamentTeam struct {
	TournamentTeam `json:"team"`
	Players        []Player `json:"players"`
}
