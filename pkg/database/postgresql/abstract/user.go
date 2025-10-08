package abstract

import (
	"context"

	"github.com/rs/zerolog"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/config"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
)

type IUserR interface {
	GetUser(logger zerolog.Logger, ctx context.Context, id *int64, email *string, cfg *config.DBConfig) (*entities.User, error)
	GetTournamentMasterByUserId(logger zerolog.Logger, ctx context.Context, userId int64, cfg *config.DBConfig) (entities.TournamentMaster, error)
	GetTournamentMasterListByCityId(logger zerolog.Logger, ctx context.Context, cityID int64, cfg *config.DBConfig) ([]entities.TournamentMaster, error)
	GetTournamentMasterList(logger zerolog.Logger, ctx context.Context, cfg *config.DBConfig) ([]entities.TournamentMaster, error)
	GetCaptainByTeamId(logger zerolog.Logger, ctx context.Context, teamID int64) (entities.User, error)
}

type IUserRW interface {
	CreateUser(logger zerolog.Logger, ctx context.Context, user entities.User) (id *int64, err error)
	UpdateUser(logger zerolog.Logger, ctx context.Context, user entities.User) error
	CreateTournamentMaster(logger zerolog.Logger, ctx context.Context, master entities.TournamentMaster, config *config.DBConfig) (id int64, err error)
	UpdateTournamentMaster(logger zerolog.Logger, ctx context.Context, master entities.UpdateTournamentMasterRequest, cfg *config.DBConfig) error
	DeleteTournamentMaster(logger zerolog.Logger, ctx context.Context, userId int64, cfg *config.DBConfig) error
}
