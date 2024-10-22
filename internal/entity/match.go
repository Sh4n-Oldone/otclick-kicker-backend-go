package entity

type Match struct {
	ID             int64   `db:"id" json:"id"`
	Date           *string `db:"date" json:"date"`
	GameID         *int64  `db:"game_id" json:"game_id"`
	Team1ID        *int64  `db:"team1_id" json:"team1_id"`
	Team2ID        *int64  `db:"team2_id" json:"team2_id"`
	Player1Team1ID *int64  `db:"player1_team1_id" json:"player1_team1_id"`
	Player2Team1ID *int64  `db:"player2_team1_id" json:"player2_team1_id"`
	Player1Team2ID *int64  `db:"player1_team2_id" json:"player1_team2_id"`
	Player2Team2ID *int64  `db:"player2_team2_id" json:"player2_team2_id"`
	ScoreTeam1     *int64  `db:"score_team1" json:"score_team1"`
	ScoreTeam2     *int64  `db:"score_team2" json:"score_team2"`
}

type CreateMatchRequest struct {
	Date           *string `db:"date" json:"date"`
	GameID         *int64  `db:"game_id" json:"game_id"`
	Team1ID        *int64  `db:"team1_id" json:"team1_id"`
	Team2ID        *int64  `db:"team2_id" json:"team2_id"`
	Player1Team1ID *int64  `db:"player1_team1_id" json:"player1_team1_id"`
	Player2Team1ID *int64  `db:"player2_team1_id" json:"player2_team1_id"`
	Player1Team2ID *int64  `db:"player1_team2_id" json:"player1_team2_id"`
	Player2Team2ID *int64  `db:"player2_team2_id" json:"player2_team2_id"`
	ScoreTeam1     *int64  `db:"score_team1" json:"score_team1"`
	ScoreTeam2     *int64  `db:"score_team2" json:"score_team2"`
}

type UpdateMatchRequest struct {
	ID             int64   `db:"id" json:"id"`
	Date           *string `db:"date" json:"date"`
	GameID         *int64  `db:"game_id" json:"game_id"`
	Team1ID        *int64  `db:"team1_id" json:"team1_id"`
	Team2ID        *int64  `db:"team2_id" json:"team2_id"`
	Player1Team1ID *int64  `db:"player1_team1_id" json:"player1_team1_id"`
	Player2Team1ID *int64  `db:"player2_team1_id" json:"player2_team1_id"`
	Player1Team2ID *int64  `db:"player1_team2_id" json:"player1_team2_id"`
	Player2Team2ID *int64  `db:"player2_team2_id" json:"player2_team2_id"`
	ScoreTeam1     *int64  `db:"score_team1" json:"score_team1"`
	ScoreTeam2     *int64  `db:"score_team2" json:"score_team2"`
}

type DeleteMatchRequest struct {
	ID int64
}
