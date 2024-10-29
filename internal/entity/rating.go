package entity

import "time"

type Rating struct {
	PlayerID	int64		`db:"player_id" json:"player_id"`
	LeagueID	int64		`db:"league_id" json:"league_id"`
	Value		int64		`db:"value" json:"value"`
	UpdatedAt	time.Time	`db:"updated_at" json:"updated_at"`
}
