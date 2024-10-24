package postgresql

import (
	"context"
	stderr "errors"
	"github.com/rs/zerolog"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/errors"
)

func (db *RWDBOperation) CreatePlayer(logger zerolog.Logger, ctx context.Context, p entities.CreatePlayerRequest) (int, error) {
	const query string = `
		INSERT INTO public.players(name, second_name, last_name, active_player, avatar, city_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id;`

	var id int

	err := db.db.QueryRow(ctx, query, p.Name, p.SecondName, p.LastName, p.ActivePlayer, p.Avatar, p.CityID).Scan(&id)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to postgresql.CreatePlayer")
		return 0, DecodeDatabaseError(stderr.New(errors.ErrCreatePlayer))
	}

	return id, nil
}

func (db *RWDBOperation) DeletePlayer(logger zerolog.Logger, ctx context.Context, id int) error {
	const query1 string = `
		UPDATE public.players
		SET deleted_at = NOW()
		WHERE id = $1;`

	const query2 string = `
		DELETE FROM public.players_teams_links
		WHERE player_id = $1;`

	tx, err := db.db.Begin(ctx)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to postgresql.DeletePlayer")
		return DecodeDatabaseError(stderr.New(errors.ErrDeletePlayer))
	}

	tag, err := tx.Exec(ctx, query1, id)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to postgresql.DeletePlayer")
		_ = tx.Rollback(ctx)
		return DecodeDatabaseError(stderr.New(errors.ErrDeletePlayer))
	}
	if tag.RowsAffected() == 0 {
		err = stderr.New(errors.ErrPlayerDontDeleted)
		logger.Error().Stack().Err(err).Msg(err.Error())
		return DecodeDatabaseError(err)
	}

	_, err := tx.Exec(ctx, query2, id)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to postgresql.DeletePlayer")
		_ = tx.Rollback(ctx)
		return DecodeDatabaseError(stderr.New(errors.ErrDeletePlayer))
	}

	if err = tx.Commit(ctx); err != nil {
		logger.Error().Stack().Err(err).Msg("failed to postgresql.DeletePlayer")
		_ = tx.Rollback(ctx)
		return DecodeDatabaseError(stderr.New(errors.ErrDeletePlayer))
	}

	return nil
}

func (db *RWDBOperation) UpdatePlayer(logger zerolog.Logger, ctx context.Context, p entities.UpdatePlayerRequest) error {
	query := `
		UPDATE public.players p
		SET
		    name = COALESCE($1, p.name),
		    second_name = COALESCE($2, p.second_name),
		    last_name = COALESCE($3, p.last_name),
		    active_player = COALESCE($4, p.active_player),
		    avatar = COALESCE($5, p.avatar),
		    city_id = COALESCE($6, p.city_id)
		WHERE id = $7;
		    `

	tag, err := db.db.Exec(ctx, query, p.Name, p.SecondName, p.LastName, p.ActivePlayer, p.Avatar, p.CityID, p.ID)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to postgresql.UpdatePlayer")
		return DecodeDatabaseError(stderr.New(errors.ErrUpdatePlayer))
	}
	if tag.RowsAffected() == 0 {
		err = stderr.New(errors.ErrPlayerDontUpdated)
		logger.Error().Stack().Err(err).Msg(err.Error())
		return DecodeDatabaseError(err)
	}

	return nil
}

func (db *RDBOperation) GetPlayerByID(logger zerolog.Logger, ctx context.Context, playerID int) (entities.Player, error) {
	const query string = `
	SELECT
        p.id,
        p.name,
        p.second_name,
        p.last_name,
        p.active_player,
        p.deleted_at,
        p.avatar,
        p.city_id,
        t.id,
        t.name,
        t.short_name
    FROM players p
    LEFT JOIN public.players_teams_links ptl ON p.id = ptl.player_id
    LEFT JOIN public.teams t ON t.id = ptl.team_id
    WHERE p.id = $1;
	`

	var p entities.Player

	err := db.db.QueryRow(ctx, query, playerID).
		Scan(&p.ID, &p.Name, &p.SecondName, &p.LastName, &p.ActivePlayer, &p.DeletedAt, &p.Avatar, &p.CityID, &p.TeamID, &p.TeamName, &p.TeamShortName)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to postgresql.GetPlayerByID")
		return entities.Player{}, DecodeDatabaseError(stderr.New(errors.ErrGetPlayer))
	}

	return p, nil
}

func (db *RDBOperation) GetPlayersByTeamID(logger zerolog.Logger, ctx context.Context, teamID int) ([]entities.Player, error) {
	const query string = `
		SELECT p.id, p.name, p.second_name, p.last_name, p.active_player, p.deleted_at, p.avatar, p.city_id, t.id, t.name, t.short_name
		FROM players p
		LEFT JOIN public.players_teams_links ptl ON p.id = ptl.player_id
		LEFT JOIN public.teams t ON t.id = ptl.team_id
		WHERE t.id = $1 AND p.deleted_at IS NULL;
	`

	var players []entities.Player

	rows, err := db.db.Query(ctx, query, teamID)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to postgresql.GetPlayersByTeamID")
		return nil, DecodeDatabaseError(stderr.New(errors.ErrGetPlayerList))
	}
	defer rows.Close()

	for rows.Next() {
		var p entities.Player

		err = rows.Scan(&p.ID, &p.Name, &p.SecondName, &p.LastName, &p.ActivePlayer, &p.DeletedAt, &p.Avatar, &p.CityID, &p.TeamID, &p.TeamName, &p.TeamShortName)
		if err != nil {
			logger.Error().Stack().Err(err).Msg("failed to postgresql.GetPlayersByTeamID")
			return nil, DecodeDatabaseError(stderr.New(errors.ErrGetPlayer))
		}

		players = append(players, p)
	}

	return players, nil
}

func (db *RDBOperation) FindPlayers(logger zerolog.Logger, ctx context.Context, player entities.FindPlayersRequest) ([]entities.Player, error) {
	const query string = `
		SELECT
			DISTINCT(p.id),
			p.name,
			p.second_name,
			p.last_name,
			p.avatar,
			p.active_player,
			p.deleted_at,
			p.city_id,
			t.id AS team_id,
			t.name AS team_name,
			t.short_name AS team_short_name
		FROM players p
				 LEFT JOIN public.players_teams_links ptl ON p.id = ptl.player_id
				 LEFT JOIN public.teams t ON t.id = ptl.team_id
				 LEFT JOIN public.rating r ON r.player_id = p.id
		WHERE 
		    ($1::int IS NULL OR t.league_id = $1) -- LeagueID
		  	AND ($2::varchar IS NULL OR (p.name ILIKE $2 OR p.second_name ILIKE $2 OR p.last_name ILIKE $2)) -- FindAny
		  	AND ($3::int IS NULL OR (
    			SELECT COUNT(*) 
    			FROM games g 
    			WHERE g.id IN (
					SELECT m.game_id 
					FROM matches m 
					WHERE m.player1_team1_id = p.id 
					   OR m.player2_team1_id = p.id 
					   OR m.player1_team2_id = p.id 
					   OR m.player2_team2_id = p.id)) = $3) -- GamesPlayedNumber
		  	AND ($4::int IS NULL OR r.value >= $4) -- Rating
		  	AND ($5::int IS NULL OR p.city_id = $5) -- CityID
		  	AND(
        		($6::bool IS NULL AND p.deleted_at IS NULL) OR -- Если $6 NULL, выбираем только тех у кого deleted_at NULL
        		($6::bool = TRUE) OR -- Если $6 TRUE, выбираем всех
        		($6::bool = FALSE AND p.deleted_at IS NULL) -- Если $6 FALSE, выбираем только тех у кого deleted_at NULL
    		)
			AND ($7::bool IS NULL OR NOT EXISTS (SELECT * FROM public.players_teams_links WHERE ptl.player_id = p.id) = $7); -- OnlyFree
	`

	var players []entities.Player

	rows, err := db.db.Query(ctx, query, player.LeagueID, player.FindAny, player.GamesPlayedNumber, player.Rating, player.CityID, player.WithDeleted, player.OnlyFree)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to postgresql.FindPlayers")
		return nil, DecodeDatabaseError(stderr.New(errors.ErrGetPlayerList))
	}
	defer rows.Close()

	for rows.Next() {
		var p entities.Player

		err = rows.Scan(&p.ID, &p.Name, &p.SecondName, &p.LastName, &p.Avatar, &p.ActivePlayer, &p.DeletedAt, &p.CityID, &p.TeamID, &p.TeamName, &p.TeamShortName)
		if err != nil {
			logger.Error().Stack().Err(err).Msg("failed to postgresql.FindPlayers")
			return nil, DecodeDatabaseError(stderr.New(errors.ErrGetPlayer))
		}

		players = append(players, p)
	}

	return players, nil
}
