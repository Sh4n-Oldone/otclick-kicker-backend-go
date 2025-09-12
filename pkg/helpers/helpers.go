package helpers

import (
	"math/rand"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"time"
)

func GenerateUniquePairs(teamIDs []int64) (team1IDs, team2IDs []int64) {
	size := len(teamIDs)
	if size < 2 {
		return []int64{}, []int64{}
	}

	pairCount := size * (size - 1) / 2
	team1IDs = make([]int64, 0, pairCount)
	team2IDs = make([]int64, 0, pairCount)

	// Функция для случайного определения порядка команд в паре
	randomOrder := func(x, y int) (int64, int64) {
		if rand.Intn(2) == 0 {
			return teamIDs[y], teamIDs[x]
		}
		return teamIDs[x], teamIDs[y]
	}

	for i := 0; i < size; i++ {
		for j := i + 1; j < size; j++ {
			team1, team2 := randomOrder(i, j)
			team1IDs = append(team1IDs, team1)
			team2IDs = append(team2IDs, team2)
		}
	}

	return team1IDs, team2IDs
}

func HaveFinishedStage(stages []entities.TournamentStage) bool {
	for _, stage := range stages {
		if stage.IsFinished == true {
			return true
		}
	}
	return false
}

func HaveStartedGames(games []entities.TournamentGame) bool {
	for _, game := range games {
		if game.Date != nil && time.Now().After(*game.Date) {
			return true
		}
	}
	return false
}
