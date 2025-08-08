package postgresql

import (
	"context"
	"errors"

	"github.com/rs/zerolog"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/config"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/constant"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	pkgerr "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/errors"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/helpers/pointer"
)

func (db *RWDBOperation) CreatePlayer(logger zerolog.Logger, ctx context.Context, p entities.CreatePlayerRequest, cfg *config.DBConfig) (int, error) {
	timeout, cancel := context.WithTimeout(ctx, cfg.MaxIdleConnectionTimeout)
	defer cancel()

	const query string = `
		INSERT INTO public.players(name, second_name, last_name, active_player, avatar, city_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id;`

	var id int

	err := db.db.QueryRow(timeout, query, p.Name, p.SecondName, p.LastName, p.ActivePlayer, p.Avatar, p.CityID).Scan(&id)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to postgresql.CreatePlayer")
		return 0, DecodeDatabaseError(errors.New(pkgerr.ErrCreatePlayer))
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
		return DecodeDatabaseError(errors.New(pkgerr.ErrDeletePlayer))
	}

	tag, err := tx.Exec(ctx, query1, id)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to postgresql.DeletePlayer")
		_ = tx.Rollback(ctx)
		return DecodeDatabaseError(errors.New(pkgerr.ErrDeletePlayer))
	}
	if tag.RowsAffected() == 0 {
		err = errors.New(pkgerr.ErrPlayerNotFound)
		logger.Error().Stack().Err(err).Msg(err.Error())
		return DecodeDatabaseError(err)
	}

	_, err = tx.Exec(ctx, query2, id)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to postgresql.DeletePlayer")
		_ = tx.Rollback(ctx)
		return DecodeDatabaseError(errors.New(pkgerr.ErrDeletePlayer))
	}

	if err = tx.Commit(ctx); err != nil {
		logger.Error().Stack().Err(err).Msg("failed to postgresql.DeletePlayer")
		_ = tx.Rollback(ctx)
		return DecodeDatabaseError(errors.New(pkgerr.ErrDeletePlayer))
	}

	return nil
}

func (db *RWDBOperation) RecoverPlayer(logger zerolog.Logger, ctx context.Context, id int) error {
	const query string = `
		UPDATE public.players
		SET deleted_at = NULL
		WHERE id = $1;`

	tag, err := db.db.Exec(ctx, query, id)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to postgresql.RecoverPlayer")
		return DecodeDatabaseError(errors.New(pkgerr.ErrRecoverPlayer))
	}
	if tag.RowsAffected() == 0 {
		err = errors.New(pkgerr.ErrPlayerNotFound)
		logger.Error().Stack().Err(err).Msg(err.Error())
		return DecodeDatabaseError(err)
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
		return DecodeDatabaseError(errors.New(pkgerr.ErrUpdatePlayer))
	}
	if tag.RowsAffected() == 0 {
		err = errors.New(pkgerr.ErrPlayerNotFound)
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
        c.ru,
        t.id,
        t.name,
        t.short_name
    FROM players p
    LEFT JOIN public.players_teams_links ptl ON p.id = ptl.player_id
    LEFT JOIN public.teams t ON t.id = ptl.team_id
    LEFT JOIN public.cities c ON p.city_id = c.id
    WHERE p.id = $1;
	`

	var p entities.Player

	err := db.db.QueryRow(ctx, query, playerID).
		Scan(&p.ID, &p.Name, &p.SecondName, &p.LastName, &p.ActivePlayer, &p.DeletedAt, &p.Avatar, &p.CityID, &p.CityName, &p.TeamID, &p.TeamName, &p.TeamShortName)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to postgresql.GetPlayerByID")
		return entities.Player{}, DecodeDatabaseError(errors.New(pkgerr.ErrGetPlayer))
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
		return nil, DecodeDatabaseError(errors.New(pkgerr.ErrGetPlayerList))
	}
	defer rows.Close()

	for rows.Next() {
		var p entities.Player

		err = rows.Scan(&p.ID, &p.Name, &p.SecondName, &p.LastName, &p.ActivePlayer, &p.DeletedAt, &p.Avatar, &p.CityID, &p.TeamID, &p.TeamName, &p.TeamShortName)
		if err != nil {
			logger.Error().Stack().Err(err).Msg("failed to postgresql.GetPlayersByTeamID")
			return nil, DecodeDatabaseError(errors.New(pkgerr.ErrGetPlayer))
		}

		players = append(players, p)
	}

	return players, nil
}

func (db *RDBOperation) FindPlayers(logger zerolog.Logger, ctx context.Context, player entities.FindPlayersRequest) ([]entities.Player, error) {
	const query string = `
		SELECT DISTINCT
			p.id,
			p.name,
			p.second_name,
			p.last_name,
			p.avatar,
			p.active_player,
			p.deleted_at,
			p.city_id,
			c.ru,
			t.id AS team_id,
			t.name AS team_name,
			t.short_name AS team_short_name,
			r.value AS rating
		FROM players p
		LEFT JOIN players_teams_links ptl ON p.id = ptl.player_id
		LEFT JOIN teams t ON t.id = ptl.team_id
		LEFT JOIN teams_leagues_links tll ON t.id = tll.team_id
		LEFT JOIN rating r ON r.player_id = p.id AND r.league_id = tll.league_id
		LEFT JOIN cities c ON c.id = p.city_id
		WHERE 
		    ($1::int IS NULL OR tll.league_id = $1)
			AND ($2::varchar IS NULL OR 
				(LOWER(COALESCE(p.name, '')) LIKE LOWER('%' || $2 || '%') OR 
				 LOWER(COALESCE(p.second_name, '')) LIKE LOWER('%' || $2 || '%') OR 
				 LOWER(COALESCE(p.last_name, '')) LIKE LOWER('%' || $2 || '%'))
			)
			AND ($3::int IS NULL OR (
				SELECT COUNT(DISTINCT g.id)
				FROM games g
				JOIN matches m ON m.game_id = g.id
				WHERE m.player1_team1_id = p.id 
				   OR m.player2_team1_id = p.id 
				   OR m.player1_team2_id = p.id 
				   OR m.player2_team2_id = p.id
			) >= $3)
			AND ($4::int IS NULL OR r.value >= $4)
			AND ($5::int IS NULL OR p.city_id = $5)
			AND (
			    CASE
					WHEN $6::bool IS NULL OR $6::bool = FALSE THEN (p.deleted_at IS NULL) = true
					ELSE true--((p.deleted_at IS NULL) = true) or ((p.deleted_at IS NULL) = false)
				END
			)
			AND (
			    CASE
                when $7::bool is true then
                    not exists (select ptl2.team_id
                                from players_teams_links ptl2
                                where ptl2.player_id = p.id)
                else true
            end
			);
	`

	var players []entities.Player

	rows, err := db.db.Query(ctx, query, player.LeagueID, player.FindAny, player.GamesPlayedNumber, player.Rating, player.CityID, player.WithDeleted, player.OnlyFree)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to postgresql.FindPlayers")
		return nil, DecodeDatabaseError(errors.New(pkgerr.ErrGetPlayerList))
	}
	defer rows.Close()

	for rows.Next() {
		var p entities.Player

		err = rows.Scan(&p.ID, &p.Name, &p.SecondName, &p.LastName, &p.Avatar, &p.ActivePlayer, &p.DeletedAt, &p.CityID, &p.CityName, &p.TeamID, &p.TeamName, &p.TeamShortName, &p.Rating)
		if err != nil {
			logger.Error().Stack().Err(err).Msg("failed to postgresql.FindPlayers")
			return nil, DecodeDatabaseError(errors.New(pkgerr.ErrGetPlayer))
		}
		if p.Rating == nil {
			p.Rating = pointer.GetPointer(constant.DefaultRating)
		}

		players = append(players, p)
	}

	return players, nil
}

func (db *RDBOperation) GetPlayerIDsByLeagueID(logger zerolog.Logger, ctx context.Context, leagueID int64) ([]int64, error) {
	const query string = `
		SELECT p.id
		FROM players p
		LEFT JOIN public.players_teams_links ptl ON p.id = ptl.player_id
		LEFT JOIN public.teams_leagues_links t ON t.team_id = ptl.team_id
		LEFT JOIN public.leagues l ON l.id = t.league_id
		WHERE l.id = $1 AND p.deleted_at IS NULL;`

	var iDs []int64

	rows, err := db.db.Query(ctx, query, leagueID)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to postgresql.GetPlayerIDsByLeagueID")
		return nil, DecodeDatabaseError(errors.New(pkgerr.ErrGetPlayerList))
	}
	defer rows.Close()

	for rows.Next() {
		var id int64

		err = rows.Scan(&id)
		if err != nil {
			logger.Error().Stack().Err(err).Msg("failed to postgresql.GetPlayersByTeamID")
			return nil, DecodeDatabaseError(errors.New(pkgerr.ErrGetPlayer))
		}

		iDs = append(iDs, id)
	}

	return iDs, nil
}
