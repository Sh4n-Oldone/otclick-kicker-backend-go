package match

import (
	"context"

	"github.com/rs/zerolog"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/database/postgresql"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/config"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/entity"
)

type IService interface {
	Create(ctx context.Context, match entity.Match) (*int64, error)
	Update(ctx context.Context, match entity.Match) error
	Delete(ctx context.Context, id int64) (bool, error)

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
