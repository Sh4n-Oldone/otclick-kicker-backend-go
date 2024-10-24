package entities

import "time"

type PlayersMatch struct {
	ID             int
	Date           time.Time
	GameID         int
	Team1ID        int
	Team2ID        int
	Player1Team1ID int
	Player2Team1ID int
	Player1Team2ID int
	Player2Team2ID int
	ScoreTeam1     int
	ScoreTeam2     int
}
