package user

import (
	"context"
	"github.com/go-playground/validator/v10"

	"github.com/rs/zerolog"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/database/postgresql"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/config"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
)

type IService interface {
	Create(ctx context.Context, request entities.CreateUserRequest) (*int64, error)
	Login(ctx context.Context, user *entities.LoginUserRequest) (*int64, *string, *int64, *string, *int64, error)
	ChangePassword(ctx context.Context, userOld, userNew entities.User) error
	CheckAuth(ctx context.Context, userID int64, token string) (*bool, *string, *int64, *int64, error)
	GetUser(ctx context.Context, userID int64) (*entities.User, error)
	CreateTournamentMaster(ctx context.Context, request entities.CreateTournamentMasterRequest) (int64, error)
	UpdateTournamentMaster(ctx context.Context, request entities.UpdateTournamentMasterRequest) error
	GetTournamentMasterListByCityId(ctx context.Context, cityID int64) ([]entities.TournamentMaster, error)
	GetTournamentMasterByUserId(ctx context.Context, userID int64) (entities.TournamentMaster, error)
	GetTournamentMasterList(ctx context.Context) ([]entities.TournamentMaster, error)
	DeleteTournamentMaster(ctx context.Context, userID int64) error

	GetLogger() *zerolog.Logger
	GetValidator() *validator.Validate
}

type Service struct {
	logger         *zerolog.Logger
	config         *config.Configuration
	validator      *validator.Validate
	rdbOperations  postgresql.RDBOperationer
	rwdbOperations postgresql.RWDBOperationer
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
		config:         config,
		logger:         logger,
		validator:      validator,
		rwdbOperations: rwdbOperationer,
		rdbOperations:  rdbOperationer,
	}
}
