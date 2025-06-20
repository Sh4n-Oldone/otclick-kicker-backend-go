package player

import (
	"context"
	"github.com/rs/zerolog"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/config"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/database/postgresql"
)

type IService interface {
	Create(ctx context.Context, player entities.CreatePlayerRequest) (int, error)
	Delete(ctx context.Context, playerID int) error
	Recover(ctx context.Context, playerID int) error
	Update(ctx context.Context, player entities.UpdatePlayerRequest) error
	Find(ctx context.Context, player entities.FindPlayersRequest) (entities.FindPlayersResponse, error)
	Get(ctx context.Context, id int) (entities.FullPlayerV2, error)
	GetByTeamID(ctx context.Context, teamID int) ([]entities.Player, error)

	GetLogger() *zerolog.Logger
}

type Service struct {
	logger         *zerolog.Logger
	rwdbOperations postgresql.RWDBOperationer
	rdbOperations  postgresql.RDBOperationer
	config         *config.Configuration
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
		logger:         logger,
		rwdbOperations: rwdbOperationer,
		rdbOperations:  rdbOperationer,
		config:         config,
	}
}
