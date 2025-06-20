package team

import (
	"context"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/player"

	"github.com/rs/zerolog"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/database/postgresql"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/config"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
)

type IService interface {
	GetTeam(ctx context.Context, teamID int64) (entities.GetTeamResponseV2, error)
	GetTeams(ctx context.Context, cityId int64, onlyFree bool) ([]entities.TeamShort, error)
	GetTeamsByCity(ctx context.Context, onlyFree bool, cityID int64) ([]entities.TeamShort, error)
	GetTeamsByLeague(ctx context.Context, leagueID int64) ([]entities.TeamByLeague, error)
	GetTeamVsTeamTable(ctx context.Context, cityID, seasonID int64) (entities.GetTeamVsTeamTableResponse, error)

	Create(ctx context.Context, team entities.CreateTeamRequest) (int64, error)
	Update(ctx context.Context, team entities.UpdateTeamRequest) (bool, error)
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
	playerSrv      player.IService
}

// GetLogger is a method of business logic layer that gets a logger for logging events in an upper layer.
func (s *Service) GetLogger() *zerolog.Logger {
	return s.logger
}

func NewService(
	config *config.Configuration,
	logger *zerolog.Logger,
	rwdbOperationer postgresql.RWDBOperationer,
	rdbOperationer postgresql.RDBOperationer,
	playerSrv player.IService,
) IService {
	return &Service{
		config:         config,
		logger:         logger,
		rwdbOperations: rwdbOperationer,
		rdbOperations:  rdbOperationer,
		playerSrv:      playerSrv,
	}
}
