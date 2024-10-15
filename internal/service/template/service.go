package template

import (
	"context"
	"github.com/bufbuild/protovalidate-go"
	"github.com/rs/zerolog"
	"node71.otclick.ru/backend/template/pkg/database/postgresql"

	"node71.otclick.ru/backend/template/internal/config"
)

type IService interface {
	Create(ctx context.Context) (*int64, error)

	GetLogger() *zerolog.Logger
}

type Service struct {
	logger         *zerolog.Logger
	validator      *protovalidate.Validator
	config         *config.Configuration
	rdbOperations  postgresql.RDBOperationer
	rwdbOperations postgresql.RWDBOperationer
}

// GetLogger is a method of business logic layer that gets a logger for logging events in a upper layer.
func (s *Service) GetLogger() *zerolog.Logger {
	return s.logger
}

// GetValidator is a method of business logic layer that gets a validator for validating data in a upper layer.
func (s *Service) GetValidator() *protovalidate.Validator {
	return s.validator
}

func NewService(
	config *config.Configuration,
	logger *zerolog.Logger,
	validator *protovalidate.Validator,
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
