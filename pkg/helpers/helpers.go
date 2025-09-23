package helpers

func GeneratePairs(teamIDs []int64, bestOf int) (team1IDs, team2IDs []int64) {
	size := len(teamIDs)
	if size < 2 || bestOf < 1 {
		return []int64{}, []int64{}
	}

	pairCount := size * (size - 1) / 2 * bestOf
	team1IDs = make([]int64, 0, pairCount)
	team2IDs = make([]int64, 0, pairCount)

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
