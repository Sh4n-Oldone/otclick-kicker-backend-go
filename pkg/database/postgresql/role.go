package postgresql

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/entity"
)

func (db *RDBOperation) GetRoleList(logger zerolog.Logger, ctx context.Context) ([]entity.Role, error) {
	rows, err := db.db.Query(ctx, queryGetRoleList)
	if err != nil {
		logger.Error().Err(err).Msg("failed to get role list")
		return nil, DecodeDatabaseError(err)
	}
	defer rows.Close()

	var roles []entity.Role

	for rows.Next() {
		var role entity.Role

		err = rows.Scan(&role.ID, &role.Name, &role.Description, &role.UpdatedAt)
		if err != nil {
			logger.Error().Err(err).Msg("failed to scan role record")
			return nil, DecodeDatabaseError(err)
		}

		roles = append(roles, role)
	}

	return roles, nil
}

func (db *RDBOperation) GetRole(logger zerolog.Logger, ctx context.Context, id *int64, name *string) (*entity.Role, error) {
	var err error

	if id == nil && name == nil {
		return nil, nil
	}

	var role entity.Role

	var row pgx.Row
	var errorMsg string

	if id != nil {
		errorMsg = "failed to get Role by ID"
		row = db.db.QueryRow(ctx, queryGetRoleByID, id)
	}

	if id == nil && name != nil {
		errorMsg = "failed to get Role by name"
		row = db.db.QueryRow(ctx, queryGetRoleByName, name)
	}

	err = row.Scan(&role.ID, &role.Name, &role.Description, &role.UpdatedAt)
	if err != nil {
		logger.Error().Stack().Err(err).Msg(errorMsg)
		return nil, DecodeDatabaseError(err)
	}

	return &role, nil
}
