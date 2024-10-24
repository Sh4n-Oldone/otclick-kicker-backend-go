package postgresql

import (
	"context"
	stderr "errors"
	"github.com/rs/zerolog"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/errors"
)

func (db *RDBOperation) GetGamesByPlayersTeam(logger zerolog.Logger, ctx context.Context, teamID int) ([]entities.Game, error) {
	const query string = `
		SELECT id, city_id, date, team1_id, team2_id
		FROM public.games
		WHERE team1_id = $1 OR team2_id = $1
	`

	var games []entities.Game

	rows, err := db.db.Query(ctx, query, teamID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to postgresql.GetGamesByPlayersTeam")
		return nil, DecodeDatabaseError(stderr.New(errors.ErrGetGameList))
	}
	defer rows.Close()

	for rows.Next() {
		var game entities.Game

		err = rows.Scan(&game.ID, &game.CityID, &game.Date, &game.Team1ID, &game.Team2ID)
		if err != nil {
			logger.Error().Err(err).Msg("failed to postgresql.GetGamesByPlayersTeam")
			return nil, DecodeDatabaseError(stderr.New(errors.ErrGetGame))
		}

		games = append(games, game)
	}

	return games, nil
}
