package game

import (
	"context"
	"github.com/rs/zerolog"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/config"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/entity"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/database/postgresql"
)

type IService interface {
	Create(ctx context.Context, request entities.CreateGameRequest) (entities.CreateGameResponse, error)
	Delete(ctx context.Context, gameID int) error
	Get(ctx context.Context, gameID int) (entities.GetGameResponse, error)
	Update(ctx context.Context, request entities.UpdateGameRequest) error
	Find(ctx context.Context, request entities.FindGameRequest) (entities.FindGameResponse, error)
	UpdateFutureGame(ctx context.Context, request entities.UpdateFutureGameRequest) error
	GetGamesYears(ctx context.Context) (entity.GetGamesYearsResponse, error)
	GetComingGames(ctx context.Context) (entities.GetComingGamesResponse, error)

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
