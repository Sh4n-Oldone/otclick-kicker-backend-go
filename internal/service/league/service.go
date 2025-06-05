package league

import (
	"context"

	"github.com/rs/zerolog"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/database/postgresql"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/config"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/entity"
)

type IService interface {
	GetList(ctx context.Context, cityID int64) ([]entity.League, error)
	Create(ctx context.Context, league entity.League, teams []int64) (*int64, error)
	Update(ctx context.Context, league entity.League, teams []int64) error
	Delete(ctx context.Context, id int64) error

	Recalc(ctx context.Context, id int64) error

	CreateExtraPoints(ctx context.Context, req *entity.CreateExtraPointsRequest) (int64, error)
	UpdateExtraPoints(ctx context.Context, req *entity.UpdateExtraPointsRequest) (bool, error)
	DeleteExtraPoints(ctx context.Context, extraPointsId int64) (bool, error)

	GetExtraPointsListByTeamAndLeagueId(ctx context.Context, teamId, leagueId int64) ([]entity.ExtraPoints, error)
	GetExtraPointsById(ctx context.Context, extraPointsId int64) (entity.ExtraPoints, error)

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
