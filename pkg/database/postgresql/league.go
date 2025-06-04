package postgresql

import (
	"context"
	stderr "errors"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"

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

	err = tx.QueryRow(ctx, queryCreateLeague, league.Name, league.CityID, league.SeasonID).Scan(&id)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to create League record")
		_ = tx.Rollback(ctx)
		return 0, DecodeDatabaseError(err)
	}

	for _, teamID := range teams {
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

func (db *RWDBOperation) UpdateLeague(logger zerolog.Logger, ctx context.Context, league entity.League, teams []int64) error {
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
