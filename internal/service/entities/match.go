package entities

import "time"

type Match struct {
	ID                     int
	Date                   time.Time
	GameID                 int
	Team1ID                int
	Team2ID                int
	Player1Team1ID         int
	Player2Team1ID         *int
	Player1Team2ID         int
	Player2Team2ID         *int
	ScoreTeam1             int
	ScoreTeam2             int
	Player1Team1RateBefore *int `db:"player1_team1_rate_before" json:"player1Team1IdRateBefore"`
	Player2Team1RateBefore *int `db:"player2_team1_rate_before" json:"player2Team1IdRateBefore"`
	Player1Team2RateBefore *int `db:"player1_team2_rate_before" json:"player1Team2IdRateBefore"`
	Player2Team2RateBefore *int `db:"player2_team2_rate_before" json:"player2Team2IdRateBefore"`
	Player1Team1RateAfter  *int `db:"player1_team1_rate_after" json:"player1Team1IdRateAfter"`
	Player2Team1RateAfter  *int `db:"player2_team1_rate_after" json:"player2Team1IdRateAfter"`
	Player1Team2RateAfter  *int `db:"player1_team2_rate_after" json:"player1Team2IdRateAfter"`
	Player2Team2RateAfter  *int `db:"player2_team2_rate_after" json:"player2Team2IdRateAfter"`
}

type GamesMatch struct {
	Date                   time.Time `json:"date" validate:"required,valid-date"`
	Team1ID                int       `json:"team1Id" validate:"required,gt=0"`
	Team2ID                int       `json:"team2Id" validate:"required,gt=0"`
	Player1Team1Id         int       `json:"player1Team1Id" validate:"required,gt=0"`
	Player2Team1Id         *int      `json:"player2Team1Id"`
	Player1Team2Id         int       `json:"player1Team2Id" validate:"required,gt=0"`
	Player2Team2Id         *int      `json:"player2Team2Id"`
	ScoreTeam1             int       `json:"scoreTeam1" validate:"gte=0"`
	ScoreTeam2             int       `json:"scoreTeam2" validate:"gte=0"`
	Player1Team1RateBefore *int      `db:"player1_team1_rate_before" json:"player1Team1IdRateBefore"`
	Player2Team1RateBefore *int      `db:"player2_team1_rate_before" json:"player2Team1IdRateBefore"`
	Player1Team2RateBefore *int      `db:"player1_team2_rate_before" json:"player1Team2IdRateBefore"`
	Player2Team2RateBefore *int      `db:"player2_team2_rate_before" json:"player2Team2IdRateBefore"`
	Player1Team1RateAfter  *int      `db:"player1_team1_rate_after" json:"player1Team1IdRateAfter"`
	Player2Team1RateAfter  *int      `db:"player2_team1_rate_after" json:"player2Team1IdRateAfter"`
	Player1Team2RateAfter  *int      `db:"player1_team2_rate_after" json:"player1Team2IdRateAfter"`
	Player2Team2RateAfter  *int      `db:"player2_team2_rate_after" json:"player2Team2IdRateAfter"`
	Sort                   int       `json:"sort" validate:"gte=0"`
}

type FullMatch struct {
	ID               *int       `json:"id"`
	Date             *time.Time `json:"date"`
	Team1ID          *int       `json:"team1Id"`
	Team2ID          *int       `json:"team2Id"`
	Player1Team1Id   *int       `json:"player1Team1Id"`
	Player1Team1Name *string    `json:"player1Team1Name"`
	Player2Team1Id   *int       `json:"player2Team1Id"`
	Player2Team1Name *string    `json:"player2Team1Name"`
	Player1Team2Id   *int       `json:"player1Team2Id"`
	Player1Team2Name *string    `json:"player1Team2Name"`
	Player2Team2Id   *int       `json:"player2Team2Id"`
	Player2Team2Name *string    `json:"player2Team2Name"`
	ScoreTeam1       *int       `json:"scoreTeam1"`
	ScoreTeam2       *int       `json:"scoreTeam2"`
}

type NewMatch struct {
	ID                     *int      `json:"id"`
	Date                   time.Time `json:"date" validate:"required,valid-date"`
	Team1ID                int       `json:"team1Id" validate:"required,gt=0"`
	Team2ID                int       `json:"team2Id" validate:"required,gt=0"`
	Player1Team1Id         int       `json:"player1Team1Id" validate:"required,gt=0"`
	Player2Team1Id         *int      `json:"player2Team1Id"`
	Player1Team2Id         int       `json:"player1Team2Id" validate:"required,gt=0"`
	Player2Team2Id         *int      `json:"player2Team2Id"`
	ScoreTeam1             int       `json:"scoreTeam1" validate:"gte=0"`
	ScoreTeam2             int       `json:"scoreTeam2" validate:"gte=0"`
	Player1Team1RateBefore *int      `db:"player1_team1_rate_before" json:"player1Team1IdRateBefore"`
	Player2Team1RateBefore *int      `db:"player2_team1_rate_before" json:"player2Team1IdRateBefore"`
	Player1Team2RateBefore *int      `db:"player1_team2_rate_before" json:"player1Team2IdRateBefore"`
	Player2Team2RateBefore *int      `db:"player2_team2_rate_before" json:"player2Team2IdRateBefore"`
	Player1Team1RateAfter  *int      `db:"player1_team1_rate_after" json:"player1Team1IdRateAfter"`
	Player2Team1RateAfter  *int      `db:"player2_team1_rate_after" json:"player2Team1IdRateAfter"`
	Player1Team2RateAfter  *int      `db:"player1_team2_rate_after" json:"player1Team2IdRateAfter"`
	Player2Team2RateAfter  *int      `db:"player2_team2_rate_after" json:"player2Team2IdRateAfter"`
	Sort                   int       `json:"sort" validate:"gte=0"`
}
