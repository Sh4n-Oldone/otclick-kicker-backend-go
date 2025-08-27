package place

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/config"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/database/postgresql"
)

type IService interface {
	GetList(ctx context.Context, barID, tableID, cityID *int64, withDeleted bool) ([]entities.Place, error)
	Get(ctx context.Context, id int64) (*entities.Place, error)
	Create(ctx context.Context, place entities.CreatePlaceRequest) (*int64, error)
	Update(ctx context.Context, city entities.UpdatePlaceRequest) error
	Delete(ctx context.Context, id int64, executor entities.User) error

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
