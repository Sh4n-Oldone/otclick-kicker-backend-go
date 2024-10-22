package postgresql

import (
	"context"
	"github.com/rs/zerolog"

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
