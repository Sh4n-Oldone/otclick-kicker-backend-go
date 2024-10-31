package entity

import "time"

type Match struct {
	ID             int64     `db:"id" json:"id"`
	Date           time.Time `db:"date" json:"date" validate:"required,valid-date"`
	GameID         *int64    `db:"game_id" json:"gameId"`
	Team1ID        *int64    `db:"team1_id" json:"team1Id"`
	Team2ID        *int64    `db:"team2_id" json:"team2Id"`
	Player1Team1ID *int64    `db:"player1_team1_id" json:"player1Team1Id"`
	Player2Team1ID *int64    `db:"player2_team1_id" json:"player2Team1Id"`
	Player1Team2ID *int64    `db:"player1_team2_id" json:"player1Team2Id"`
	Player2Team2ID *int64    `db:"player2_team2_id" json:"player2Team2Id"`
	ScoreTeam1     *int64    `db:"score_team1" json:"scoreTeam1"`
	ScoreTeam2     *int64    `db:"score_team2" json:"scoreTeam2"`
	Player1Team1RateBefore *int64    `db:"player1_team1_rate_before" json:"player1Team1IdRateBefore"`
	Player2Team1RateBefore *int64    `db:"player2_team1_rate_before" json:"player2Team1IdRateBefore"`
	Player1Team2RateBefore *int64    `db:"player1_team2_rate_before" json:"player1Team2IdRateBefore"`
	Player2Team2RateBefore *int64    `db:"player2_team2_rate_before" json:"player2Team2IdRateBefore"`
	Player1Team1RateAfter *int64    `db:"player1_team1_rate_after" json:"player1Team1IdRateAfter"`
	Player2Team1RateAfter *int64    `db:"player2_team1_rate_after" json:"player2Team1IdRateAfter"`
	Player1Team2RateAfter *int64    `db:"player1_team2_rate_after" json:"player1Team2IdRateAfter"`
	Player2Team2RateAfter *int64    `db:"player2_team2_rate_after" json:"player2Team2IdRateAfter"`
}

type CreateMatchRequest struct {
	Date           time.Time `db:"date" json:"date" validate:"required,valid-date"`
	GameID         *int64    `db:"game_id" json:"gameId"`
	Team1ID        *int64    `db:"team1_id" json:"team1Id"`
	Team2ID        *int64    `db:"team2_id" json:"team2Id"`
	Player1Team1ID *int64    `db:"player1_team1_id" json:"player1Team1Id"`
	Player2Team1ID *int64    `db:"player2_team1_id" json:"player2Team1Id"`
	Player1Team2ID *int64    `db:"player1_team2_id" json:"player1Team2Id"`
	Player2Team2ID *int64    `db:"player2_team2_id" json:"player2Team2Id"`
	ScoreTeam1     *int64    `db:"score_team1" json:"scoreTeam1"`
	ScoreTeam2     *int64    `db:"score_team2" json:"scoreTeam2"`
}

type UpdateMatchRequest struct {
	ID             int64     `db:"id" json:"id"`
	Date           time.Time `db:"date" json:"date" validate:"required,valid-date"`
	GameID         *int64    `db:"game_id" json:"gameId"`
	Team1ID        *int64    `db:"team1_id" json:"team1Id"`
	Team2ID        *int64    `db:"team2_id" json:"team2Id"`
	Player1Team1ID *int64    `db:"player1_team1_id" json:"player1Team1Id"`
	Player2Team1ID *int64    `db:"player2_team1_id" json:"player2Team1Id"`
	Player1Team2ID *int64    `db:"player1_team2_id" json:"player1Team2Id"`
	Player2Team2ID *int64    `db:"player2_team2_id" json:"player2Team2Id"`
	ScoreTeam1     *int64    `db:"score_team1" json:"scoreTeam1"`
	ScoreTeam2     *int64    `db:"score_team2" json:"scoreTeam2"`
}

type DeleteMatchRequest struct {
	ID int64
}
