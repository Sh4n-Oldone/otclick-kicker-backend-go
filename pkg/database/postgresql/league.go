package postgresql

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/database/postgresql/tx"
	pkgerr "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/errors"
)

// GetLeagueList deprecated
func (db *RDBOperation) GetLeagueList(logger zerolog.Logger, ctx context.Context, cityID int64) ([]entities.League, error) {
	query := queryGetLeagueList

	rows, err := db.db.Query(ctx, query, cityID)

	if err != nil {
		logger.Error().Err(err).Msg("")
		return nil, DecodeDatabaseError(err)
	}
	defer rows.Close()

	var leagues []entities.League

	for rows.Next() {
		var league entities.League

		err = rows.Scan(&league.ID, &league.Name, &league.CityID)
		if err != nil {
			logger.Error().Err(err).Msg("failed to scan league list")
			return nil, DecodeDatabaseError(err)
		}

		leagues = append(leagues, league)
	}

	return leagues, nil
}

// CreateLeague deprecated
func (db *RWDBOperation) CreateLeague(logger zerolog.Logger, ctx context.Context, request *entities.CreateLeagueRequest) (int64, error) {
	var id int64

	tx, err := db.db.Begin(ctx)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to create League record")
		return 0, DecodeDatabaseError(err)
	}

	err = tx.QueryRow(ctx, queryCreateLeague, request.Name, request.CityID, request.SeasonID).Scan(&id)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to create League record")
		_ = tx.Rollback(ctx)
		return 0, DecodeDatabaseError(err)
	}

	for _, teamID := range request.Teams {
		_, err = tx.Exec(ctx, queryUpdateTeamsLeagueID, teamID, id)
		if err != nil {
			logger.Error().Stack().Err(err).Msg("failed to create League record")
			_ = tx.Rollback(ctx)
			return 0, DecodeDatabaseError(err)
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to create League record")
		_ = tx.Rollback(ctx)
		return 0, DecodeDatabaseError(err)
	}

	return id, nil
}

// UpdateLeague deprecated
func (db *RWDBOperation) UpdateLeague(logger zerolog.Logger, ctx context.Context, league entities.League, teams []int64) error {
	tx, err := db.db.Begin(ctx)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to update League record")
		return DecodeDatabaseError(err)
	}

	_, err = tx.Exec(ctx, queryUpdateLeague, league.ID, league.Name, league.SeasonID)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to update League record")
		_ = tx.Rollback(ctx)
		return DecodeDatabaseError(err)
	}

	// убираем связи league_id из всех команд этой лиги
	_, err = tx.Exec(ctx, queryDeleteTeamsLeagueID, league.ID)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to update League record")
		_ = tx.Rollback(ctx)
		return DecodeDatabaseError(err)
	}
	for _, teamID := range teams {
		// добавляем связи league_id в переданные команды
		_, err = tx.Exec(ctx, queryUpdateTeamsLeagueID, teamID, league.ID)
		if err != nil {
			logger.Error().Stack().Err(err).Msg("failed to update League record")
			_ = tx.Rollback(ctx)
			return DecodeDatabaseError(err)
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to update League record")
		_ = tx.Rollback(ctx)
		return DecodeDatabaseError(err)
	}

	return nil
}

// DeleteLeague deprecated
func (db *RWDBOperation) DeleteLeague(logger zerolog.Logger, ctx context.Context, id int64) error {
	tx, err := db.db.Begin(ctx)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to delete League record")
		return DecodeDatabaseError(err)
	}

	// убираем league_id связь из всех команд этой лиги
	_, err = tx.Exec(ctx, queryDeleteTeamsLeagueID, id)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to delete League record")
		_ = tx.Rollback(ctx)
		return DecodeDatabaseError(err)
	}

	// удаляем рейтинг
	_, err = tx.Exec(ctx, queryDeleteRatingByLeagueID, id)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to delete League record")
		_ = tx.Rollback(ctx)
		return DecodeDatabaseError(err)
	}

	res, err := tx.Exec(ctx, queryDeleteLeague, id)
	if err != nil {
		logger.Error().Err(err).Msg("failed to delete League record")
		return DecodeDatabaseError(err)
	}

	if res.RowsAffected() == 0 {
		err = pgx.ErrNoRows
		logger.Error().Stack().Err(err).Msg("failed find to postgresql.DeleteLeague")
		return DecodeDatabaseError(err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to delete League record")
		_ = tx.Rollback(ctx)
		return DecodeDatabaseError(err)
	}

	return nil
}

// GetLeaguesByPlayerID deprecated
func (db *RDBOperation) GetLeaguesByPlayerID(logger zerolog.Logger, ctx context.Context, playerID int) ([]entities.PlayersLeague, error) {
	const query string = `
  		SELECT l.id, l.name, l.city_id, r.value
  		FROM public.players p
  			JOIN public.players_teams_links ptl ON p.id = ptl.player_id
  			JOIN public.teams t ON ptl.team_id = t.id
  			JOIN public.teams_leagues_links tll ON t.id = tll.team_id
  			JOIN public.leagues l ON tll.league_id = l.id
  			LEFT JOIN public.rating r ON p.id = r.player_id AND r.league_id = l.id 
  		WHERE p.id = $1`

	leagues := make([]entities.PlayersLeague, 0)

	rows, err := db.db.Query(ctx, query, playerID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to postgresql.GetLeaguesByPlayerID")
		return nil, DecodeDatabaseError(errors.New(pkgerr.ErrGetLeagueList))
	}

	for rows.Next() {
		var league entities.PlayersLeague

		err = rows.Scan(&league.ID, &league.Name, &league.CityID, &league.Rating)
		if err != nil {
			logger.Error().Err(err).Msg("failed to postgresql.GetLeaguesByPlayerID")
			return nil, DecodeDatabaseError(errors.New(pkgerr.ErrGetLeague))
		}

		leagues = append(leagues, league)
	}

	return leagues, nil
}

// GetLeagueById deprecated
func (db *RDBOperation) GetLeagueById(logger zerolog.Logger, ctx context.Context, leagueId int64, tx tx.ITx) (entities.League, error) {
	timeout, cancel := context.WithTimeout(ctx, db.cfg.MaxIdleConnectionTimeout)
	defer cancel()

	const query string = `SELECT l.id, l.name, l.city_id, l.season_id FROM public.leagues l WHERE l.id = $1;`

	var league entities.League

	err := poolOrTx(db.db, tx).QueryRow(timeout, query, leagueId).
		Scan(&league.ID, &league.Name, &league.CityID, &league.SeasonID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to postgresql.GetLeagueById")
		return entities.League{}, DecodeDatabaseError(err)
	}

	return league, nil
}

func (db *RDBOperation) GetLeagueListToMigrate(ctx context.Context, logger zerolog.Logger) ([]entities.League, error) {
	timeout, cancel := context.WithTimeout(ctx, db.cfg.MaxIdleConnectionTimeout)
	defer cancel()

	const query string = `
		SELECT 
    		id,
    		name,
    		city_id,
    		season_id
		FROM public.leagues AS l;`

	rows, err := db.db.Query(timeout, query)
	if err != nil {
		logger.Error().Err(err).Msg("Failed GetLeagueListToMigrate")
		return nil, DecodeDatabaseError(err)
	}
	defer rows.Close()

	var leagues []entities.League

	for rows.Next() {
		var league entities.League

		err = rows.Scan(&league.ID, &league.Name, &league.CityID, &league.SeasonID)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to scan league")
			return nil, DecodeDatabaseError(err)
		}

		leagues = append(leagues, league)
	}

	return leagues, nil
}

func (db *RDBOperation) GetLeagueGamesToMigrate(ctx context.Context, logger zerolog.Logger, leagueId int64) ([]entities.TournamentGame, error) {
	timeout, cancel := context.WithTimeout(ctx, db.cfg.MaxIdleConnectionTimeout)
	defer cancel()

	const query string = `
		SELECT 
		    g.id,
			g.city_id,
			g.place_id,
			g.date,
			g.team1_id,
			g.team2_id,
			g.updated_at,
			g.tech_loose_team_id,
			g.is_tiebreak
		FROM public.games AS g
		WHERE g.league_id = $1;
	`

	rows, err := db.db.Query(timeout, query, leagueId)
	if err != nil {
		logger.Error().Err(err).Msg("Failed GetLeagueListToMigrate")
		return nil, DecodeDatabaseError(err)
	}
	defer rows.Close()

	var games []entities.TournamentGame

	for rows.Next() {
		var game entities.TournamentGame

		err = rows.Scan(
			&game.ID,
			&game.CityID,
			&game.PlaceID,
			&game.Date,
			&game.Team1ID,
			&game.Team2ID,
			&game.UpdatedAt,
			&game.TechLooseTeamID,
			&game.IsTiebreak,
		)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to scan league game")
			return nil, DecodeDatabaseError(err)
		}

		games = append(games, game)
	}

	return games, nil
}

func (db *RDBOperation) GetExtraPointsByLeagueIdToMigrate(ctx context.Context, logger zerolog.Logger, leagueId int64) ([]entities.ExtraPoints, error) {
	timeout, cancel := context.WithTimeout(ctx, db.cfg.MaxIdleConnectionTimeout)
	defer cancel()

	const query string = `
		SELECT
			tep.team_id,
			tep.reason,
			tep.points
		FROM public.team_extra_points AS tep
		WHERE tep.league_id = $1;
	`

	rows, err := db.db.Query(timeout, query, leagueId)
	if err != nil {
		logger.Error().Err(err).Msg("Failed GetExtraPointsByLeagueIdToMigrate")
		return nil, DecodeDatabaseError(err)
	}
	defer rows.Close()

	var extraPoints []entities.ExtraPoints

	for rows.Next() {
		var extraPoint entities.ExtraPoints

		err = rows.Scan(
			&extraPoint.TeamId,
			&extraPoint.Reason,
			&extraPoint.Points,
		)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to scan league game")
			return nil, DecodeDatabaseError(err)
		}

		extraPoints = append(extraPoints, extraPoint)
	}

	return extraPoints, nil
}
