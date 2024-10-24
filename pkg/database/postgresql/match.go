package postgresql

import (
	"context"
	stdrerr "errors"
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
	_, err := db.db.Exec(ctx, queryUpdateMatch,
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
		logger.Error().Err(err).Msg("failed to update Match record")
		return DecodeDatabaseError(err)
	}

	return nil
}

func (db *RWDBOperation) DeleteMatch(logger zerolog.Logger, ctx context.Context, id int64) error {
	_, err := db.db.Exec(ctx, queryDeleteMatch, id)
	if err != nil {
		logger.Error().Err(err).Msg("failed to delete Match record")
		return DecodeDatabaseError(err)
	}

	return nil
}

func (db *RDBOperation) GetMatchesByPlayerID(logger zerolog.Logger, ctx context.Context, playerID int) ([]entities.PlayersMatch, error) {
	query := `
		SELECT id, date, game_id, team1_id, team2_id, player1_team1_id, player2_team1_id, player1_team2_id, player2_team2_id, score_team1, score_team2
		FROM public.matches
		WHERE player1_team1_id = $1
		   OR player2_team1_id = $1
		   OR player1_team2_id = $1
		   OR player2_team2_id = $1`

	matches := make([]entities.PlayersMatch, 0)

	rows, err := db.db.Query(ctx, query, playerID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to postgresql.GetMatchesByPlayerID")
		return nil, DecodeDatabaseError(stdrerr.New(errors.ErrGetMatches))
	}
	defer rows.Close()

	for rows.Next() {
		var m entities.PlayersMatch
		err = rows.Scan(&m.ID, &m.Date, &m.GameID, &m.Team1ID, &m.Team2ID, &m.Player1Team1ID, &m.Player2Team1ID,
			&m.Player1Team2ID, &m.Player2Team2ID, &m.ScoreTeam1, &m.ScoreTeam2)
		if err != nil {
			logger.Error().Err(err).Msg("failed to postgresql.GetMatchesByPlayerID")
			return nil, DecodeDatabaseError(stdrerr.New(errors.ErrGetMatches))
		}

		matches = append(matches, m)
	}

	return matches, nil
}
