package entities

import "time"

type Game struct {
	ID      int
	CityID  int
	Date    time.Time
	Team1ID int
	Team2ID int
}
