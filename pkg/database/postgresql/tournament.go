package postgresql

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/config"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/database/postgresql/tx"
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

func (db *RWDBOperation) CreateTournament(logger zerolog.Logger, ctx context.Context, request entities.CreateTournamentRequest, tx tx.ITx) (int64, error) {
	timeout, cancel := context.WithTimeout(ctx, db.cfg.MaxIdleConnectionTimeout)
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

	exec := poolOrTx(db.db, tx)

	err = exec.QueryRow(timeout, query1,
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

	for _, t := range request.TeamsIDs {
		_, err = exec.Exec(timeout, query2, t, id)
		if err != nil {
			logger.Error().Err(err).Msg("Failed tx.Exec postgresql.CreateTournament")
			return 0, DecodeDatabaseError(err)
		}
	}

	return id, nil
}

func (db *RWDBOperation) UpdateTournament(logger zerolog.Logger, ctx context.Context, request entities.UpdateTournamentRequest, tx tx.ITx) error {
	timeout, cancel := context.WithTimeout(ctx, db.cfg.MaxIdleConnectionTimeout)
	defer cancel()

	exec := poolOrTx(db.db, tx)

	const query1 string = `
		UPDATE tournaments
		SET
			type_id = COALESCE($2, type_id),
			name = COALESCE($3, name),
			rules = COALESCE($4, rules),
			city_id = COALESCE($5, city_id),
			season_id = COALESCE($6, season_id),
			updated_at = NOW()
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

	_, err = exec.Exec(timeout, query1,
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

func (db *RWDBOperation) UpdateTournamentTeamsLinks(logger zerolog.Logger, ctx context.Context, teamIDs []int64, tournamentId int64, tx tx.ITx) error {
	timeout, cancel := context.WithTimeout(ctx, db.cfg.MaxIdleConnectionTimeout)
	defer cancel()

	const query1 string = `DELETE FROM tournaments_teams_link WHERE tournament_id = $1;`
	const query2 string = `INSERT INTO tournaments_teams_link (team_id, tournament_id) VALUES ($1, $2);`

	exec := poolOrTx(db.db, tx)

	if teamIDs != nil {
		_, err := exec.Exec(timeout, query1, tournamentId)
		if err != nil {
			logger.Error().Err(err).Msg("Failed UpdateTournamentTeamsLinks")
			return DecodeDatabaseError(err)
		}

		for _, id := range teamIDs {
			_, err = exec.Exec(timeout, query2, id, tournamentId)
			if err != nil {
				logger.Error().Err(err).Msg("Failed UpdateTournamentTeamsLinks")
				return DecodeDatabaseError(err)
			}
		}
	}

	return nil
}

func (db *RWDBOperation) DeleteTournament(logger zerolog.Logger, ctx context.Context, tournamentId int64, tx tx.ITx) error {
	timeout, cancel := context.WithTimeout(ctx, db.cfg.MaxIdleConnectionTimeout)
	defer cancel()

	const query string = `DELETE FROM tournaments WHERE id = $1;`

	_, err := poolOrTx(db.db, tx).Exec(timeout, query, tournamentId)
	if err != nil {
		logger.Error().Err(err).Msg("Exec postgresql.DeleteTournament")
		return DecodeDatabaseError(err)
	}

	return nil
}

func (db *RWDBOperation) DeleteTournamentGamesTeamLinks(logger zerolog.Logger, ctx context.Context, tournamentId int64, tx tx.ITx) error {
	timeout, cancel := context.WithTimeout(ctx, db.cfg.MaxIdleConnectionTimeout)
	defer cancel()

	exec := poolOrTx(db.db, tx)

	const query1 string = `DELETE FROM tournaments_teams_link WHERE tournament_id = $1;`
	const query2 string = `DELETE FROM games WHERE stage_id IN (SELECT id FROM tournament_stages WHERE tournament_id = $1);`

	_, err := exec.Exec(timeout, query1, tournamentId)
	if err != nil {
		logger.Error().Err(err).Msg("Failed DeleteTournamentGamesTeamLinks")
		return DecodeDatabaseError(err)
	}

	_, err = exec.Exec(timeout, query2, tournamentId)
	if err != nil {
		logger.Error().Err(err).Msg("Failed DeleteTournamentGamesTeamLinks")
		return DecodeDatabaseError(err)
	}

	return nil
}

func (db *RWDBOperation) CreateTournamentStage(logger zerolog.Logger, ctx context.Context, tournamentID int64, tx tx.ITx) (int64, error) {
	timeout, cancel := context.WithTimeout(ctx, db.cfg.MaxIdleConnectionTimeout)
	defer cancel()

	const query string = `INSERT INTO tournament_stages(tournament_id, is_finished) VALUES ($1, FALSE) RETURNING id;`

	var id int64

	err := poolOrTx(db.db, tx).QueryRow(timeout, query, tournamentID).Scan(&id)
	if err != nil {
		logger.Error().Err(err).Msg("Failed QueryRow postgresql.CreateTournamentStage")
		return 0, DecodeDatabaseError(err)
	}

	return id, nil
}

func (db *RDBOperation) GetTournamentStageGames(logger zerolog.Logger, ctx context.Context, stageId int64, cfg *config.DBConfig) ([]entities.TournamentGame, error) {
	timeout, cancel := context.WithTimeout(ctx, cfg.MaxIdleConnectionTimeout)
	defer cancel()

	const query string = `
		SELECT 
			id, 
			city_id, 
			place_id,
			date,
			team1_id,
			team2_id,
			tech_loose_team_id,
			is_tiebreak,
			stage_id
		FROM games WHERE stage_id = $1;`

	var games []entities.TournamentGame

	rows, err := db.db.Query(timeout, query, stageId)
	if err != nil {
		logger.Error().Err(err).Msg("Failed Query postgresql.GetStageGames")
		return nil, DecodeDatabaseError(err)
	}
	defer rows.Close()

	for rows.Next() {
		var g entities.TournamentGame
		err = rows.Scan(&g.ID, &g.CityID, &g.PlaceID, &g.Date, &g.Team1ID, &g.Team2ID, &g.TechLooseTeamID, &g.IsTiebreak, &g.StageID)
		if err != nil {
			logger.Error().Err(err).Msg("Failed rows.Scan")
			return nil, DecodeDatabaseError(err)
		}

		games = append(games, g)
	}

	return games, nil
}

func (db *RWDBOperation) UpdateTournamentStage(logger zerolog.Logger, ctx context.Context, stage entities.NullableStage) error {
	timeout, cancel := context.WithTimeout(ctx, db.cfg.MaxIdleConnectionTimeout)
	defer cancel()

	const query string = `UPDATE tournament_stages
		SET
			tournament_id = COALESCE($2, tournament_id),
			is_finished = COALESCE($3, is_finished)
		WHERE id = $1;`

	tag, err := db.db.Exec(timeout, query, stage.ID, stage.TournamentID, stage.IsFinished)
	if err != nil {
		logger.Error().Err(err).Msg("Failed tx.Exec postgresql.UpdateTournamentStage")
		return DecodeDatabaseError(err)
	}

	if tag.RowsAffected() == 0 {
		err = pgx.ErrNoRows
		logger.Error().Err(err).Msg("Failed tx.Exec postgresql.UpdateTournamentStage")
		return DecodeDatabaseError(err)
	}

	return nil
}

func (db *RWDBOperation) UnlinkTeamsFromTournament(logger zerolog.Logger, ctx context.Context, tournamentID int64, tx tx.ITx) error {
	timeout, cancel := context.WithTimeout(ctx, db.cfg.MaxIdleConnectionTimeout)
	defer cancel()

	const query string = `DELETE FROM tournaments_teams_link WHERE tournament_id = $1;`

	exec := poolOrTx(db.db, tx)

	_, err := exec.Exec(timeout, query, tournamentID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed Exec UnlinkTeamsFromTournament")
		return DecodeDatabaseError(err)
	}

	return nil
}

func (db *RWDBOperation) DeleteTournamentGames(logger zerolog.Logger, ctx context.Context, tournamentId int64, tx tx.ITx) error {
	timeout, cancel := context.WithTimeout(ctx, db.cfg.MaxIdleConnectionTimeout)
	defer cancel()

	const query string = `DELETE FROM games WHERE stage_id IN (SELECT id FROM tournament_stages WHERE tournament_id = $1);`

	_, err := poolOrTx(db.db, tx).Exec(timeout, query, tournamentId)
	if err != nil {
		logger.Error().Err(err).Msg("Failed Exec DeleteTournamentGames")
		return DecodeDatabaseError(err)
	}

	return nil
}

func (db *RWDBOperation) DeleteTournamentStages(logger zerolog.Logger, ctx context.Context, tournamentID int64, tx tx.ITx) error {
	timeout, cancel := context.WithTimeout(ctx, db.cfg.MaxIdleConnectionTimeout)
	defer cancel()

	const query string = `DELETE FROM tournament_stages WHERE tournament_id = $1;`

	_, err := poolOrTx(db.db, tx).Exec(timeout, query, tournamentID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed Exec DeleteTournamentStage")
		return DecodeDatabaseError(err)
	}

	return nil
}

func (db *RDBOperation) GetTournamentStageList(logger zerolog.Logger, ctx context.Context, tournamentId int64) ([]entities.TournamentStage, error) {
	timeout, cancel := context.WithTimeout(ctx, db.cfg.MaxIdleConnectionTimeout)
	defer cancel()

	const query string = `SELECT id, tournament_id, is_finished FROM tournament_stages WHERE tournament_id = $1;`

	var stages []entities.TournamentStage

	rows, err := db.db.Query(timeout, query, tournamentId)
	if err != nil {
		logger.Error().Err(err).Msg("Failed tx.Exec postgresql.GetTournamentStageList")
		return nil, DecodeDatabaseError(err)
	}
	defer rows.Close()

	for rows.Next() {
		var s entities.TournamentStage
		err = rows.Scan(&s.ID, &s.TournamentID, &s.IsFinished)
		if err != nil {
			logger.Error().Err(err).Msg("Failed rows.Scan")
			return nil, DecodeDatabaseError(err)
		}

		stages = append(stages, s)
	}

	return stages, nil
}

func (db *RWDBOperation) CreateTournamentTx(logger zerolog.Logger, ctx context.Context, request entities.CreateTournamentRequest, tx tx.ITx) (int64, error) {
	timeout, cancel := context.WithTimeout(ctx, db.cfg.MaxIdleConnectionTimeout)
	defer cancel()

	exec := poolOrTx(db.db, tx)

	const query1 string = `
		INSERT INTO tournaments (type_id, name, rules, city_id, season_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id;`

	ruleJSON, err := json.Marshal(request.Rules)
	if err != nil {
		logger.Error().Err(err).Msg("Failed json.Marshal postgresql.CreateTournamentTx")
		return 0, errtmpl.New(err.Error(), err, codes.Internal, http.StatusInternalServerError)
	}

	var id int64

	err = exec.QueryRow(timeout, query1,
		request.TournamentTypeID,
		request.Name,
		ruleJSON,
		request.CityID,
		request.SeasonID,
	).Scan(&id)
	if err != nil {
		logger.Error().Err(err).Msg("Failed Scan postgresql.CreateTournamentTx")
		return 0, DecodeDatabaseError(err)
	}

	return id, nil
}

func (db *RWDBOperation) AddTeamsToTournamentTx(logger zerolog.Logger, ctx context.Context, teams []int64, tournamentId int64, tx tx.ITx) error {
	timeout, cancel := context.WithTimeout(ctx, db.cfg.MaxIdleConnectionTimeout)
	defer cancel()

	exec := poolOrTx(db.db, tx)

	const query2 string = `
		INSERT INTO tournaments_teams_link (team_id, tournament_id)
		VALUES ($1, $2);`

	for _, t := range teams {
		_, err := exec.Exec(timeout, query2, t, tournamentId)
		if err != nil {
			logger.Error().Err(err).Msg("Failed tx.Exec postgresql.AddTeamsToTournamentTx")
			return DecodeDatabaseError(err)
		}
	}

	return nil
}

func (db *RWDBOperation) CreateTournamentStageTx(logger zerolog.Logger, ctx context.Context, tournamentID int64, tx tx.ITx) (int64, error) {
	timeout, cancel := context.WithTimeout(ctx, db.cfg.MaxIdleConnectionTimeout)
	defer cancel()

	const query string = `INSERT INTO tournament_stages(tournament_id, is_finished) VALUES ($1, FALSE) RETURNING id;`

	var id int64

	exec := poolOrTx(db.db, tx)

	err := exec.QueryRow(timeout, query, tournamentID).Scan(&id)
	if err != nil {
		logger.Error().Err(err).Msg("Failed QueryRow postgresql.CreateTournamentStageTx")
		return 0, DecodeDatabaseError(err)
	}

	return id, nil
}

func (db *RDBOperation) GetTournamentList(logger zerolog.Logger, ctx context.Context, req entities.GetTournamentListRequest) ([]entities.TournamentShort, int64, error) {
	timeout, cancel := context.WithTimeout(ctx, db.cfg.MaxIdleConnectionTimeout)
	defer cancel()

	const query string = `
		WITH filtered_tournaments AS (
			SELECT t.id, t.type_id, t.name, t.rules, t.city_id, t.season_id,
				   COUNT(*) OVER() as total_count
			FROM tournaments t
			WHERE 
				($1::INTEGER IS NULL OR t.id = $1) AND
				($2::INTEGER IS NULL OR t.city_id = $2) AND
				($3::INTEGER IS NULL OR t.type_id = $3) AND
				($4::INTEGER IS NULL OR t.season_id = $4)
			ORDER BY t.id DESC
		)
		SELECT id, type_id, name, rules, city_id, season_id, total_count FROM filtered_tournaments;`

	rows, err := db.db.Query(timeout, query, req.TournamentID, req.CityID, req.TournamentTypeID, req.SeasonID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed GetTournamentList")
		return nil, 0, DecodeDatabaseError(err)
	}
	defer rows.Close()

	var tournaments []entities.TournamentShort
	var totalCount int64

	for rows.Next() {
		var t entities.TournamentShort

		err = rows.Scan(&t.ID, &t.TypeID, &t.Name, &t.Rules, &t.CityID, &t.SeasonID, &totalCount)
		if err != nil {
			logger.Error().Err(err).Msg("Failed GetTournamentList")
			return nil, 0, DecodeDatabaseError(err)
		}

		tournaments = append(tournaments, t)
	}

	return tournaments, totalCount, nil
}

func (db *RDBOperation) GetTournamentListByPlayerID(logger zerolog.Logger, ctx context.Context, playerId int64) ([]entities.PlayersTournament, error) {
	timeout, cancel := context.WithTimeout(ctx, db.cfg.MaxIdleConnectionTimeout)
	defer cancel()

	const query string = `
		SELECT tts.id, tts.name, tts.city_id, r.value
  		FROM public.players p
  			JOIN public.players_teams_links ptl ON p.id = ptl.player_id
  			JOIN public.teams t ON ptl.team_id = t.id
  			JOIN public.tournaments_teams_link tll ON t.id = tll.team_id
  			JOIN public.tournaments tts ON tll.tournament_id = tts.id
  			LEFT JOIN public.tournament_rating r ON p.id = r.player_id AND r.tournament_id = tts.id 
  		WHERE p.id = $1`

	rows, err := db.db.Query(timeout, query, playerId)
	if err != nil {
		logger.Error().Err(err).Msg("Failed GetTournamentListByPlayerID")
		return nil, DecodeDatabaseError(err)
	}
	defer rows.Close()

	var tournaments []entities.PlayersTournament

	for rows.Next() {
		var t entities.PlayersTournament

		err = rows.Scan(&t.ID, &t.Name, &t.CityID, &t.Rating)
		if err != nil {
			logger.Error().Err(err).Msg("Failed GetTournamentListByPlayerID")
			return nil, DecodeDatabaseError(err)
		}

		tournaments = append(tournaments, t)
	}

	return tournaments, nil
}
