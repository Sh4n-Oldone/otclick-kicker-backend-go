package postgresql

import (
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"net/http"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/config"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	errtmpl "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/error_templates"
)

const (
	numberOfTypes = 5
)

func (db *RDBOperation) GetTournamentTypeList(logger zerolog.Logger, ctx context.Context, withDeleted bool, cfg *config.DBConfig) ([]entities.TournamentType, error) {
	timeout, cancel := context.WithTimeout(ctx, cfg.MaxIdleConnectionTimeout)
	defer cancel()

	const query string = `
		SELECT id, name, description, deleted_at
		FROM tournament_types
		WHERE $1 = TRUE OR deleted_at IS NULL
		ORDER BY id;`

	rows, err := db.db.Query(timeout, query, withDeleted)
	if err != nil {
		logger.Error().Err(err).Msg("Failed postgresql.GetTournamentTypeList")
		return nil, DecodeDatabaseError(err)
	}
	defer rows.Close()

	types := make([]entities.TournamentType, 0, numberOfTypes)

	for rows.Next() {
		var t entities.TournamentType

		err = rows.Scan(&t.ID, &t.Name, &t.Description, &t.DeletedAt)
		if err != nil {
			logger.Error().Err(err).Msg("Failed postgresql.GetTournamentTypeList")
			return nil, DecodeDatabaseError(err)
		}

		types = append(types, t)
	}

	return types, nil
}

func (db *RDBOperation) GetTournamentById(logger zerolog.Logger, ctx context.Context, tournamentId int64, cfg *config.DBConfig) (entities.Tournament, error) {
	timeout, cancel := context.WithTimeout(ctx, cfg.MaxIdleConnectionTimeout)
	defer cancel()

	var t entities.Tournament

	const query1 string = `SELECT id, type_id, name, rules, city_id, season_id FROM tournaments	WHERE id = $1;`
	const query2 string = `SELECT id, tournament_id, is_finished FROM tournament_stages WHERE tournament_id = $1;`
	const query3 string = `SELECT team_id FROM tournaments_teams_link WHERE tournament_id = $1;`

	var rulesBytes []byte

	err := db.db.QueryRow(timeout, query1, tournamentId).
		Scan(&t.ID, &t.TypeID, &t.Name, &rulesBytes, &t.CityID, &t.SeasonID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed postgresql.GetTournamentById")
		return entities.Tournament{}, DecodeDatabaseError(err)
	}

	err = json.Unmarshal(rulesBytes, &t.Rules)
	if err != nil {
		logger.Error().Err(err).Msg("Failed json.Unmarshal postgresql.GetTournamentById")
		return entities.Tournament{}, DecodeDatabaseError(err)
	}

	rows, err := db.db.Query(timeout, query2, tournamentId)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get stages postgresql.GetTournamentById")
		return entities.Tournament{}, DecodeDatabaseError(err)
	}
	defer rows.Close()

	stages := make([]entities.TournamentStage, 0)

	for rows.Next() {
		var ts entities.TournamentStage
		err = rows.Scan(&ts.ID, &ts.TournamentID, &ts.IsFinished)
		if err != nil {
			logger.Error().Err(err).Msg("Failed rows.Scan postgresql.GetTournamentById")
		}
		stages = append(stages, ts)
	}

	t.Stages = stages

	rows, err = db.db.Query(timeout, query3, tournamentId)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get teamIDs postgresql.GetTournamentById")
		return entities.Tournament{}, DecodeDatabaseError(err)
	}

	teamIDs := make([]int64, 0)

	for rows.Next() {
		var id int64
		err = rows.Scan(&id)
		if err != nil {
			logger.Error().Err(err).Msg("Failed rows.Scan postgresql.GetTournamentById")
			return entities.Tournament{}, DecodeDatabaseError(err)
		}
		teamIDs = append(teamIDs, id)
	}

	t.TeamIDs = teamIDs

	return t, nil
}

func (db *RDBOperation) GetTournamentStage(logger zerolog.Logger, ctx context.Context, stageId int64, cfg *config.DBConfig) (entities.TournamentStage, error) {
	timeout, cancel := context.WithTimeout(ctx, cfg.MaxIdleConnectionTimeout)
	defer cancel()

	var s entities.TournamentStage

	const query string = `SELECT id, tournament_id, is_finished FROM tournament_stages WHERE id = $1;`

	err := db.db.QueryRow(timeout, query, stageId).Scan(&s.ID, &s.TournamentID, &s.IsFinished)
	if err != nil {
		logger.Error().Err(err).Msg("Failed QueryRow postgresql.GetTournamentStage")
		return entities.TournamentStage{}, DecodeDatabaseError(err)
	}

	return s, nil
}

func (db *RWDBOperation) CreateRegularTournament(logger zerolog.Logger, ctx context.Context, request *entities.CreateTournamentRequest, cfg *config.DBConfig) (int64, error) {
	timeout, cancel := context.WithTimeout(ctx, cfg.MaxIdleConnectionTimeout)
	defer cancel()

	const query1 string = `
		INSERT INTO tournaments (type_id, name, rules, city_id, season_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id;`

	const query2 string = `
		INSERT INTO tournaments_teams_link (team_id, tournament_id)
		VALUES ($1, $2);`

	ruleJSON, err := json.Marshal(request.Rules)
	if err != nil {
		logger.Error().Err(err).Msg("Failed json.Marshal postgresql.CreateTournament")
		return 0, errtmpl.New(err.Error(), err, codes.Internal, http.StatusInternalServerError)
	}

	var id int64

	tx, err := db.db.Begin(timeout)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to create tx")
		return 0, DecodeDatabaseError(err)
	}

	err = tx.QueryRow(timeout, query1,
		request.TournamentTypeID,
		request.Name,
		ruleJSON,
		request.CityID,
		request.SeasonID,
	).Scan(&id)
	if err != nil {
		logger.Error().Err(err).Msg("Failed Scan postgresql.CreateTournament")
		return 0, DecodeDatabaseError(err)
	}

	for _, t := range request.TeamIDs {
		_, err = tx.Exec(timeout, query2, t, id)
		if err != nil {
			logger.Error().Err(err).Msg("Failed tx.Exec postgresql.CreateTournament")
			_ = tx.Rollback(timeout)
			return 0, DecodeDatabaseError(err)
		}
	}

	err = tx.Commit(timeout)
	if err != nil {
		logger.Error().Err(err).Msg("Failed tx.Commit postgresql.CreateTournament")
		return 0, DecodeDatabaseError(err)
	}

	return id, nil
}

func (db *RWDBOperation) UpdateTournament(logger zerolog.Logger, ctx context.Context, request *entities.UpdateTournamentRequest, cfg *config.DBConfig) error {
	timeout, cancel := context.WithTimeout(ctx, cfg.MaxIdleConnectionTimeout)
	defer cancel()

	const query1 string = `
		UPDATE tournaments
		SET
			type_id = COALESCE($2, type_id),
			name = COALESCE($3, name),
			rules = COALESCE($4, rules),
			city_id = COALESCE($5, city_id),
			season_id = COALESCE($6, season_id)
		WHERE id = $1;`

	var ruleJSON []byte = nil
	var err error

	if request.Rules != nil {
		ruleJSON, err = json.Marshal(request.Rules)
		if err != nil {
			logger.Error().Err(err).Msg("Failed json.Marshal postgresql.UpdateTournament")
			return errtmpl.New(err.Error(), err, codes.Internal, http.StatusInternalServerError)
		}
	}

	_, err = db.db.Exec(timeout, query1,
		request.ID,
		request.TournamentTypeID,
		request.Name,
		ruleJSON,
		request.CityID,
		request.SeasonID,
	)
	if err != nil {
		logger.Error().Err(err).Msg("Failed Exec postgresql.UpdateTournament")
		return DecodeDatabaseError(err)
	}

	return nil
}

func (db *RWDBOperation) UpdateTournamentTeamsLinks(logger zerolog.Logger, ctx context.Context, teamIDs []int64, tournamentId int64, cfg *config.DBConfig) error {
	timeout, cancel := context.WithTimeout(ctx, cfg.MaxIdleConnectionTimeout)
	defer cancel()

	const query1 string = `DELETE FROM tournaments_teams_link WHERE tournament_id = $1;`
	const query2 string = `INSERT INTO tournaments_teams_link (team_id, tournament_id) VALUES ($1, $2);`

	tx, err := db.db.Begin(timeout)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to create tx")
		return DecodeDatabaseError(err)
	}

	if teamIDs != nil {
		_, err = tx.Exec(timeout, query1, tournamentId)
		if err != nil {
			logger.Error().Err(err).Msg("Failed tx.Exec postgresql.UpdateTournament")
			_ = tx.Rollback(timeout)
			return DecodeDatabaseError(err)
		}

		for _, id := range teamIDs {
			_, err = tx.Exec(timeout, query2, id, tournamentId)
			if err != nil {
				logger.Error().Err(err).Msg("Failed tx.Exec postgresql.UpdateTournament")
				_ = tx.Rollback(timeout)
				return DecodeDatabaseError(err)
			}
		}
	}
	err = tx.Commit(timeout)
	if err != nil {
		logger.Error().Err(err).Msg("Failed tx.Commit postgresql.UpdateTournament")
		_ = tx.Rollback(timeout)
		return DecodeDatabaseError(err)
	}

	return nil
}

func (db *RWDBOperation) DeleteTournamentCascade(logger zerolog.Logger, ctx context.Context, id int64, cfg *config.DBConfig) error {
	timeout, cancel := context.WithTimeout(ctx, cfg.MaxIdleConnectionTimeout)
	defer cancel()

	const query1 string = `DELETE FROM tournaments_teams_link WHERE tournament_id = $1;`
	const query2 string = `DELETE FROM games WHERE stage_id IN (SELECT id FROM tournament_stages WHERE tournament_id = $1);`
	const query3 string = `DELETE FROM tournament_stages WHERE tournament_id = $1;`
	const query4 string = `DELETE FROM tournaments WHERE id = $1;`

	tx, err := db.db.Begin(timeout)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to create tx")
		return DecodeDatabaseError(err)
	}

	_, err = tx.Exec(timeout, query1, id)
	if err != nil {
		logger.Error().Err(err).Msg("Failed tx.Exec postgresql.DeleteTournament")
		_ = tx.Rollback(timeout)
		return DecodeDatabaseError(err)
	}

	_, err = tx.Exec(timeout, query2, id)
	if err != nil {
		logger.Error().Err(err).Msg("Failed tx.Exec postgresql.DeleteTournament")
		_ = tx.Rollback(timeout)
		return DecodeDatabaseError(err)
	}

	_, err = tx.Exec(timeout, query3, id)
	if err != nil {
		logger.Error().Err(err).Msg("Failed tx.Exec postgresql.DeleteTournament")
		_ = tx.Rollback(timeout)
		return DecodeDatabaseError(err)
	}

	tag, err := tx.Exec(timeout, query4, id)
	if err != nil {
		logger.Error().Err(err).Msg("Failed tx.Exec postgresql.DeleteTournament")
		_ = tx.Rollback(timeout)
		return DecodeDatabaseError(err)
	}

	err = tx.Commit(timeout)
	if err != nil {
		logger.Error().Err(err).Msg("Failed tx.Commit postgresql.DeleteTournament")
		_ = tx.Rollback(timeout)
		return DecodeDatabaseError(err)
	}

	if tag.RowsAffected() == 0 {
		err = pgx.ErrNoRows
		logger.Error().Err(err).Msg("no rows affected postgresql.DeleteTournament")
		return DecodeDatabaseError(err)
	}

	return nil
}

func (db *RWDBOperation) DeleteTournamentGamesTeamLinks(logger zerolog.Logger, ctx context.Context, tournamentId int64, cfg *config.DBConfig) error {
	timeout, cancel := context.WithTimeout(ctx, cfg.MaxIdleConnectionTimeout)
	defer cancel()

	const query1 string = `DELETE FROM tournaments_teams_link WHERE tournament_id = $1;`
	const query2 string = `DELETE FROM games WHERE stage_id IN (SELECT id FROM tournament_stages WHERE tournament_id = $1);`

	tx, err := db.db.Begin(timeout)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to create tx")
		return DecodeDatabaseError(err)
	}

	_, err = tx.Exec(timeout, query1, tournamentId)
	if err != nil {
		logger.Error().Err(err).Msg("Failed tx.Exec postgresql.DeleteTournament")
		_ = tx.Rollback(timeout)
		return DecodeDatabaseError(err)
	}

	_, err = tx.Exec(timeout, query2, tournamentId)
	if err != nil {
		logger.Error().Err(err).Msg("Failed tx.Exec postgresql.DeleteTournament")
		_ = tx.Rollback(timeout)
		return DecodeDatabaseError(err)
	}

	err = tx.Commit(timeout)
	if err != nil {
		logger.Error().Err(err).Msg("Failed tx.Commit postgresql.DeleteTournament")
		_ = tx.Rollback(timeout)
		return DecodeDatabaseError(err)
	}

	//todo: что с рейтингами?

	return nil
}

func (db *RWDBOperation) CreateTournamentStage(logger zerolog.Logger, ctx context.Context, tournamentID int64, isFinished bool, cfg *config.DBConfig) (int64, error) {
	timeout, cancel := context.WithTimeout(ctx, cfg.MaxIdleConnectionTimeout)
	defer cancel()

	const query string = `INSERT INTO tournament_stages(tournament_id, is_finished) VALUES ($1, $2) RETURNING id;`

	var id int64

	err := db.db.QueryRow(timeout, query, tournamentID, isFinished).Scan(&id)
	if err != nil {
		logger.Error().Err(err).Msg("Failed QueryRow postgresql.CreateTournamentStage")
		return 0, DecodeDatabaseError(err)
	}

	return id, nil
}
