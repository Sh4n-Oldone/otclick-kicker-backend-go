package helpers

import (
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"google.golang.org/grpc/codes"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/constant"
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

// GeneratePairsRegularPlayoff генерирует пары для regular+playoff турнира
// Генерирует по принципу: лучшая команда играет с худшей, предлучшая играет с предхудшей и т.д.
func GeneratePairsRegularPlayoff[T int64 | int | int32 | int16](teamIDs []T, bestOf int) (team1IDs, team2IDs []T) {
	pairCount := len(teamIDs) / 2

	team1IDs = make([]T, 0, pairCount*bestOf)
	team2IDs = make([]T, 0, pairCount*bestOf)

	homeTeams := make([]T, 0, pairCount)
	awayTeams := make([]T, 0, pairCount)

	generatedTeamIDs := generateFirstPlayoffStageRegularPlayoff(teamIDs)

	for i := 0; i < len(generatedTeamIDs)/2; i++ {
		homeTeams = append(homeTeams, generatedTeamIDs[i])
		awayTeams = append(awayTeams, generatedTeamIDs[len(generatedTeamIDs)-i-1])
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

// generateFirstPlayoffStageRegularPlayoff генерирует пары первого этапа playoff для турнира типа regular+playoff
// Алгоритм разработан по KBACK-167: https://plane.otclick.ru/otclick/browse/KBACK-167/
func generateFirstPlayoffStageRegularPlayoff[T int64 | int | int32 | int16](teamIDs []T) []T {
	if len(teamIDs) == 2 {
		return teamIDs
	}

	middle := len(teamIDs) / 2
	leftHalf := generateFirstPlayoffStageRegularPlayoff[T](teamIDs[:middle])
	rightHalf := generateFirstPlayoffStageRegularPlayoff[T](teamIDs[middle:])

	result := make([]T, 0)
	for i := 0; i < len(leftHalf); i++ {
		result = append(result, leftHalf[i])
		result = append(result, rightHalf[i])
	}

	return result
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

// CheckQtyTeamsInRegularPlayoffStage Проверяет кол-во команд, которые пройдут в следующий этап playoff в regular+playoff турнире
func CheckQtyTeamsInRegularPlayoffStage(neededTeamsForFirstPlayoffStageQty, currentStageTeamsQty, nextStageNumber, stageQty int) error {
	teamsCountPerStage := neededTeamsForFirstPlayoffStageQty

	for i := 1; i <= stageQty; i++ {
		if i == nextStageNumber {
			if teamsCountPerStage > currentStageTeamsQty {
				err := errors.New("недостаточное мин. кол-во команд для перехода в следующий этап")
				return error_templates.New(err.Error(), err, codes.FailedPrecondition, http.StatusConflict)
			}
			return nil
		}

		// если в первый этап playoff (это уже второй этап турнира!) требовалось 8 лучших команд, то в следующий потребуется 8 / 2 = 4 и т.д.
		if i > constant.FirstStage {
			teamsCountPerStage = teamsCountPerStage / 2
		}
	}

	err := errors.New("непредвиденная ошибка")
	return error_templates.New(err.Error(), err, codes.Internal, http.StatusInternalServerError)
}

// CheckQtyTeamsInRegularPlayoffWithLooserBracket Проверяет кол-во команд, которые пройдут в следующий этап playoff в regular+playoff турнире с сеткой лузеров
// Логика турнира с loser bracket:
// - Этап 2 (первый playoff): neededTeamsForFirstPlayoffStageQty команд
// - Нечетные этапы > 2: количество команд делится на 2 (команды приходят из двух источников: верхняя и нижняя сетки)
// - Четные этапы: количество команд остается таким же
func CheckQtyTeamsInRegularPlayoffWithLooserBracket(neededTeamsForFirstPlayoffStageQty, currentStageTeamsQty, nextStageNumber, stageQty int) error {
	if nextStageNumber < 2 || nextStageNumber > stageQty {
		err := errors.New("некорректный номер следующего этапа")
		return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	// Вычисляем требуемое количество команд для следующего этапа сразу
	var requiredTeams int
	if nextStageNumber == 2 {
		requiredTeams = neededTeamsForFirstPlayoffStageQty
	} else {
		// Для этапов > 2: считаем количество нечетных этапов от 3 до nextStageNumber включительно
		// На каждом нечетном этапе количество команд делится на 2
		// Формула: количество нечетных чисел от 3 до N = ceil((N - 2) / 2)
		oddStagesCount := (nextStageNumber - 2 + (nextStageNumber % 2)) / 2
		requiredTeams = neededTeamsForFirstPlayoffStageQty >> oddStagesCount // Используем битовый сдвиг вместо деления на int(math.Pow(2, float64(oddStagesCount)), чтобы избежать ошибки округления при преобразовании float64 в int
	}

	// Проверка для нечетных этапов > 2: команды приходят из двух источников
	if nextStageNumber > 2 && nextStageNumber%2 == 0 { // значит текущий curentStage нечетный, и в винерах команд меньше (ушли в лузеры)
		// На нечетных этапах команды приходят из верхней и нижней сеток
		// Минимальное количество команд из текущего этапа должно быть >= requiredTeams / 2
		// Но так как команды приходят из двух источников, проверяем: requiredTeams <= 2 * currentStageTeamsQty
		if requiredTeams > 2*currentStageTeamsQty {
			err := errors.New("недостаточное мин. кол-во команд для перехода в следующий этап")
			return error_templates.New(err.Error(), err, codes.FailedPrecondition, http.StatusConflict)
		}
		return nil
	}

	// Проверка для этапа 2 и четных этапов > 2
	if requiredTeams > currentStageTeamsQty {
		err := errors.New("недостаточное мин. кол-во команд для перехода в следующий этап")
		return error_templates.New(err.Error(), err, codes.FailedPrecondition, http.StatusConflict)
	}

	return nil
}

func CheckQtyWinnersInPreviousRegularStage(neededTeamsCountToNextStage int32, currentTeamsCount int32) error {
	if neededTeamsCountToNextStage > currentTeamsCount {
		err := errors.New("ошибочное кол-во команд для следующего playoff этапа")
		return error_templates.New(err.Error(), err, codes.FailedPrecondition, http.StatusConflict)
	}

	return nil
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

func ExtractWinnersAndLosers(games []entities.GameStat) ([]int64, []int64) {
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
	losers := make([]int64, 0, len(pairGames))

	for k, gameList := range pairGames {

		winners1 := make([]int64, 0, len(gameList))
		winners2 := make([]int64, 0, len(gameList))
		losers1 := make([]int64, 0, len(gameList))
		losers2 := make([]int64, 0, len(gameList))

		for _, game := range gameList {
			if game.WinnerId == k[0] {
				winners1 = append(winners1, k[0])
				losers1 = append(losers1, k[1])
			} else if game.WinnerId == k[1] {
				winners2 = append(winners2, k[1])
				losers2 = append(losers2, k[0])
			}
		}

		if len(winners1) > len(winners2) {
			winners = append(winners, winners1[0])
			losers = append(losers, losers1[0])
		} else if len(winners2) > len(winners1) {
			winners = append(winners, winners2[0])
			losers = append(losers, losers2[0])
		}
	}

	return winners, losers
}

// ExtractWinnersRegularPlayoff используется для получения отсортированного слайса команд
// по их результатам каждого этапа в турнире типа regular+playoff
func ExtractWinnersRegularPlayoff(games []entities.GameStat, teamsExtraPoints map[int64]int64) ([]int64, error) {
	pairGames := make(map[[2]int64][]entities.GameStat) // Пара команд в играх -> их игры
	teamsRatings := make(map[int64]int64)               // Команда -> Её общий рейтинг

	for _, game := range games {
		var key [2]int64
		if game.Game.Team1ID < game.Game.Team2ID {
			key = [2]int64{game.Game.Team1ID, game.Game.Team2ID}
		} else {
			key = [2]int64{game.Game.Team2ID, game.Game.Team1ID}
		}

		pairGames[key] = append(pairGames[key], game)
	}

	for _, teamsGames := range pairGames {
		for _, g := range teamsGames {
			if g.Game.TechLooseTeamID != nil {
				if *g.Game.TechLooseTeamID == g.Game.Team2ID {
					teamsRatings[g.Game.Team1ID] += int64(constant.TechWinGoals)
					teamsRatings[g.Game.Team2ID] += int64(constant.TechLooseGoals)
				} else {
					teamsRatings[g.Game.Team1ID] += int64(constant.TechWinGoals)
					teamsRatings[g.Game.Team2ID] += int64(constant.TechLooseGoals)
				}
			} else {
				for _, m := range g.Matches {
					teamsRatings[m.Team1ID] += m.ScoreTeam1
					teamsRatings[m.Team2ID] += m.ScoreTeam2
				}
			}
		}
	}

	eTeamRatingList := make([]entities.TeamRating, 0)

	for teamId, rating := range teamsRatings {
		rating += teamsExtraPoints[teamId]
		eTeamRatingList = append(eTeamRatingList, entities.TeamRating{TeamID: teamId, Rating: rating})
	}

	// Получится отсортированный слайс, где команды будут расположены от наибольшего рейтинга (команда, сыгравшая лучше всех)
	// к наименьшему (команда, сыгравшая хуже всех)
	sort.Slice(eTeamRatingList, func(i, j int) bool { return eTeamRatingList[i].Rating > eTeamRatingList[j].Rating })

	// Итоговый слайс команд по убыванию очков, чтобы отобрать в плейофф лучшие N команд
	teams := make([]int64, 0, len(eTeamRatingList))
	for _, res := range eTeamRatingList {
		teams = append(teams, res.TeamID)
	}

	return teams, nil
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
