package player

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/config"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/database/postgresql"
)

type IService interface {
	Create(ctx context.Context, player *entities.CreatePlayerRequest) (int, error)
	Delete(ctx context.Context, playerID int) error
	Recover(ctx context.Context, playerID int) error
	Update(ctx context.Context, player entities.UpdatePlayerRequest) error
	Find(ctx context.Context, player entities.FindPlayersRequest) (entities.FindPlayersResponse, error)
	FindV2(ctx context.Context, player entities.FindPlayersRequest) (entities.FindPlayersResponse, error)
	Get(ctx context.Context, id int) (entities.FullPlayerV2, error)
	GetByTeamID(ctx context.Context, teamID int) ([]entities.Player, error)
	GetTournamentPlayerList(ctx context.Context, req *entities.GetTournamentPlayerListRequest) ([]entities.TournamentPlayerItem, error)

	GetLogger() *zerolog.Logger
	GetValidator() *validator.Validate
}

type Service struct {
	logger         *zerolog.Logger
	validator      *validator.Validate
	rwdbOperations postgresql.RWDBOperationer
	rdbOperations  postgresql.RDBOperationer
	config         *config.Configuration
}

// GetLogger is a method of business logic layer that gets a logger for logging events in an upper layer.
func (s *Service) GetLogger() *zerolog.Logger {
	return s.logger
}

func (s *Service) GetValidator() *validator.Validate {
	return s.validator
}

func NewService(
	config *config.Configuration,
	logger *zerolog.Logger,
	validator *validator.Validate,
	rwdbOperationer postgresql.RWDBOperationer,
	rdbOperationer postgresql.RDBOperationer,
) IService {
	return &Service{
		logger:         logger,
		validator:      validator,
		rwdbOperations: rwdbOperationer,
		rdbOperations:  rdbOperationer,
		config:         config,
	}
}
