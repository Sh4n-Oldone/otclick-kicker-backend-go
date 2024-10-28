package team

import (
	"context"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/entity"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
)

// GetTeam {id}
// GetTeams
// GetTeamsByCity {city_id}
// GetTeamsByLeague {league_id}
// GetTeamVsTeamTable {city_id}

// Create
// Update
// Delete {id}

// AddPlayerIntoTeam
// RemovePlayerFromTeam
// //////////////////////////////////////////////////////////////////////////////////////////////////////////////
func (s *Service) GetTeam(ctx context.Context, teamID int64) (entity.GetTeamResponse, error) {
	logger := s.logger.With().Interface("service", "GetTeam").Logger()

	response, err := s.rdbOperations.GetTeam(logger, ctx, teamID)
	if err != nil {
		return response, err
	}

	if err := s.addPlayersProperties(ctx, response.Players); err != nil {
		logger.Error().Err(err).Msg("Failed to add player properties")
		return response, err
	}

	return response, nil
}

func (s *Service) addPlayersProperties(ctx context.Context, players []entity.PlayerGetTeam) error {
	logger := s.logger.With().Interface("service", "addPlayersProperties").Logger()
	timeout, cancel := context.WithTimeout(ctx, s.config.RDB.MaxIdleConnectionTimeout)
	defer cancel()

	for i := range players {
		p := &players[i] // Создаём указатель на текущий элемент массива

		matches, err := s.rdbOperations.GetMatchesByPlayerID(logger, timeout, p.ID)
		if err != nil {
			return err
		}

		propertyCounting(p, matches)
	}

	return nil
}

func propertyCounting(player *entity.PlayerGetTeam, matches []entities.Match) {
	playersGames := make(map[int]struct{})

	var (
		goalsScoredNumber   int
		goalsConcededNumber int
	)

	for _, match := range matches {
		playersGames[match.GameID] = struct{}{}

		if match.Player1Team1ID == player.ID || match.Player2Team1ID == player.ID {
			goalsScoredNumber += match.ScoreTeam1
			goalsConcededNumber += match.ScoreTeam2
		}
		if match.Player1Team2ID == player.ID || match.Player2Team2ID == player.ID {
			goalsScoredNumber += match.ScoreTeam2
			goalsConcededNumber += match.ScoreTeam1
		}
	}

	player.MatchesPlayed = len(matches)
	player.GamesPlayedNumber = len(playersGames)
	player.GoalsScoredNumber = goalsScoredNumber
	player.GoalsConcededNumber = goalsConcededNumber
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (s *Service) GetTeams(ctx context.Context, onlyFree bool) ([]entity.TeamShort, error) {
	logger := s.logger.With().Interface("service", "GetTeams").Logger()

	teams, err := s.rdbOperations.GetTeams(logger, ctx, onlyFree)
	if err != nil {
		return nil, err
	}

	return teams, nil
}

func (s *Service) GetTeamsByCity(ctx context.Context, onlyFree bool, cityID int64) ([]entity.TeamShort, error) {
	logger := s.logger.With().Interface("service", "GetTeamsByCity").Logger()

	teams, err := s.rdbOperations.GetTeamsByCity(logger, ctx, onlyFree, cityID)
	if err != nil {
		return nil, err
	}

	return teams, nil
}

func (s *Service) GetTeamsByLeague(ctx context.Context, leagueID int64) ([]entity.TeamByLeague, error) {
	logger := s.logger.With().Interface("service", "GetTeamsByLeague").Logger()

	teams, err := s.rdbOperations.GetTeamsByLeague(logger, ctx, leagueID)
	if err != nil {
		return nil, err
	}

	return teams, nil
}

func (s *Service) GetTeamVsTeamTable(ctx context.Context, cityID, year int64) (entity.GetTeamVsTeamTableResponse, error) {
	logger := s.logger.With().Interface("service", "GetTeamVsTeamTable").Logger()

	resp, err := s.rdbOperations.GetTeamVsTeamTable(logger, ctx, cityID, year)
	if err != nil {
		return entity.GetTeamVsTeamTableResponse{}, err
	}

	return resp, nil
}

// ///////////////////////////////////////////////////////////////////////////////////
func (s *Service) Create(ctx context.Context, team entity.CreateTeamRequest) (int64, error) {
	logger := s.logger.With().Interface("service", "Create").Logger()

	id, err := s.rwdbOperations.CreateTeam(logger, ctx, team)
	if err != nil {
		return id, err
	}

	return id, nil
}

func (s *Service) Update(ctx context.Context, UpdateTeamRequest entity.UpdateTeamRequest) (bool, error) {
	logger := s.logger.With().Interface("service", "Update").Logger()

	res, err := s.rwdbOperations.UpdateTeam(logger, ctx, UpdateTeamRequest)
	if err != nil {
		return res, err
	}

	return res, nil
}

func (s *Service) Delete(ctx context.Context, id int64) (bool, error) {
	logger := s.logger.With().Interface("service", "Delete").Logger()

	res, err := s.rwdbOperations.DeleteTeam(logger, ctx, id)
	if err != nil {
		return res, err
	}

	return res, nil
}

// ///////////////////////////////////////////////////////////////////////////////////
func (s *Service) AddPlayerIntoTeam(ctx context.Context, playerID, teamID int64) (bool, error) {
	logger := s.logger.With().Interface("service", "AddPlayerIntoTeam").Logger()

	res, err := s.rwdbOperations.AddPlayerIntoTeam(logger, ctx, playerID, teamID)
	if err != nil {
		return res, err
	}

	return res, nil
}

func (s *Service) RemovePlayerFromTeam(ctx context.Context, playerID, teamID int64) (bool, error) {
	logger := s.logger.With().Interface("service", "RemovePlayerFromTeam").Logger()

	res, err := s.rwdbOperations.RemovePlayerFromTeam(logger, ctx, playerID, teamID)
	if err != nil {
		return res, err
	}

	return res, nil
}
