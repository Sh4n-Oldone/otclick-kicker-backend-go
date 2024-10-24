package postgresql

import (
	"context"
	stderr "errors"
	"github.com/rs/zerolog"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/errors"
)

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
