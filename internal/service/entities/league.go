package entities

type GetLeagueListRequest struct {
	CityID int64 `db:"city_id" json:"cityId"`
}

type GetLeagueListResponse struct {
	Leagues []League `json:"leagues"`
}

type League struct {
	ID       int64  `db:"id" json:"id"`
	Name     string `db:"name" json:"name"`
	CityID   int64  `db:"city_id" json:"cityId"`
	SeasonID *int64 `db:"seasonId" json:"seasonId"`
}

type CreateLeagueRequest struct {
	CityID   int64   `db:"city_id" json:"cityId"`
	Name     string  `db:"name" json:"name"`
	Teams    []int64 `json:"teams"`
	SeasonID *int64  `db:"seasonId" json:"seasonId"`
	Creator  User    `json:"creator"`
}

type CreateLeagueResponse struct {
	ID int64 `db:"id" json:"id"`
}

type UpdateLeagueRequest struct {
	ID       int64   `db:"id" json:"id"`
	Name     string  `db:"name" json:"name"`
	Teams    []int64 `json:"teams"`
	SeasonID *int64  `db:"seasonId" json:"seasonId"`
}

type UpdateLeagueResponse struct {
	ID int64 `db:"id" json:"id"`
}

type DeleteLeagueRequest struct {
	ID int64
}

type RecalcLeagueRequest struct {
	ID int64
}

type ExtraPoints struct {
	Id       int64  `json:"id"`
	TeamId   int64  `json:"teamId"`
	LeagueId int64  `json:"leagueId"`
	Reason   string `json:"reason"`
	Points   int64  `json:"points"`
}

type CreateExtraPointsRequest struct {
	TeamId   int64  `json:"teamId"`
	LeagueId int64  `json:"leagueId"`
	Reason   string `json:"reason"`
	Points   int64  `json:"points"`
}

type UpdateExtraPointsRequest struct {
	Id       int64   `json:"id"`
	TeamId   *int64  `json:"teamId"`
	LeagueId *int64  `json:"leagueId"`
	Reason   *string `json:"reason"`
	Points   *int64  `json:"points"`
}

type IdRequest struct {
	Id int64 `json:"id"`
}

type TeamLeagueIdRequest struct {
	TeamId   int64 `json:"teamId"`
	LeagueId int64 `json:"leagueId"`
}

type LeagueItem struct {
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

type TeamItem struct {
	ID        int    `json:"id"`
	Name      string `json:"name,omitempty"`
	ShortName string `json:"shortName,omitempty"`
	Avatar    []byte `json:"avatar,omitempty"`
	CityID    int    `json:"cityId"`
	Leagues   []int  `json:"leagues"`
}
