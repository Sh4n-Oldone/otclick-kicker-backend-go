package league

import (
	"context"

	"github.com/rs/zerolog"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/database/postgresql"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/config"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
)

type IService interface {
	GetList(ctx context.Context, cityID int64) ([]entities.League, error)
	Create(ctx context.Context, req *entities.CreateLeagueRequest) (*int64, error)
	Update(ctx context.Context, league entities.League, teams []int64) error
	Delete(ctx context.Context, id int64) error

	Recalc(ctx context.Context, id int64) error

	CreateExtraPoints(ctx context.Context, req *entities.CreateExtraPointsRequest) (int64, error)
	UpdateExtraPoints(ctx context.Context, req *entities.UpdateExtraPointsRequest) (bool, error)
	DeleteExtraPoints(ctx context.Context, extraPointsId int64) (bool, error)

	GetExtraPointsListByTeamAndLeagueId(ctx context.Context, teamId, leagueId int64) ([]entities.ExtraPoints, error)
	GetExtraPointsById(ctx context.Context, extraPointsId int64) (entities.ExtraPoints, error)

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
