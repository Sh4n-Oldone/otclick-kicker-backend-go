package helpers

import (
	"strconv"
	"strings"
	"unicode"
)

func GeneratePairs[T int64 | int | int32 | int16](teamIDs []T, bestOf int) (team1IDs, team2IDs []T) {
	size := len(teamIDs)
	if size < 2 || bestOf < 1 {
		return []T{}, []T{}
	}

	pairCount := size * (size - 1) / 2 * bestOf
	team1IDs = make([]T, 0, pairCount)
	team2IDs = make([]T, 0, pairCount)

	for round := 0; round < bestOf; round++ {
		for i := 0; i < size; i++ {
			for j := i + 1; j < size; j++ {
				if round%2 == 1 {
					// меняем порядок в нечетных раундах
					team1IDs = append(team1IDs, teamIDs[j])
					team2IDs = append(team2IDs, teamIDs[i])
				} else {
					team1IDs = append(team1IDs, teamIDs[i])
					team2IDs = append(team2IDs, teamIDs[j])
				}
			}
		}
	}

	return team1IDs, team2IDs
}

func GeneratePlayoffPairs[T int64 | int | int32 | int16](teamIDs []T, bestOf int) (team1IDs, team2IDs []T) {
	if len(teamIDs) < 2 {
		return team1IDs, team2IDs
	}

	pairCount := len(teamIDs) / 2

	team1IDs = make([]T, 0, pairCount*bestOf)
	team2IDs = make([]T, 0, pairCount*bestOf)

	homeTeams := make([]T, pairCount)
	awayTeams := make([]T, pairCount)

	for i := 0; i < pairCount; i++ {
		homeTeams[i] = teamIDs[i*2]
		awayTeams[i] = teamIDs[i*2+1]
	}

	// чередуем домашние и гостевые игры
	for round := 0; round < bestOf; round++ {
		if round%2 == 0 {
			// четный раунд: домашние команды заполняются
			team1IDs = append(team1IDs, homeTeams...)
			team2IDs = append(team2IDs, awayTeams...)
		} else {
			// нечетный раунд: команды меняются местами
			team1IDs = append(team1IDs, awayTeams...)
			team2IDs = append(team2IDs, homeTeams...)
		}
	}

	return team1IDs, team2IDs
}

func IsPowTwo[T int64 | int | int32 | int16 | int8 | uint64 | uint | uint32 | uint16 | uint8](n T) bool {
	if n <= 0 {
		return false
	}
	return n&(n-1) == 0
}

func BuildTeamName(name, secondName *string, lastName string, id int64) (string, string) {
	firstRune := func(s string) string {
		var rr rune
		for _, r := range s {
			if unicode.IsLetter(r) {
				rr = r
				return string(unicode.ToUpper(rr))
			} else {
				rr = r
				return string(rr)
			}
		}
		return string(rr)
	}

	var nameArr []string
	var letters []string

	if name != nil {
		nameArr = append(nameArr, *name)
		letters = append(letters, firstRune(*name))
	}

	if secondName != nil {
		nameArr = append(nameArr, *secondName)
		letters = append(letters, firstRune(*secondName))
	}

	nameArr = append(nameArr, lastName)
	letters = append(letters, firstRune(lastName))

	fullName := strings.Join(nameArr, " ") + "+" + strconv.FormatInt(id, 10)
	short := strings.Join(letters, "") + "+" + strconv.FormatInt(id, 10)

	return fullName, short
}
