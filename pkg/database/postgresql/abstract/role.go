package abstract

import (
	"context"

	"github.com/rs/zerolog"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
)

type IRoleR interface {
	GetRoleList(logger zerolog.Logger, ctx context.Context) ([]entities.Role, error)
	GetRole(logger zerolog.Logger, ctx context.Context, id *int64, name *string) (*entities.Role, error)
}

type IRoleRW interface {
}
