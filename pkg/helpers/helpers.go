package helpers

import (
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"google.golang.org/grpc/codes"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/error_templates"
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
		s = strings.TrimSpace(s)

		if s == "" {
			return ""
		}

		for _, r := range s {
			if unicode.IsPrint(r) {
				if unicode.IsLetter(r) {
					return string(unicode.ToUpper(r))
				} else {
					return string(r)
				}
			} else {
				return ""
			}
		}
		return ""
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

func SortKeysValues[K int64 | int, T any](m map[K]T) ([]K, []T, error) {
	keys := make([]K, 0, len(m))
	vals := make([]T, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	for i, k := range keys {
		if K(i+1) != k {
			err := errors.New("нарушена нумерация значений")
			return nil, nil, err
		}
		vals = append(vals, m[k])
	}

	return keys, vals, nil
}

func CheckQtyWinnersInPreviousPlayoffStage(startTeamsQty int, lastStageWinnersQty int, nextStageNumber int, stageQty int) error {
	teamsCount := startTeamsQty

	for i := 1; i <= stageQty; i++ {

		if nextStageNumber == i {
			if lastStageWinnersQty != teamsCount {
				err := errors.New("ошибочное кол-во победителей")
				return error_templates.New(err.Error(), err, codes.FailedPrecondition, http.StatusConflict)
			}
			return nil
		}
		teamsCount = teamsCount / 2
	}

	err := errors.New("непредвиденная ошибка")
	return error_templates.New(err.Error(), err, codes.Internal, http.StatusInternalServerError)
}

func CheckQtyWinnersInCurrentPlayoffStage(startTeamsQty int, currentStageWinnersQty int, currentStageNumber int, stageQty int) error {
	teamsCount := startTeamsQty

	for i := 1; i <= stageQty; i++ {

		if currentStageNumber == i {
			if currentStageWinnersQty != (teamsCount / 2) {
				err := errors.New("не удалось выявить победителя")
				return error_templates.New(err.Error(), err, codes.FailedPrecondition, http.StatusConflict)
			}
			return nil
		}
		teamsCount = teamsCount / 2
	}

	err := errors.New("непредвиденная ошибка")
	return error_templates.New(err.Error(), err, codes.Internal, http.StatusInternalServerError)
}

func ExtractWinners(games []entities.GameStat) []int64 {
	pairGames := make(map[[2]int64][]entities.GameStat)

	for _, game := range games {
		var key [2]int64
		if game.Game.Team1ID < game.Game.Team2ID {
			key = [2]int64{game.Game.Team1ID, game.Game.Team2ID}
		} else {
			key = [2]int64{game.Game.Team2ID, game.Game.Team1ID}
		}

		pairGames[key] = append(pairGames[key], game)
	}

	winners := make([]int64, 0, len(pairGames))

	for k, gameList := range pairGames {

		winners1 := make([]int64, 0, len(gameList))
		winners2 := make([]int64, 0, len(gameList))

		for _, game := range gameList {
			if game.WinnerId == k[0] {
				winners1 = append(winners1, k[0])
			} else if game.WinnerId == k[1] {
				winners2 = append(winners2, k[1])
			}
		}

		if len(winners1) > len(winners2) {
			winners = append(winners, winners1[0])
		} else if len(winners2) > len(winners1) {
			winners = append(winners, winners2[0])
		}
	}

	return winners
}

func ParseTournamentStageNumber(number, separator string) (int64, int64, error) {
	orderNumberList := strings.Split(number, separator)
	if len(orderNumberList) != 2 {
		return 0, 0, errors.New("номер этапа не соответствует шаблону")
	}

	orderNumber, err := strconv.ParseInt(orderNumberList[0], 10, 64)
	if err != nil {
		return 0, 0, err
	}

	if orderNumber <= 0 {
		err = errors.New("порядковый номер этапа должен быть больше 0")
		return 0, 0, err
	}

	stageQty, err := strconv.ParseInt(orderNumberList[1], 10, 64)
	if err != nil {
		return 0, 0, err
	}

	if stageQty < orderNumber {
		err = errors.New("порядковый номер этапа не может превышать число этапов")
		return 0, 0, err
	}

	return orderNumber, stageQty, nil
}
