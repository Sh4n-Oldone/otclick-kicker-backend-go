package postgresql

import (
	"context"
	"errors"

	"github.com/rs/zerolog"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/config"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/database/postgresql/tx"
)

func (db *RWDBOperation) CreateMatch(logger zerolog.Logger, ctx context.Context, match entities.Match) (int64, error) {
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

func (db *RWDBOperation) CreateGameMatch(logger zerolog.Logger, ctx context.Context, match entities.GamesMatch, gameId int64, cfg *config.DBConfig) (int64, error) {
	timeout, cancel := context.WithTimeout(ctx, cfg.MaxIdleConnectionTimeout)
	defer cancel()

	const query string = `
	INSERT INTO matches (
		date,
		game_id,
		team1_id,
		team2_id,
		player1_team1_id,
		player2_team1_id,
		player1_team2_id,
		player2_team2_id,
		score_team1,
		score_team2,
		player1_team1_rate_before,
		player1_team2_rate_before,
		player2_team1_rate_before,
		player2_team2_rate_before,
		player1_team1_rate_after,
		player1_team2_rate_after,
		player2_team1_rate_after,
		player2_team2_rate_after,
		sort
	)
	VALUES ($1,	$2,	$3,	$4,	$5,	$6,	$7,	$8,	$9,	$10,$11, $12, $13, $14, $15, $16, $17, $18, $19) RETURNING id;`

	var id int64

	err := db.db.QueryRow(
		timeout,
		query,
		match.Date,
		gameId,
		match.Team1ID,
		match.Team2ID,
		match.Player1Team1Id,
		match.Player2Team1Id,
		match.Player1Team2Id,
		match.Player2Team2Id,
		match.ScoreTeam1,
		match.ScoreTeam2,
		match.Player1Team1RateBefore,
		match.Player1Team2RateBefore,
		match.Player2Team1RateBefore,
		match.Player2Team2RateBefore,
		match.Player1Team1RateAfter,
		match.Player1Team2RateAfter,
		match.Player2Team1RateAfter,
		match.Player2Team2RateAfter,
		match.Sort,
	).Scan(&id)
	if err != nil {
		logger.Error().Err(err).Msg("failed to create Match")
		return 0, DecodeDatabaseError(err)
	}

	return id, nil
}

func (db *RWDBOperation) UpdateMatch(logger zerolog.Logger, ctx context.Context, match entities.Match) error {
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
		return errors.New("failed to update match, it does not exist")
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
		return false, errors.New("failed to delete match, it does not exist")
	}

	return true, nil
}

func (db *RDBOperation) GetPastMatchesByPlayerID(logger zerolog.Logger, ctx context.Context, playerID int) ([]entities.MatchV2, error) {
	query := `
		SELECT 
			pm.id,
			pm.date,
			pg.league_id,
			pm.game_id,
			ts.tournament_id,
			pm.team1_id,
			pm.team2_id,
			pm.player1_team1_id,
			pm.player2_team1_id,
			pm.player1_team2_id,
			pm.player2_team2_id,
			pm.score_team1,
			pm.score_team2
		FROM public.matches AS pm
		JOIN public.games AS pg ON pm.game_id = pg.id
		LEFT JOIN public.tournament_stages AS ts ON pg.stage_id = ts.id
		WHERE (pm.player1_team1_id = $1
		   OR pm.player2_team1_id = $1
		   OR pm.player1_team2_id = $1
		   OR pm.player2_team2_id = $1)
		  AND pm.date < NOW()
		ORDER BY pm.sort, pm.updated_at;`

	matches := make([]entities.MatchV2, 0)

	rows, err := db.db.Query(ctx, query, playerID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to postgresql.GetPastMatchesByPlayerID")
		return nil, DecodeDatabaseError(err)
	}
	defer rows.Close()

	for rows.Next() {
		var m entities.MatchV2
		err = rows.Scan(&m.ID, &m.Date, &m.LeagueID, &m.GameID, &m.TournamentID, &m.Team1ID, &m.Team2ID, &m.Player1Team1ID, &m.Player2Team1ID,
			&m.Player1Team2ID, &m.Player2Team2ID, &m.ScoreTeam1, &m.ScoreTeam2)
		if err != nil {
			logger.Error().Err(err).Msg("failed to postgresql.GetPastMatchesByPlayerID")
			return nil, DecodeDatabaseError(err)
		}

		matches = append(matches, m)
	}

	return matches, nil
}

func (db *RDBOperation) GetMatchListByGameID(ctx context.Context, logger zerolog.Logger, gameID int64, cfg *config.DBConfig) ([]entities.MatchV2, error) {
	matches := make([]entities.MatchV2, 0)

	rows, err := db.db.Query(ctx, queryGetMatchListByGameID, gameID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to postgresql.GetPastMatchesByPlayerID")
		return nil, DecodeDatabaseError(err)
	}
	defer rows.Close()

	for rows.Next() {
		var m entities.MatchV2
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
			return nil, DecodeDatabaseError(err)
		}

		matches = append(matches, m)
	}

	return matches, nil
}

func (db *RDBOperation) GetMatchListByLeagueID(logger zerolog.Logger, ctx context.Context, leagueID int64) ([]entities.Match, error) {
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
		ORDER BY g.date, g.updated_at, m.sort, m.updated_at;`

	matches := make([]entities.Match, 0)

	rows, err := db.db.Query(ctx, query, leagueID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to postgresql.GetMatchListByLeagueID")
		return nil, DecodeDatabaseError(err)
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
			return nil, DecodeDatabaseError(err)
		}

		matches = append(matches, m)
	}

	return matches, nil
}

func (db *RDBOperation) GetMatchListByTournamentId(logger zerolog.Logger, ctx context.Context, tournamentId int64) ([]entities.Match, error) {
	timeout, cancel := context.WithTimeout(ctx, db.cfg.MaxIdleConnectionTimeout)
	defer cancel()

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
		JOIN games as g ON m.game_id = g.id
		JOIN tournaments_teams_link as t1 ON g.team1_id = t1.team_id
		JOIN tournaments_teams_link as t2 ON g.team2_id = t2.team_id
		JOIN tournaments as tt ON t1.tournament_id = tt.id AND t2.tournament_id = tt.id
		WHERE tt.id = $1
		ORDER BY g.date, g.updated_at, m.sort, m.updated_at;`

	matches := make([]entities.Match, 0)

	rows, err := db.db.Query(timeout, query, tournamentId)
	if err != nil {
		logger.Error().Err(err).Msg("failed to postgresql.GetMatchListByTournamentId")
		return nil, DecodeDatabaseError(err)
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
			logger.Error().Err(err).Msg("failed to postgresql.GetMatchListByTournamentId")
			return nil, DecodeDatabaseError(err)
		}

		matches = append(matches, m)
	}

	return matches, nil
}

func (db *RWDBOperation) RewriteMatchesAndPlayerRatings(logger zerolog.Logger, ctx context.Context, matches []entities.Match, ratings map[int64]entities.Rating) error {
	tx, err := db.db.Begin(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("failed to postgresql.UpdateGame")
		return DecodeDatabaseError(err)
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
			match.Player2Team2RateAfter)
		if err != nil {
			_ = tx.Rollback(ctx)
			logger.Error().Err(err).Msg("failed to update match in postgresql.RewriteMatchesAndPlayerRatings")
			return DecodeDatabaseError(err)
		}
		if res.RowsAffected() == 0 {
			_ = tx.Rollback(ctx)
			err = errors.New("no rows affected")
			logger.Error().Err(err).Msg("failed to update match in postgresql.RewriteMatchesAndPlayerRatings")
			return DecodeDatabaseError(err)
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
			return DecodeDatabaseError(err)
		}
		if res.RowsAffected() == 0 {
			_ = tx.Rollback(ctx)
			err = errors.New("no rows affected")
			logger.Error().Err(err).Msg("failed to update match in postgresql.RewriteMatchesAndPlayerRatings")
			return DecodeDatabaseError(err)
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		_ = tx.Rollback(ctx)
		logger.Error().Err(err).Msg("failed to postgresql.RewriteMatchesAndPlayerRatings")
		return DecodeDatabaseError(err)
	}

	return nil
}

func (db *RWDBOperation) RewriteTournamentMatchesAndPlayerRatings(logger zerolog.Logger, ctx context.Context, matches []entities.Match, ratings map[int64]entities.TournamentRating, tx tx.ITx) error {
	timeout, cancel := context.WithTimeout(ctx, db.cfg.MaxIdleConnectionTimeout)
	defer cancel()

	exec := poolOrTx(db.db, tx)

	for _, match := range matches {

		if match.Player2Team1ID != nil && *match.Player2Team1ID == 0 {
			match.Player2Team1ID = nil
		}
		if match.Player2Team2ID != nil && *match.Player2Team2ID == 0 {
			match.Player2Team2ID = nil
		}

		res, err := exec.Exec(timeout, queryUpdateMatchWithRatings,
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
			match.Player2Team2RateAfter)

		if err != nil {
			logger.Error().Err(err).Msg("failed to update match in postgresql.RewriteTournamentMatchesAndPlayerRatings")
			return DecodeDatabaseError(err)
		}
		if res.RowsAffected() == 0 {
			err = errors.New("no rows affected")
			logger.Error().Err(err).Msg("failed to update match in postgresql.RewriteTournamentMatchesAndPlayerRatings")
			return DecodeDatabaseError(err)
		}
	}

	for _, rating := range ratings {
		res, err := exec.Exec(timeout, queryCreateTournamentRatingInsertIgnore,
			rating.PlayerID,
			rating.TournamentID,
			rating.Value,
		)
		if err != nil {
			logger.Error().Err(err).Msg("failed to update match in postgresql.RewriteTournamentMatchesAndPlayerRatings")
			return DecodeDatabaseError(err)
		}
		if res.RowsAffected() == 0 {
			err = errors.New("no rows affected")
			logger.Error().Err(err).Msg("failed to update match in postgresql.RewriteTournamentMatchesAndPlayerRatings")
			return DecodeDatabaseError(err)
		}
	}

	return nil
}

func (db *RWDBOperation) DeleteOldGameMatches(logger zerolog.Logger, ctx context.Context, gameId int64, newMatchesIds []int, cfg *config.DBConfig) error {
	timeout, cancel := context.WithTimeout(ctx, cfg.MaxIdleConnectionTimeout)
	defer cancel()

	_, err := db.db.Exec(timeout, queryDeleteMatchesToUpdateGame, newMatchesIds, gameId)
	if err != nil {
		logger.Error().Err(err).Msg("failed to postgresql.DeleteOldGameMatches")
		return DecodeDatabaseError(err)
	}

	return nil
}

func (db *RWDBOperation) CreateNewMatch(logger zerolog.Logger, ctx context.Context, gameId int64, match *entities.NewMatch, cfg *config.DBConfig) error {
	timeout, cancel := context.WithTimeout(ctx, cfg.MaxIdleConnectionTimeout)
	defer cancel()

	_, err := db.db.Exec(
		timeout,
		queryInsertMatchesToUpdateGame,
		match.Date,
		gameId,
		match.Team1ID,
		match.Team2ID,
		match.Player1Team1Id,
		match.Player2Team1Id,
		match.Player1Team2Id,
		match.Player2Team2Id,
		match.ScoreTeam1,
		match.ScoreTeam2,
		match.Player1Team1RateBefore,
		match.Player2Team1RateBefore,
		match.Player1Team2RateBefore,
		match.Player2Team2RateBefore,
		match.Player1Team1RateAfter,
		match.Player2Team1RateAfter,
		match.Player1Team2RateAfter,
		match.Player2Team2RateAfter,
		match.Sort,
	)
	if err != nil {
		logger.Error().Err(err).Msg("failed to postgresql.CreateNewMatch")
		return DecodeDatabaseError(err)
	}

	return nil
}

func (db *RWDBOperation) UpdateOldMatch(logger zerolog.Logger, ctx context.Context, gameId int64, match *entities.NewMatch, cfg *config.DBConfig) error {
	timeout, cancel := context.WithTimeout(ctx, cfg.MaxIdleConnectionTimeout)
	defer cancel()

	tag2, err := db.db.Exec(timeout, queryUpdateMatchesToUpdateGame,
		match.ID,
		match.Date,
		match.Team1ID,
		match.Team2ID,
		match.Player1Team1Id,
		match.Player2Team1Id,
		match.Player1Team2Id,
		match.Player2Team2Id,
		match.ScoreTeam1,
		match.ScoreTeam2,
		match.Player1Team1RateBefore,
		match.Player1Team2RateBefore,
		match.Player2Team1RateBefore,
		match.Player2Team2RateBefore,
		match.Player1Team1RateAfter,
		match.Player1Team2RateAfter,
		match.Player2Team1RateAfter,
		match.Player2Team2RateAfter,
		match.Sort,
		gameId)
	if err != nil {
		logger.Error().Err(err).Msg("failed to update match in postgresql.UpdateGame")
		return DecodeDatabaseError(err)
	}
	if tag2.RowsAffected() == 0 {
		err = errors.New("no rows affected")
		logger.Error().Err(err).Msg("failed to update match in postgresql.UpdateGame")
		return DecodeDatabaseError(err)
	}

	return nil
}

func (db *RWDBOperation) DeleteGameMatches(logger zerolog.Logger, ctx context.Context, gameId int64, cfg *config.DBConfig) error {
	timeout, cancel := context.WithTimeout(ctx, cfg.MaxIdleConnectionTimeout)
	defer cancel()

	const query string = "DELETE FROM matches WHERE game_id = $1"

	_, err := db.db.Exec(timeout, query, gameId)
	if err != nil {
		logger.Error().Err(err).Msg("failed to delete match in postgresql.DeleteMatches")
		return DecodeDatabaseError(err)
	}

	return nil
}

func (db *RWDBOperation) DeleteTournamentMatches(logger zerolog.Logger, ctx context.Context, tournamentId int64, tx tx.ITx) error {
	timeout, cancel := context.WithTimeout(ctx, db.cfg.MaxIdleConnectionTimeout)
	defer cancel()

	const query string = `
		DELETE FROM matches 
		WHERE game_id IN (
			SELECT g.id 
				FROM games g
				JOIN tournament_stages ts ON g.stage_id = ts.id
			WHERE ts.tournament_id = $1);`

	_, err := poolOrTx(db.db, tx).Exec(timeout, query, tournamentId)
	if err != nil {
		logger.Error().Err(err).Msg("failed Exec DeleteTournamentMatches")
		return DecodeDatabaseError(err)
	}

	return nil
}
