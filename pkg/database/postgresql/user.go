package postgresql

import (
	"context"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/config"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/constant"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/error_templates"
)

func (db *RDBOperation) GetUser(logger zerolog.Logger, ctx context.Context, id *int64, email *string, cfg *config.DBConfig) (*entities.User, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, cfg.MaxIdleConnectionTimeout)
	defer cancel()

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
		row = db.db.QueryRow(timeoutCtx, queryGetUserByID, id)
	}

	if id == nil && email != nil {
		errorMsg = "failed to get User by email"
		row = db.db.QueryRow(timeoutCtx, queryGetUserByEmail, email)
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

func (db *RWDBOperation) CreateTournamentMaster(logger zerolog.Logger, ctx context.Context, master entities.TournamentMaster, cfg *config.DBConfig) (int64, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, cfg.MaxIdleConnectionTimeout)
	defer cancel()

	const queryUser string = `
		INSERT INTO users (email, password, role_id)
		VALUES ($1, $2, $3) RETURNING id;`

	const queryLinkUserToCity string = `
		INSERT INTO users_cities_links (user_id, city_id)
		VALUES ($1, $2);`

	var id int64

	tx, err := db.db.Begin(timeoutCtx)
	if err != nil {
		logger.Error().Err(err).Msg("failed postgresql.CreateTournamentMaster")
		return 0, DecodeDatabaseError(err)
	}

	err = tx.QueryRow(timeoutCtx, queryUser, master.User.Email, master.User.Password, master.User.Role.ID).Scan(&id)
	if err != nil {
		logger.Error().Err(err).Msg("failed postgresql.CreateTournamentMaster")
		_ = tx.Rollback(timeoutCtx)
		return 0, DecodeDatabaseError(err)
	}

	_, err = tx.Exec(timeoutCtx, queryLinkUserToCity, id, master.City.ID)
	if err != nil {
		logger.Error().Err(err).Msg("failed postgresql.CreateTournamentMaster")
		_ = tx.Rollback(timeoutCtx)
		return 0, DecodeDatabaseError(err)
	}

	if err = tx.Commit(timeoutCtx); err != nil {
		logger.Error().Err(err).Msg("failed postgresql.CreateTournamentMaster")
		_ = tx.Rollback(timeoutCtx)
		return 0, DecodeDatabaseError(err)
	}

	return id, nil
}

func (db *RWDBOperation) UpdateTournamentMaster(logger zerolog.Logger, ctx context.Context, request entities.UpdateTournamentMasterRequest, cfg *config.DBConfig) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, cfg.MaxIdleConnectionTimeout)
	defer cancel()

	const query string = `UPDATE users_cities_links SET city_id = $1 WHERE user_id = $2;`

	tag, err := db.db.Exec(timeoutCtx, query, request.CityID, request.UserID)
	if err != nil {
		logger.Error().Err(err).Msg("failed postgresql.UpdateTournamentMaster")
		return DecodeDatabaseError(err)
	}
	if tag.RowsAffected() == 0 {
		err = pgx.ErrNoRows
		logger.Error().Err(err).Msg("failed postgresql.UpdateTournamentMaster")
		return DecodeDatabaseError(err)
	}

	return nil
}

func (db *RDBOperation) GetTournamentMasterByUserId(logger zerolog.Logger, ctx context.Context, userId int64) (entities.TournamentMaster, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, db.cfg.MaxIdleConnectionTimeout)
	defer cancel()

	const query string = `
		SELECT 
			u.id, 
			u.email,
			r.id,
			r.name,
			r.description,
			c.id,
			c.name,
			c.ru
		FROM users as u
			JOIN user_roles as r ON u.role_id = r.id
			JOIN users_cities_links ucl ON u.id = ucl.user_id
			JOIN cities AS c ON ucl.city_id = c.id
		WHERE u.id = $1 AND c.deleted_at IS NULL;`

	var tm entities.TournamentMaster
	tm.User = entities.User{Role: &entities.Role{}}

	err := db.db.QueryRow(timeoutCtx, query, userId).Scan(
		&tm.User.ID, &tm.User.Email, &tm.User.Role.ID, &tm.User.Role.Name, &tm.User.Role.Description, &tm.City.ID, &tm.City.Name, &tm.City.Ru)
	if err != nil {
		logger.Error().Err(err).Msg("failed postgresql.GetTournamentMasterByUserId")
		return entities.TournamentMaster{}, DecodeDatabaseError(err)
	}

	return tm, nil
}

func (db *RDBOperation) GetTournamentMasterListByCityId(logger zerolog.Logger, ctx context.Context, cityID int64, cfg *config.DBConfig) ([]entities.TournamentMaster, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, cfg.MaxIdleConnectionTimeout)
	defer cancel()

	const query string = `
		SELECT 
			u.id, 
			u.email,
			r.id,
			r.name,
			r.description,
			c.id,
			c.name,
			c.ru
		FROM users u
		JOIN user_roles r ON u.role_id = r.id
		JOIN users_cities_links ucl ON u.id = ucl.user_id
		JOIN cities c ON ucl.city_id = c.id
		WHERE c.id = $1 
			AND c.deleted_at IS NULL
			AND r.name = $2;`

	var tMasters []entities.TournamentMaster

	rows, err := db.db.Query(timeoutCtx, query, cityID, constant.TournamentMaster)
	if err != nil {
		logger.Error().Err(err).Msg("failed postgresql.GetTournamentMasterByCityId")
		return nil, DecodeDatabaseError(err)
	}

	for rows.Next() {
		tm := entities.TournamentMaster{
			User: entities.User{Role: &entities.Role{}},
		}

		err = rows.Scan(
			&tm.User.ID,
			&tm.User.Email,
			&tm.User.Role.ID,
			&tm.User.Role.Name,
			&tm.User.Role.Description,
			&tm.City.ID,
			&tm.City.Name,
			&tm.City.Ru,
		)
		if err != nil {
			logger.Error().Err(err).Msg("failed postgresql.GetTournamentMasterByCityId")
			return nil, DecodeDatabaseError(err)
		}

		tMasters = append(tMasters, tm)
	}

	return tMasters, nil
}

func (db *RDBOperation) GetTournamentMasterList(logger zerolog.Logger, ctx context.Context, cfg *config.DBConfig) ([]entities.TournamentMaster, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, cfg.MaxIdleConnectionTimeout)
	defer cancel()

	const query string = `
		SELECT 
			u.id, 
			u.email,
			r.id,
			r.name,
			r.description,
			c.id,
			c.name,
			c.ru
		FROM users u
		JOIN user_roles r ON u.role_id = r.id
		JOIN users_cities_links ucl ON u.id = ucl.user_id
		JOIN cities c ON ucl.city_id = c.id
		WHERE c.deleted_at IS NULL
			AND r.name = $1;`

	var tMasters []entities.TournamentMaster

	rows, err := db.db.Query(timeoutCtx, query, constant.TournamentMaster)
	if err != nil {
		logger.Error().Err(err).Msg("failed postgresql.GetTournamentMasterList")
		return nil, DecodeDatabaseError(err)
	}

	for rows.Next() {
		tm := entities.TournamentMaster{
			User: entities.User{Role: &entities.Role{}},
		}

		err = rows.Scan(
			&tm.User.ID,
			&tm.User.Email,
			&tm.User.Role.ID,
			&tm.User.Role.Name,
			&tm.User.Role.Description,
			&tm.City.ID,
			&tm.City.Name,
			&tm.City.Ru,
		)
		if err != nil {
			logger.Error().Err(err).Msg("failed postgresql.GetTournamentMasterList")
			return nil, DecodeDatabaseError(err)
		}

		tMasters = append(tMasters, tm)
	}

	return tMasters, nil
}

func (db *RWDBOperation) DeleteTournamentMaster(logger zerolog.Logger, ctx context.Context, userId int64, cfg *config.DBConfig) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, cfg.MaxIdleConnectionTimeout)
	defer cancel()

	const query1 string = `DELETE FROM users_cities_links WHERE user_id = $1;`

	const query2 string = `DELETE FROM users WHERE id = $1;`

	tx, err := db.db.Begin(timeoutCtx)
	if err != nil {
		logger.Error().Err(err).Msg("failed postgresql.DeleteTournamentMaster")
		return DecodeDatabaseError(err)
	}

	tag1, err := tx.Exec(timeoutCtx, query1, userId)
	if err != nil {
		logger.Error().Err(err).Msg("failed postgresql.DeleteTournamentMaster")
		_ = tx.Rollback(timeoutCtx)
		return DecodeDatabaseError(err)
	}

	tag2, err := tx.Exec(timeoutCtx, query2, userId)
	if err != nil {
		logger.Error().Err(err).Msg("failed postgresql.DeleteTournamentMaster")
		_ = tx.Rollback(timeoutCtx)
		return DecodeDatabaseError(err)
	}

	err = tx.Commit(timeoutCtx)
	if err != nil {
		logger.Error().Err(err).Msg("failed postgresql.DeleteTournamentMaster")
		_ = tx.Rollback(timeoutCtx)
		return DecodeDatabaseError(err)
	}

	if tag1.RowsAffected() == 0 && tag2.RowsAffected() == 0 {
		err = errors.New("nothing to delete in postgresql.DeleteTournamentMaster")
		logger.Error().Err(err).Msg("failed postgresql.DeleteTournamentMaster")
		return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	return nil
}
