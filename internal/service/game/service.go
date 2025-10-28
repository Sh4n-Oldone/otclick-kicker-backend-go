package game

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/config"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/database/postgresql"
)

type IService interface {
	Create(ctx context.Context, request entities.CreateGameRequest) (entities.CreateGameResponse, error)
	Delete(ctx context.Context, req *entities.DeleteGameRequest) error
	Get(ctx context.Context, gameID int) (entities.GetGameResponse, error)
	Update(ctx context.Context, request entities.UpdateGameRequest) error
	Find(ctx context.Context, request entities.FindGameRequest) (entities.FindGameResponse, error)
	UpdateFutureGame(ctx context.Context, request entities.UpdateFutureGameRequest) error
	GetGamesYears(ctx context.Context) (entities.GetGamesYearsResponse, error)
	GetGameList(ctx context.Context, request entities.GetGameListRequest) (entities.GetGameListResponse, error)
	GetComingGames(ctx context.Context) (entities.GetComingGamesResponse, error)
	GetFutureGames(ctx context.Context, cityID int) (entities.GetFutureGamesResponse, error)
	CreateFutureGame(ctx context.Context, request entities.CreateFutureGameRequest) (entities.CreateFutureGameResponse, error)
	GetTeamGames(ctx context.Context, teamID int) (entities.GetTeamGamesResponse, error)
	DeleteFutureGame(ctx context.Context, req entities.DeleteFutureGameRequest) error

	CreateFutureTournamentGame(ctx context.Context, request *entities.CreateFutureTournamentGameRequest) (int64, error)
	UpdateFutureTournamentGame(ctx context.Context, request *entities.UpdateFutureTournamentGameRequest) error
	DeleteFutureTournamentGame(ctx context.Context, request *entities.DeleteFutureTournamentGameRequest) error
	DeletePlayedTournamentGame(ctx context.Context, request *entities.DeletePlayedTournamentGameRequest) error
	CreatePlayedTournamentGame(ctx context.Context, request *entities.CreatePlayedTournamentGameRequest) (int64, error)
	UpdatePlayedTournamentGame(ctx context.Context, request *entities.UpdatePlayedTournamentGameRequest) error
	GetTournamentGameList(ctx context.Context, tournamentId int64) ([]entities.FullTournamentGame, error)
	GetFutureTournamentGameList(ctx context.Context, request *entities.GetTournamentGameList) ([]entities.TournamentGame, error)
	GetPlayedTournamentGameList(ctx context.Context, request *entities.GetTournamentGameList) ([]entities.FullTournamentGame, error)

	GetLogger() *zerolog.Logger
	GetValidator() *validator.Validate
}

type Service struct {
	config         *config.Configuration
	logger         *zerolog.Logger
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
