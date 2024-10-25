package postgresql

import (
	"context"
	stderr "errors"
	"github.com/rs/zerolog"
	"strconv"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/entity"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/errors"
)

func (db *RDBOperation) GetLeagueList(logger zerolog.Logger, ctx context.Context, cityID int64) ([]entity.League, error) {
	query := queryGetLeagueList

	rows, err := db.db.Query(ctx, query, cityID)

	if err != nil {
		logger.Error().Err(err).Msg("")
		return nil, DecodeDatabaseError(err)
	}
	defer rows.Close()

	var leagues []entity.League

	for rows.Next() {
		var league entity.League

		err = rows.Scan(&league.ID, &league.Name, &league.CityID)
		if err != nil {
			logger.Error().Err(err).Msg("failed to scan league list")
			return nil, DecodeDatabaseError(err)
		}

		leagues = append(leagues, league)
	}

	return leagues, nil
}

func (db *RWDBOperation) CreateLeague(logger zerolog.Logger, ctx context.Context, league entity.League, teams []int64) (int64, error) {
	var id int64

	tx, err := db.db.Begin(ctx)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to create League record")
		return 0, DecodeDatabaseError(err)
	}

	var teamLeagueID *int64
	for _, teamID := range teams {
		err := tx.QueryRow(ctx, queryGetTeamLeagueID, teamID).Scan(&teamLeagueID)
		if err != nil {
			logger.Error().Stack().Err(err).Msg("failed to update League record")
			_ = tx.Rollback(ctx)
			return 0, DecodeDatabaseError(err)
		}
		if teamLeagueID != nil {
			logger.Error().Stack().Err(err).Msg("Failed to create League record, team " + strconv.FormatInt(teamID, 10) + " is already in another league")
			_ = tx.Rollback(ctx)
			return 0, stderr.New(errors.ErrTeamAlreadyInLeague)
		}
	}

	err = tx.QueryRow(ctx, queryCreateLeague, league.Name, league.CityID).Scan(&id)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to create League record")
		_ = tx.Rollback(ctx)
		return 0, DecodeDatabaseError(err)
	}

	for _, teamID := range teams {
		_, err = tx.Exec(ctx, queryUpdateTeamsLeagueID, id, teamID)
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

func (db *RWDBOperation) UpdateLeague(logger zerolog.Logger, ctx context.Context, league entity.League, teams []int64) error {
	tx, err := db.db.Begin(ctx)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to update League record")
		return DecodeDatabaseError(err)
	}

	var teamLeagueID *int64
	for _, teamID := range teams {
		err = tx.QueryRow(ctx, queryGetTeamLeagueID, teamID).Scan(&teamLeagueID)
		if err != nil {
			logger.Error().Stack().Err(err).Msg("failed to update League record")
			_ = tx.Rollback(ctx)
			return DecodeDatabaseError(err)
		}
		if teamLeagueID != nil {
			logger.Error().Stack().Err(err).Msg("Failed to update League record, team " + strconv.FormatInt(teamID, 10) + " is already in another league")
			_ = tx.Rollback(ctx)
			return stderr.New(errors.ErrTeamAlreadyInLeague)
		}
	}

	_, err = tx.Exec(ctx, queryUpdateLeague, league.Name, league.ID)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to update League record")
		_ = tx.Rollback(ctx)
		return DecodeDatabaseError(err)
	}

	for _, teamID := range teams {
		_, err = tx.Exec(ctx, queryDeleteTeamsLeagueID, league.ID)
		if err != nil {
			logger.Error().Stack().Err(err).Msg("failed to update League record")
			_ = tx.Rollback(ctx)
			return DecodeDatabaseError(err)
		}

		_, err = tx.Exec(ctx, queryUpdateTeamsLeagueID, league.ID, teamID)
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

func (db *RWDBOperation) DeleteLeague(logger zerolog.Logger, ctx context.Context, id int64) error {
	_, err := db.db.Exec(ctx, queryDeleteLeague, id)
	if err != nil {
		logger.Error().Err(err).Msg("failed to delete League record")
		return DecodeDatabaseError(err)
	}

	return nil
}

func (db *RDBOperation) GetLeaguesByPlayerID(logger zerolog.Logger, ctx context.Context, playerID int) ([]entities.PlayersLeague, error) {
	const query string = `
		SELECT l.id, l.name, l.city_id, r.value
		FROM public.rating r
		JOIN public.leagues l ON l.id = r.league_id
		WHERE player_id = $1`

	leagues := make([]entities.PlayersLeague, 0)

	rows, err := db.db.Query(ctx, query, playerID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to postgresql.GetLeaguesByPlayerID")
		return nil, DecodeDatabaseError(stderr.New(errors.ErrGetLeagueList))
	}

	for rows.Next() {
		var league entities.PlayersLeague

		err = rows.Scan(&league.ID, &league.Name, &league.CityID, &league.Rating)
		if err != nil {
			logger.Error().Err(err).Msg("failed to postgresql.GetLeaguesByPlayerID")
			return nil, DecodeDatabaseError(stderr.New(errors.ErrGetLeague))
		}

		leagues = append(leagues, league)
	}

	return leagues, nil
}
