package postgresql

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
)

func (db *RDBOperation) GetUser(logger zerolog.Logger, ctx context.Context, id *int64, email *string) (*entities.User, error) {
	var err error

	if id == nil && email == nil {
		return nil, nil
	}

	role := &entities.Role{}
	team := &entities.Team{}
	user := &entities.User{}
	user.Role = role
	user.Team = team

	var row pgx.Row
	var errorMsg string

	if id != nil {
		errorMsg = "failed to get User by ID"
		row = db.db.QueryRow(ctx, queryGetUserByID, id)
	}

	if id == nil && email != nil {
		errorMsg = "failed to get User by email"
		row = db.db.QueryRow(ctx, queryGetUserByEmail, email)
	}

	err = row.Scan(&user.ID, &user.Email, &user.Password, &user.Role.ID, &user.Role.Name, &user.Team.ID)
	if err != nil {
		logger.Error().Stack().Err(err).Msg(errorMsg)
		return nil, DecodeDatabaseError(err)
	}

	return user, nil
}

func (db *RWDBOperation) CreateUser(logger zerolog.Logger, ctx context.Context, user entities.User) (*int64, error) {
	var id int64

	if user.Role == nil {
		return nil, nil
	}

	if user.Team != nil && user.Team.ID != 0 {
		err := db.db.QueryRow(ctx, queryCreateUserWithTeamID, user.Email, user.Password, user.Role.ID, user.Team.ID).Scan(&id)
		if err != nil {
			logger.Error().Err(err).Msg("failed to create User record")
			return nil, DecodeDatabaseError(err)
		}
	} else {
		err := db.db.QueryRow(ctx, queryCreateUser, user.Email, user.Password, user.Role.ID).Scan(&id)
		if err != nil {
			logger.Error().Err(err).Msg("failed to create User record")
			return nil, DecodeDatabaseError(err)
		}
	}

	return &id, nil
}

func (db *RWDBOperation) UpdateUser(logger zerolog.Logger, ctx context.Context, user entities.User) error {
	tag, err := db.db.Exec(ctx, queryUpdateUser, user.Email, user.Password, user.Role.ID, user.Team.ID, user.ID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to update User record")
		return DecodeDatabaseError(err)
	}
	if tag.RowsAffected() == 0 {
		err = pgx.ErrNoRows
		logger.Error().Err(err).Msg("not found User record to update")
		return DecodeDatabaseError(err)
	}

	return nil
}
