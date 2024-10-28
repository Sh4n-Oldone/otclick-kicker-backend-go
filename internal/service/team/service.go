package team

import (
	"context"

	"github.com/rs/zerolog"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/database/postgresql"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/config"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/entity"
)

type IService interface {
	GetTeam(ctx context.Context, teamID int64) (entity.GetTeamResponse, error)
	GetTeams(ctx context.Context, onlyFree bool) ([]entity.TeamShort, error)
	GetTeamsByCity(ctx context.Context, onlyFree bool, cityID int64) ([]entity.TeamShort, error)
	GetTeamsByLeague(ctx context.Context, leagueID int64) ([]entity.TeamByLeague, error)
	// GetTeamVsTeamTable(ctx context.Context, team entity.Team) (*int64, error)

	Create(ctx context.Context, team entity.CreateTeamRequest) (int64, error)
	Update(ctx context.Context, team entity.UpdateTeamRequest) (bool, error)
	Delete(ctx context.Context, id int64) (bool, error)

	AddPlayerIntoTeam(ctx context.Context, playerID, teamID int64) (bool, error)
	RemovePlayerFromTeam(ctx context.Context, playerID, teamID int64) (bool, error)

	GetLogger() *zerolog.Logger
}

type Service struct {
	logger         *zerolog.Logger
	config         *config.Configuration
	rdbOperations  postgresql.RDBOperationer
	rwdbOperations postgresql.RWDBOperationer
}

// GetLogger is a method of business logic layer that gets a logger for logging events in a upper layer.
func (s *Service) GetLogger() *zerolog.Logger {
	return s.logger
}

func NewService(
	config *config.Configuration,
	logger *zerolog.Logger,
	rwdbOperationer postgresql.RWDBOperationer,
	rdbOperationer postgresql.RDBOperationer,
) IService {
	return &Service{
		config:         config,
		logger:         logger,
		rwdbOperations: rwdbOperationer,
		rdbOperations:  rdbOperationer,
	}
}
