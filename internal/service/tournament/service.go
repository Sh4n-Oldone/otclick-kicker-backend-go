package tournament

import (
	"context"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/game"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/config"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/database/postgresql"
)

type IService interface {
	GetTournamentTypeList(ctx context.Context, withDeleted bool) ([]entities.TournamentType, error)
	Create(ctx context.Context, request *entities.CreateTournamentRequest) (int64, error)
	Update(ctx context.Context, request *entities.UpdateTournamentRequest) error
	Delete(ctx context.Context, request *entities.DeleteTournamentRequest) error
	FinishStage(ctx context.Context, request *entities.FinishStageRequest) error
	GetTournamentStageList(ctx context.Context, id int64) ([]entities.TournamentStageItem, error)
	GetTournamentList(ctx context.Context, request *entities.GetTournamentListRequest) ([]entities.TournamentShort, int64, error)
	StartNextStage(ctx context.Context, request *entities.StartNextStageRequest) (int64, error)

	Recalc(ctx context.Context, tournamentId int64) error

	GetLogger() *zerolog.Logger
	GetValidator() *validator.Validate
}

type Service struct {
	logger         *zerolog.Logger
	config         *config.Configuration
	validator      *validator.Validate
	rdbOperations  postgresql.RDBOperationer
	rwdbOperations postgresql.RWDBOperationer
	gameSvc        game.IService
}

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
	gameSvc game.IService,
) IService {
	return &Service{
		logger:         logger,
		config:         config,
		validator:      validator,
		rwdbOperations: rwdbOperationer,
		rdbOperations:  rdbOperationer,
		gameSvc:        gameSvc,
	}
}
