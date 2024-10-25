package team

import (
	"context"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/entity"
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

// func (s *Service) GetTeam(ctx context.Context, teamID int64) (entity.GetTeamResponse, error) {
// 	logger := s.logger.With().Interface("service", "GetTeam").Logger()

// 	response, err := s.rdbOperations.GetTeam(logger, ctx, teamID)
// 	if err != nil {
// 		return response, err
// 	}

// 	return response, nil
// }

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// func (s *Service) GetTeams(ctx context.Context, team entity.Team) (*int64, error) {
// 	logger := s.logger.With().Interface("service", "GetTeams").Logger()

// 	id, err := s.rdbOperations.GetTeams(logger, ctx, team)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &id, nil
// }

// func (s *Service) GetTeamsByCity(ctx context.Context, team entity.Team) (*int64, error) {
// 	logger := s.logger.With().Interface("service", "GetTeamsByCity").Logger()

// 	id, err := s.rdbOperations.GetTeamsByCity(logger, ctx, team)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &id, nil
// }

// func (s *Service) GetTeamsByLeague(ctx context.Context, team entity.Team) (*int64, error) {
// 	logger := s.logger.With().Interface("service", "GetTeamsByLeague").Logger()

// 	id, err := s.rdbOperations.GetTeamsByLeague(logger, ctx, team)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &id, nil
// }

// func (s *Service) GetTeamVsTeamTable(ctx context.Context, team entity.Team) (*int64, error) {
// 	logger := s.logger.With().Interface("service", "GetTeamVsTeamTable").Logger()

// 	id, err := s.rdbOperations.GetTeamVsTeamTable(logger, ctx, team)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &id, nil
// }

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
