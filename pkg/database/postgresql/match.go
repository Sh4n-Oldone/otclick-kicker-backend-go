package postgresql

import (
	"context"
	stderr "errors"
	"github.com/rs/zerolog"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/errors"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/entity"
)

func (db *RWDBOperation) CreateMatch(logger zerolog.Logger, ctx context.Context, match entity.Match) (int64, error) {
	var id int64

	err := db.db.QueryRow(ctx, queryCreateMatch,
		match.Date,
		match.GameID,
		match.Team1ID,
		match.Team2ID,
		match.Player1Team1ID,
		match.Player2Team1ID,
		match.Player1Team2ID,
		match.Player2Team2ID,
		match.ScoreTeam1,
		match.ScoreTeam2,
	).Scan(&id)
	if err != nil {
		logger.Error().Err(err).Msg("failed to create Match record")
		return 0, DecodeDatabaseError(err)
	}

	return id, nil
}

func (db *RWDBOperation) UpdateMatch(logger zerolog.Logger, ctx context.Context, match entity.Match) error {
	result, err := db.db.Exec(ctx, queryUpdateMatch,
		match.Date,
		match.GameID,
		match.Team1ID,
		match.Team2ID,
		match.Player1Team1ID,
		match.Player2Team1ID,
		match.Player1Team2ID,
		match.Player2Team2ID,
		match.ScoreTeam1,
		match.ScoreTeam2,
		match.ID,
	)
	if err != nil {
		logger.Error().Err(err).Msg("failed to update match record")
		return DecodeDatabaseError(err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		logger.Error().Err(err).Msg("failed to get affected rows")
		return stderr.New("Failed to update match, it does not exist")
	}

	return nil
}

func (db *RWDBOperation) DeleteMatch(logger zerolog.Logger, ctx context.Context, id int64) (bool, error) {
	result, err := db.db.Exec(ctx, queryDeleteMatch, id)
	if err != nil {
		logger.Error().Err(err).Msg("failed to delete match record")
		return false, DecodeDatabaseError(err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		logger.Error().Err(err).Msg("failed to get affected rows")
		return false, stderr.New("Failed to delete match, it does not exist")
	}

	return true, nil
}

func (db *RDBOperation) GetMatchesByPlayerID(logger zerolog.Logger, ctx context.Context, playerID int) ([]entities.Match, error) {
	query := `
		SELECT id, date, game_id, team1_id, team2_id, player1_team1_id, player2_team1_id, player1_team2_id, player2_team2_id, score_team1, score_team2
		FROM public.matches
		WHERE player1_team1_id = $1
		   OR player2_team1_id = $1
		   OR player1_team2_id = $1
		   OR player2_team2_id = $1
		ORDER BY updated_at;`

	matches := make([]entities.Match, 0)

	rows, err := db.db.Query(ctx, query, playerID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to postgresql.GetMatchesByPlayerID")
		return nil, DecodeDatabaseError(stderr.New(errors.ErrGetMatches))
	}
	defer rows.Close()

	for rows.Next() {
		var m entities.Match
		err = rows.Scan(&m.ID, &m.Date, &m.GameID, &m.Team1ID, &m.Team2ID, &m.Player1Team1ID, &m.Player2Team1ID,
			&m.Player1Team2ID, &m.Player2Team2ID, &m.ScoreTeam1, &m.ScoreTeam2)
		if err != nil {
			logger.Error().Err(err).Msg("failed to postgresql.GetMatchesByPlayerID")
			return nil, DecodeDatabaseError(stderr.New(errors.ErrGetMatches))
		}

		matches = append(matches, m)
	}

	return matches, nil
}

func (db *RDBOperation) GetMatchListByGameID(ctx context.Context, logger zerolog.Logger, gameID int) ([]entities.Match, error) {
	matches := make([]entities.Match, 0)

	rows, err := db.db.Query(ctx, queryGetMatchListByGameID, gameID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to postgresql.GetMatchesByPlayerID")
		return nil, DecodeDatabaseError(stderr.New(errors.ErrGetMatches))
	}
	defer rows.Close()

	for rows.Next() {
		var m entities.Match
		err = rows.Scan(&m.ID,
			&m.Date,
			&m.GameID,
			&m.Team1ID,
			&m.Team2ID,
			&m.Player1Team1ID,
			&m.Player2Team1ID,
			&m.Player1Team2ID,
			&m.Player2Team2ID,
			&m.ScoreTeam1,
			&m.ScoreTeam2,
			&m.Player1Team1RateBefore,
			&m.Player1Team2RateBefore,
			&m.Player2Team1RateBefore,
			&m.Player2Team2RateBefore,
			&m.Player1Team1RateAfter,
			&m.Player1Team2RateAfter,
			&m.Player2Team1RateAfter,
			&m.Player2Team2RateAfter)
		if err != nil {
			logger.Error().Err(err).Msg("failed to postgresql.GetMatchListByGameID")
			return nil, DecodeDatabaseError(stderr.New(errors.ErrGetMatches))
		}

		matches = append(matches, m)
	}

	return matches, nil
}

func (db *RDBOperation) GetMatchListByLeagueID(logger zerolog.Logger, ctx context.Context, leagueID int64) ([]entity.Match, error) {
	query := `
		SELECT m.id,
			m.date,
			m.game_id,
			m.team1_id,
			m.team2_id,
			m.player1_team1_id,
			m.player2_team1_id,
			m.player1_team2_id,
			m.player2_team2_id,
			m.score_team1,
			m.score_team2,
			m.player1_team1_rate_before,
			m.player1_team2_rate_before,
			m.player2_team1_rate_before,
			m.player2_team2_rate_before,
			m.player1_team1_rate_after,
			m.player1_team2_rate_after,
			m.player2_team1_rate_after,
			m.player2_team2_rate_after 
		FROM matches as m
		JOIN games as g ON g.id = m.game_id
		JOIN teams_leagues_links as t1 ON g.team1_id = t1.team_id
		JOIN teams_leagues_links as t2 ON g.team2_id = t2.team_id
		JOIN leagues as l ON t1.league_id = l.id AND t2.league_id = l.id
		WHERE l.id = $1 
		ORDER BY g.date, m.date, m.updated_at;`

	matches := make([]entity.Match, 0)

	rows, err := db.db.Query(ctx, query, leagueID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to postgresql.GetMatchListByLeagueID")
		return nil, DecodeDatabaseError(stderr.New(errors.ErrGetMatches))
	}
	defer rows.Close()

	for rows.Next() {
		var m entity.Match
		err = rows.Scan(&m.ID,
			&m.Date,
			&m.GameID,
			&m.Team1ID,
			&m.Team2ID,
			&m.Player1Team1ID,
			&m.Player2Team1ID,
			&m.Player1Team2ID,
			&m.Player2Team2ID,
			&m.ScoreTeam1,
			&m.ScoreTeam2,
			&m.Player1Team1RateBefore,
			&m.Player1Team2RateBefore,
			&m.Player2Team1RateBefore,
			&m.Player2Team2RateBefore,
			&m.Player1Team1RateAfter,
			&m.Player1Team2RateAfter,
			&m.Player2Team1RateAfter,
			&m.Player2Team2RateAfter)
		if err != nil {
			logger.Error().Err(err).Msg("failed to postgresql.GetMatchListByGameID")
			return nil, DecodeDatabaseError(stderr.New(errors.ErrGetMatches))
		}

		matches = append(matches, m)
	}

	return matches, nil
}

func (db *RWDBOperation) RewriteMatchesAndPlayerRatings(logger zerolog.Logger, ctx context.Context, matches []entity.Match, ratings map[int64]entity.Rating) error {
	tx, err := db.db.Begin(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("failed to postgresql.UpdateGame")
		return DecodeDatabaseError(stderr.New(errors.ErrUpdateGame))
	}

	for _, match := range matches {

		if match.Player2Team1ID != nil && *match.Player2Team1ID == 0 {
			match.Player2Team1ID = nil
		}
		if match.Player2Team2ID != nil && *match.Player2Team2ID == 0 {
			match.Player2Team2ID = nil
		}

		res, err := tx.Exec(ctx, queryUpdateMatchWithRatings,
			match.ID,
			match.Date,
			match.GameID,
			match.Team1ID,
			match.Team2ID,
			match.Player1Team1ID,
			match.Player2Team1ID,
			match.Player1Team2ID,
			match.Player2Team2ID,
			match.ScoreTeam1,
			match.ScoreTeam2,
			match.Player1Team1RateBefore,
			match.Player1Team2RateBefore,
			match.Player2Team1RateBefore,
			match.Player2Team2RateBefore,
			match.Player1Team1RateAfter,
			match.Player1Team2RateAfter,
			match.Player2Team1RateAfter,
			match.Player2Team2RateAfter,)
		if err != nil {
			_ = tx.Rollback(ctx)
			logger.Error().Err(err).Msg("failed to update match in postgresql.RewriteMatchesAndPlayerRatings")
			return DecodeDatabaseError(stderr.New(errors.ErrUpdateMatch))
		}
		if res.RowsAffected() == 0 {
			_ = tx.Rollback(ctx)
			logger.Error().Err(err).Msg("failed to update match in postgresql.RewriteMatchesAndPlayerRatings")
			return DecodeDatabaseError(stderr.New(errors.ErrMatchNotFound))
		}
	}

	for _, rating := range ratings {
		res, err := tx.Exec(ctx, queryCreateRatingInsertIgnore,
			rating.PlayerID,
			rating.LeagueID,
			rating.Value,
		)
		if err != nil {
			_ = tx.Rollback(ctx)
			logger.Error().Err(err).Msg("failed to update match in postgresql.RewriteMatchesAndPlayerRatings")
			return DecodeDatabaseError(stderr.New(errors.ErrUpdateMatch))
		}
		if res.RowsAffected() == 0 {
			_ = tx.Rollback(ctx)
			logger.Error().Err(err).Msg("failed to update match in postgresql.RewriteMatchesAndPlayerRatings")
			return DecodeDatabaseError(stderr.New(errors.ErrMatchNotFound))
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		_ = tx.Rollback(ctx)
		logger.Error().Err(err).Msg("failed to postgresql.RewriteMatchesAndPlayerRatings")
		return DecodeDatabaseError(stderr.New(errors.ErrUpdateGame))
	}

	return nil
}
