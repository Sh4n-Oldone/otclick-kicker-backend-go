package postgresql

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/config"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/database/postgresql/tx"
	pkgerr "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/errors"
)

func (db *RDBOperation) GetPastGamesByPlayersTeam(logger zerolog.Logger, ctx context.Context, teamID int) ([]entities.GameShort, error) {
	const query string = `
		SELECT id, city_id, date, team1_id, team2_id
		FROM public.games
		WHERE (team1_id = $1 OR team2_id = $1) AND date < NOW()
	`

	var games []entities.GameShort

	rows, err := db.db.Query(ctx, query, teamID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to postgresql.GetPastGamesByPlayersTeam")
		return nil, DecodeDatabaseError(errors.New(pkgerr.ErrGetGameList))
	}
	defer rows.Close()

	for rows.Next() {
		var game entities.GameShort

		err = rows.Scan(&game.ID, &game.CityID, &game.Date, &game.Team1ID, &game.Team2ID)
		if err != nil {
			logger.Error().Err(err).Msg("failed to postgresql.GetPastGamesByPlayersTeam")
			return nil, DecodeDatabaseError(errors.New(pkgerr.ErrGetGame))
		}

		games = append(games, game)
	}

	return games, nil
}

func (db *RDBOperation) GetPastGamesByTeamAndLeague(logger zerolog.Logger, ctx context.Context, teamId, leagueId int) ([]entities.GameShort, error) {
	const query string = `
		SELECT id
		FROM public.games
		WHERE (team1_id = $1 OR team2_id = $1) AND league_id = $2 AND date < NOW()
	`

	var games []entities.GameShort

	rows, err := db.db.Query(ctx, query, teamId, leagueId)
	if err != nil {
		logger.Error().Err(err).Msg("failed to postgresql.GetPastGamesByTeamAndLeague")
		return nil, DecodeDatabaseError(errors.New(pkgerr.ErrGetGameList))
	}
	defer rows.Close()

	for rows.Next() {
		var game entities.GameShort

		err = rows.Scan(&game.ID)
		if err != nil {
			logger.Error().Err(err).Msg("failed to postgresql.GetPastGamesByTeamAndLeague")
			return nil, DecodeDatabaseError(errors.New(pkgerr.ErrGetGame))
		}

		games = append(games, game)
	}

	return games, nil
}

func (db *RDBOperation) GetPastGamesByPlayersTeams(logger zerolog.Logger, ctx context.Context, teamIDs []int) ([]entities.GameShort, error) {
	const query string = `
		SELECT 
			pg.id,
			pg.league_id, 
			pg.city_id, 
			pg.date,
			pg.team1_id, 
			pg.team2_id,
			ts.tournament_id
		FROM public.games AS pg
			LEFT JOIN public.tournament_stages ts ON pg.stage_id = ts.id
		WHERE ((pg.team1_id = ANY ($1::int[])) OR (pg.team2_id = ANY ($1::int[]))) AND pg.date < NOW();
	`

	var games []entities.GameShort

	rows, err := db.db.Query(ctx, query, teamIDs)
	if err != nil {
		logger.Error().Err(err).Msg("failed to postgresql.GetPastGamesByPlayersTeams")
		return nil, DecodeDatabaseError(errors.New(pkgerr.ErrGetGameList))
	}
	defer rows.Close()

	for rows.Next() {
		var game entities.GameShort

		err = rows.Scan(&game.ID, &game.LeagueID, &game.CityID, &game.Date, &game.Team1ID, &game.Team2ID, &game.TournamentID)
		if err != nil {
			logger.Error().Err(err).Msg("failed to postgresql.GetPastGamesByPlayersTeams")
			return nil, DecodeDatabaseError(errors.New(pkgerr.ErrGetGame))
		}

		games = append(games, game)
	}

	return games, nil
}

func (db *RWDBOperation) DeleteGame(logger zerolog.Logger, ctx context.Context, gameID int64, cfg *config.DBConfig) error {
	timeout, cancel := context.WithTimeout(ctx, cfg.MaxIdleConnectionTimeout)
	defer cancel()

	const query1 string = `
		DELETE FROM public.matches
		WHERE game_id = $1
	`

	const query2 string = `
		DELETE FROM public.games
		WHERE id = $1
	`

	tx, err := db.db.Begin(timeout)
	if err != nil {
		logger.Error().Err(err).Msg("failed to postgresql.DeleteGame")
		return DecodeDatabaseError(errors.New(pkgerr.ErrDeleteGame))
	}

	_, err = tx.Exec(timeout, query1, gameID)
	if err != nil {
		_ = tx.Rollback(ctx)
		logger.Error().Err(err).Msg("failed to postgresql.DeleteGame")
		return DecodeDatabaseError(errors.New(pkgerr.ErrDeleteMatch))
	}

	tag, err := tx.Exec(timeout, query2, gameID)
	if err != nil {
		_ = tx.Rollback(ctx)
		logger.Error().Err(err).Msg("failed to postgresql.DeleteGame")
		return DecodeDatabaseError(errors.New(pkgerr.ErrDeleteGame))
	}
	if tag.RowsAffected() == 0 {
		_ = tx.Rollback(ctx)
		err = errors.New(pkgerr.ErrGameNotFound)
		logger.Error().Err(err).Msg("failed to postgresql.DeleteGame")
		return DecodeDatabaseError(err)
	}

	err = tx.Commit(timeout)
	if err != nil {
		_ = tx.Rollback(ctx)
		logger.Error().Err(err).Msg("failed to postgresql.DeleteGame")
		return DecodeDatabaseError(errors.New(pkgerr.ErrDeleteGame))
	}

	return nil
}

func (db *RDBOperation) GetGame(logger zerolog.Logger, ctx context.Context, gameID int) (entities.GetGameResponse, error) {
	timeout, cancel := context.WithTimeout(ctx, db.cfg.MaxIdleConnectionTimeout)
	defer cancel()

	const query string = `
	SELECT 
		g.id AS game_id,
		g.city_id,
		g.date AS game_date,
		g.place_id AS game_place_id,
		b.id AS game_bar_id,
		b.name AS game_bar_name,
		tbl.id AS game_table_id,
		tbl.name AS game_table_name,
		g.league_id,
		t1.id AS team1_id,
		t1.name AS team1_name,
		t2.id AS team2_id,
		t2.name AS team2_name,
		g.tech_loose_team_id,
		m.id AS match_id,
		m.date AS match_date,
		m.team1_id AS match_team1_id,
		m.team2_id AS match_team2_id,
		m.player1_team1_id AS player1_team1_id,
		pt1_1.name AS player1_team1_name,
		pt1_1.last_name AS player1_team1_last_name,
		m.player2_team1_id AS player2_team1_id,
		pt1_2.name AS player2_team1_name,
		pt1_2.last_name AS player2_team1_last_name,
		m.player1_team2_id AS player1_team2_id,
		pt2_1.name AS player1_team2_name,
		pt2_1.last_name AS player1_team2_last_name,
		m.player2_team2_id AS player2_team2_id,
		pt2_2.name AS player2_team2_name,
		pt2_2.last_name AS player2_team2_last_name,
		m.score_team1,
		m.score_team2
	FROM 
		public.games g
	LEFT JOIN 
		places p ON g.place_id = p.id
	LEFT JOIN 
		bars b ON p.bar_id = b.id
	LEFT JOIN 
		tables tbl ON p.table_id = tbl.id
	LEFT JOIN 
		public.teams t1 ON g.team1_id = t1.id
	LEFT JOIN 
		public.teams t2 ON g.team2_id = t2.id
	LEFT JOIN 
		public.matches m ON g.id = m.game_id
	LEFT JOIN 
		public.players pt1_1 ON m.player1_team1_id = pt1_1.id
	LEFT JOIN 
		public.players pt1_2 ON m.player2_team1_id = pt1_2.id
	LEFT JOIN 
		public.players pt2_1 ON m.player1_team2_id = pt2_1.id
	LEFT JOIN 
		public.players pt2_2 ON m.player2_team2_id = pt2_2.id
	WHERE 
		g.id = $1
	ORDER BY m.sort, m.updated_at
	`

	var game entities.GetGameResponse
	matches := make([]entities.FullMatch, 0)

	rows, err := db.db.Query(timeout, query, gameID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to postgresql.GetGame")
		return entities.GetGameResponse{}, DecodeDatabaseError(errors.New(pkgerr.ErrGetGame))
	}
	defer rows.Close()

	for rows.Next() {
		var match entities.FullMatch

		var p1t1Name, p2t1Name, p1t2Name, p2t2Name, p1t1LastName, p2t1LastName, p1t2LastName, p2t2LastName *string

		err = rows.Scan(
			&game.ID,
			&game.CityID,
			&game.Date,
			&game.Place.ID,
			&game.Place.Bar.ID,
			&game.Place.Bar.Name,
			&game.Place.Table.ID,
			&game.Place.Table.Name,
			&game.LeagueID,
			&game.Team1ID,
			&game.Team1Name,
			&game.Team2ID,
			&game.Team2Name,
			&game.TechLooseTeamID,
			&match.ID,
			&match.Date,
			&match.Team1ID,
			&match.Team2ID,
			&match.Player1Team1Id,
			&p1t1Name,
			&p1t1LastName,
			&match.Player2Team1Id,
			&p2t1Name,
			&p2t1LastName,
			&match.Player1Team2Id,
			&p1t2Name,
			&p1t2LastName,
			&match.Player2Team2Id,
			&p2t2Name,
			&p2t2LastName,
			&match.ScoreTeam1,
			&match.ScoreTeam2,
		)
		if err != nil {
			logger.Error().Err(err).Msg("failed to postgresql.GetGame")
			return entities.GetGameResponse{}, DecodeDatabaseError(errors.New(pkgerr.ErrGetGame))
		}

		if p1t1Name != nil {
			match.Player1Team1Name = p1t1Name
			*match.Player1Team1Name = *match.Player1Team1Name + " "
		}
		if p1t1LastName != nil {
			*match.Player1Team1Name = *match.Player1Team1Name + *p1t1LastName
		}

		if p2t1Name != nil {
			match.Player2Team1Name = p2t1Name
			*match.Player2Team1Name = *match.Player2Team1Name + " "
		}
		if p2t1LastName != nil {
			*match.Player2Team1Name = *match.Player2Team1Name + *p2t1LastName
		}

		if p1t2Name != nil {
			match.Player1Team2Name = p1t2Name
			*match.Player1Team2Name = *match.Player1Team2Name + " "
		}
		if p1t2LastName != nil {
			*match.Player1Team2Name = *match.Player1Team2Name + *p1t2LastName
		}

		if p2t2Name != nil {
			match.Player2Team2Name = p2t2Name
			*match.Player2Team2Name = *match.Player2Team2Name + " "
		}
		if p2t2LastName != nil {
			*match.Player2Team2Name = *match.Player2Team2Name + *p2t2LastName
		}

		if match.ID != nil {
			matches = append(matches, match)
		}
	}

	return entities.GetGameResponse{
		ID:              game.ID,
		CityID:          game.CityID,
		Date:            game.Date,
		Place:           game.Place,
		LeagueID:        game.LeagueID,
		Team1ID:         game.Team1ID,
		Team1Name:       game.Team1Name,
		Team2ID:         game.Team2ID,
		Team2Name:       game.Team2Name,
		TechLooseTeamID: game.TechLooseTeamID,
		Matches:         matches,
	}, nil
}

func (db *RDBOperation) GetGameById(logger zerolog.Logger, ctx context.Context, gameID int, tx tx.ITx) (entities.GameLeagueTournament, error) {
	timeout, cancel := context.WithTimeout(ctx, db.cfg.MaxIdleConnectionTimeout)
	defer cancel()

	const query string = `
		SELECT id, city_id, place_id, date, team1_id, team2_id, league_id, tech_loose_team_id, is_tiebreak, stage_id  
		FROM games
		WHERE id = $1;`

	var g entities.GameLeagueTournament

	err := poolOrTx(db.db, tx).QueryRow(timeout, query, gameID).
		Scan(&g.ID, &g.CityID, &g.PlaceID, &g.Date, &g.Team1ID, &g.Team2ID, &g.LeagueID, &g.TechLooseTeamID, &g.IsTiebreak, &g.StageID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to postgresql.GetGameById")
		return entities.GameLeagueTournament{}, DecodeDatabaseError(err)
	}

	return g, nil
}

func (db *RWDBOperation) UpdateGame(logger zerolog.Logger, ctx context.Context, game entities.UpdateGameRequest, tx tx.ITx) error {
	timeout, cancel := context.WithTimeout(ctx, db.cfg.MaxIdleConnectionTimeout)
	defer cancel()

	tag, err := poolOrTx(db.db, tx).Exec(timeout, queryUpdateGame, game.ID, game.Date, game.PlaceID, game.LeagueID, game.Team1ID, game.Team2ID, game.TechLooseTeamID)

	if err != nil {
		logger.Error().Err(err).Msg("failed to postgresql.UpdateGame")
		return DecodeDatabaseError(err)
	}
	if tag.RowsAffected() == 0 {
		err = pgx.ErrNoRows
		logger.Error().Err(err).Msg("failed to postgresql.UpdateGame")
		return DecodeDatabaseError(err)
	}

	return nil
}

func (db *RDBOperation) FindGames(logger zerolog.Logger, ctx context.Context, request entities.FindGameRequest) ([]entities.FindGame, error) {
	now := time.Now()

	const query string = `
	SELECT 
		g.id AS gameId, 
		g.date AS dateOfGame, 
		g.team1_id AS team1Id, 
		t1.short_name AS team1ShortName, 
		g.team2_id AS team2Id, 
		t2.short_name AS team2ShortName,
		COALESCE(SUM(m.score_team1), 0) AS totalScoreTeam1,
		COALESCE(SUM(m.score_team2), 0) AS totalScoreTeam2
	FROM public.games g
	JOIN teams t1 ON g.team1_id = t1.id
	JOIN teams t2 ON g.team2_id = t2.id
	LEFT JOIN matches m ON m.game_id = g.id
	WHERE ((g.team1_id = $1 AND g.team2_id = $2) OR (g.team1_id = $2 AND g.team2_id = $1))
	AND g.date <= $3
	GROUP BY g.id, t1.short_name, t2.short_name;
	`

	var games []entities.FindGame

	rows, err := db.db.Query(ctx, query, request.Team1ID, request.Team2ID, now)
	if err != nil {
		logger.Error().Err(err).Msg("failed to postgresql.FindGame")
		return nil, DecodeDatabaseError(errors.New(pkgerr.ErrGetGame))
	}

	for rows.Next() {
		var game entities.FindGame

		err = rows.Scan(
			&game.ID,
			&game.Date,
			&game.Team1ID,
			&game.Team1ShortName,
			&game.Team2ID,
			&game.Team2ShortName,
			&game.TotalScoreTeam1,
			&game.TotalScoreTeam2,
		)
		if err != nil {
			logger.Error().Err(err).Msg("failed to postgresql.FindGame")
			return nil, DecodeDatabaseError(errors.New(pkgerr.ErrGetGame))
		}

		games = append(games, game)
	}

	return games, nil
}

func (db *RWDBOperation) UpdateFutureGame(logger zerolog.Logger, ctx context.Context, req entities.UpdateFutureGameRequest, tx tx.ITx) error {
	timeout, cancel := context.WithTimeout(ctx, db.cfg.MaxIdleConnectionTimeout)
	defer cancel()

	const query string = `
		UPDATE public.games
		SET 
		    league_id = $2,
			date = $3,
			place_id = $4,
			team1_id = $5,
			team2_id = $6,
			updated_at = NOW()
		WHERE id = $1;`

	tag, err := poolOrTx(db.db, tx).Exec(timeout, query, req.ID, req.LeagueID, req.Date, req.PlaceID, req.Team1ID, req.Team2ID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to postgresql.UpdateFutureGame")
		return DecodeDatabaseError(err)
	}
	if tag.RowsAffected() == 0 {
		err = errors.New(pkgerr.ErrGameNotFound)
		logger.Error().Err(err).Msg("failed to postgresql.UpdateFutureGame")
		return DecodeDatabaseError(err)
	}

	return nil
}

func (db *RDBOperation) GetGamesYears(logger zerolog.Logger, ctx context.Context) (entities.GetGamesYearsResponse, error) {
	rows, err := db.db.Query(ctx, queryGetGamesYears)
	if err != nil {
		logger.Error().Err(err).Msg("failed to get years list")
		return entities.GetGamesYearsResponse{}, DecodeDatabaseError(err)
	}
	defer rows.Close()

	var years []entities.Year

	for rows.Next() {
		var year entities.Year

		err = rows.Scan(&year.Year)
		if err != nil {
			logger.Error().Err(err).Msg("failed to scan year record")
			return entities.GetGamesYearsResponse{}, DecodeDatabaseError(err)
		}

		years = append(years, year)
	}

	return entities.GetGamesYearsResponse{Years: years}, nil
}

func (db *RDBOperation) GetGameList(logger zerolog.Logger, ctx context.Context, r entities.GetGameListRequest) ([]entities.GameV2, error) {
	const query string = `
		SELECT 
			g.id,
			g.date,
			g.city_id,
			g.league_id,
			s.id,
			s.name,
			s.description,
			g.place_id,
			b.id,
			b.name,
			tbl.id,
			tbl.name,
			t1.id,
			t1.name,
			t1.short_name,
			t2.id,
			t2.name,
			t2.short_name,
			sum(score_team1) as score_team1,
			sum(score_team2) as score_team2,
			g.tech_loose_team_id,
			g.is_tiebreak
			FROM games g
				LEFT JOIN places p ON g.place_id = p.id
				LEFT JOIN teams t1 ON g.team1_id = t1.id
				LEFT JOIN teams t2 ON g.team2_id = t2.id
				LEFT JOIN bars b ON p.bar_id = b.id
				LEFT JOIN tables tbl ON p.table_id = tbl.id
				LEFT JOIN matches m ON g.id = m.game_id
				LEFT JOIN seasons s ON g.league_id = s.id
			WHERE
				($1::INTEGER IS NULL OR g.city_id = $1)
				AND ($2::INTEGER IS NULL OR g.league_id = $2)
				AND ($3::INTEGER IS NULL OR s.id = $3)
				AND ($4::TIMESTAMPTZ IS NULL OR g.date >= $4)
				AND ($5::TIMESTAMPTZ IS NULL OR g.date <= $5)
				AND ($6::INTEGER IS NULL OR g.place_id = $6)
				AND ($7::INTEGER IS NULL OR t1.id = $7)
				AND ($8::INTEGER IS NULL OR t2.id = $8)
				AND ($9::BOOLEAN IS NULL OR g.is_tiebreak = $9)
			GROUP BY g.id, p.id, t1.id, t2.id, b.id, tbl.id, s.id
			ORDER BY
				g.league_id,
				CASE WHEN $10 = 1 THEN g.id END DESC,
				g.id
			LIMIT $11
			OFFSET $12;
	`
	games := make([]entities.GameV2, 0)

	rows, err := db.db.Query(ctx, query,
		r.CityId, r.LeagueId, r.SeasonId, r.DateFrom, r.DateTo, r.PlaceId,
		r.Team1Id, r.Team2Id, r.IsTiebreak, r.SortType, r.Limit, r.Offset)
	if err != nil {
		logger.Error().Err(err).Msg("failed to postgresql.GetGameList")
		return nil, DecodeDatabaseError(err)
	}

	for rows.Next() {
		var game entities.GameV2
		game.Place = entities.PlaceShort{}
		game.Place.Bar = entities.BarShort{}
		game.Place.Table = entities.TableShort{}
		game.Season = entities.Season{}
		var seasonId *int64
		var seasonName *string
		var seasonDescription *string

		err = rows.Scan(
			&game.Id,
			&game.Date,
			&game.CityId,
			&game.LeagueId,
			&seasonId,
			&seasonName,
			&seasonDescription,
			&game.Place.ID,
			&game.Place.Bar.ID,
			&game.Place.Bar.Name,
			&game.Place.Table.ID,
			&game.Place.Table.Name,
			&game.Team1.ID,
			&game.Team1.Name,
			&game.Team1.ShortName,
			&game.Team2.ID,
			&game.Team2.Name,
			&game.Team2.ShortName,
			&game.ScoreTeam1,
			&game.ScoreTeam2,
			&game.TechLooseTeamId,
			&game.IsTiebreak,
		)
		if err != nil {
			logger.Error().Err(err).Msg("failed to postgresql.GetGameList")
			return nil, DecodeDatabaseError(err)
		}
		if seasonId != nil {
			game.Season.ID = *seasonId
		}
		if seasonName != nil {
			game.Season.Name = *seasonName
		}
		if seasonDescription != nil {
			game.Season.Description = *seasonDescription
		}

		games = append(games, game)
	}

	return games, nil
}

func (db *RDBOperation) GetComingGames(logger zerolog.Logger, ctx context.Context) ([]entities.ComingGame, error) {
	const query string = `
		SELECT 
			g.id,
			g.date,
			g.city_id,
			b.name,
			tbl.name,
			t1.id,
			t1.name,
			t2.id,
			t2.name
		FROM games g
			JOIN  places p ON g.place_id = p.id
			LEFT JOIN teams t1 ON g.team1_id = t1.id
			LEFT JOIN teams t2 ON g.team2_id = t2.id
			LEFT JOIN bars b ON p.bar_id = b.id
			LEFT JOIN tables tbl ON p.table_id = tbl.id
		WHERE 
			g.date > CURRENT_DATE
			AND p.deleted_at IS NULL;
	`

	games := make([]entities.ComingGame, 0)

	rows, err := db.db.Query(ctx, query)
	if err != nil {
		logger.Error().Err(err).Msg("failed to postgresql.GetComingGames")
		return nil, DecodeDatabaseError(errors.New(pkgerr.ErrGetGame))
	}

	for rows.Next() {
		var game entities.ComingGame

		err = rows.Scan(
			&game.ID,
			&game.Date,
			&game.CityID,
			&game.Bar,
			&game.Table,
			&game.Team1ID,
			&game.Team1Name,
			&game.Team2ID,
			&game.Team2Name,
		)
		if err != nil {
			logger.Error().Err(err).Msg("failed to postgresql.GetComingGames")
			return nil, DecodeDatabaseError(errors.New(pkgerr.ErrGetGame))
		}

		games = append(games, game)
	}

	return games, nil
}

func (db *RDBOperation) GetFutureGames(logger zerolog.Logger, ctx context.Context, cityID int) ([]entities.ShortGame, error) {
	const query string = `
		SELECT g.id, 
		       g.date,
		       g.city_id,
			   p.id,
			   b.id,
			   b.name,
			   tbl.id,
			   tbl.name,
		       l.id,
		       l.name,
		       g.team1_id, 
		       t1.short_name, 
		       t1.name, 
		       g.team2_id, 
		       t2.short_name, 
		       t2.name
		FROM games g
		LEFT JOIN leagues l ON g.league_id = l.id
		LEFT JOIN places p ON g.place_id = p.id
		LEFT JOIN bars b ON p.bar_id = b.id
		LEFT JOIN tables tbl ON p.table_id = tbl.id
		JOIN teams t1 ON g.team1_id = t1.id
		JOIN teams t2 ON g.team2_id = t2.id
		WHERE g.date > NOW() AND g.city_id = $1
		ORDER BY g.date;`

	rows, err := db.db.Query(ctx, query, cityID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to postgresql.GetFutureGames")
		return nil, DecodeDatabaseError(errors.New(pkgerr.ErrGetGameList))
	}
	defer rows.Close()

	games := make([]entities.ShortGame, 0)

	for rows.Next() {
		var g entities.ShortGame
		var t1 entities.TeamShort
		var t2 entities.TeamShort

		err = rows.Scan(
			&g.ID,
			&g.Date,
			&g.CityID,
			&g.Place.ID,
			&g.Place.Bar.ID,
			&g.Place.Bar.Name,
			&g.Place.Table.ID,
			&g.Place.Table.Name,
			&g.LeagueID,
			&g.LeagueName,
			&t1.ID,
			&t1.ShortName,
			&t1.Name,
			&t2.ID,
			&t2.ShortName,
			&t2.Name,
		)

		g.Teams = []entities.TeamShort{t1, t2}

		games = append(games, g)
	}

	return games, nil
}

func (db *RWDBOperation) CreateFutureGame(logger zerolog.Logger, ctx context.Context, req entities.CreateFutureGameRequest, tx tx.ITx) (int, error) {
	timeout, cancel := context.WithTimeout(ctx, db.cfg.MaxIdleConnectionTimeout)
	defer cancel()

	const query string = `
		INSERT INTO games (city_id, league_id, date, place_id, team1_id, team2_id, is_tiebreak)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id;`

	var id int

	err := poolOrTx(db.db, tx).QueryRow(
		timeout, query, req.CityID, req.LeagueID, req.Date, req.PlaceID, req.Team1ID, req.Team2ID, req.IsTiebreak,
	).Scan(&id)
	if err != nil {
		logger.Error().Err(err).Msg("failed to postgresql.CreateFutureGame")
		return 0, DecodeDatabaseError(err)
	}

	return id, nil
}

func (db *RDBOperation) GetTeamGames(logger zerolog.Logger, ctx context.Context, teamID int) ([]entities.TeamGame, error) {
	timeout, cancel := context.WithTimeout(ctx, db.cfg.MaxIdleConnectionTimeout)
	defer cancel()

	const query = `
    SELECT
        g.id as game_id,
        t.id as team_id,
        t.name,
        t.short_name,
        p.id as place_id,
        b.id as bar_id,
        b.name as bar_name,
        tbl.id as table_id,
        tbl.name as table_name,
        g.date,
        (g.team1_id = $1) as is_home_game
    FROM games g
    JOIN teams t ON t.id = CASE
        WHEN g.team1_id = $1 THEN g.team2_id
        WHEN g.team2_id = $1 THEN g.team1_id
    END
    LEFT JOIN places p ON g.place_id = p.id
    LEFT JOIN bars b ON p.bar_id = b.id
    LEFT JOIN tables tbl ON p.table_id = tbl.id
    WHERE (g.team1_id = $1 OR g.team2_id = $1)
    ORDER BY t.id;`

	games := make([]entities.TeamGame, 0)

	rows, err := db.db.Query(timeout, query, teamID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to postgresql.GetTeamGames")
		return nil, errors.New(pkgerr.ErrGetGameList)
	}

	for rows.Next() {
		var game entities.TeamGame

		err = rows.Scan(&game.ID, &game.Team.ID, &game.Team.Name, &game.Team.ShortName, &game.Place.ID,
			&game.Place.Bar.ID, &game.Place.Bar.Name, &game.Place.Table.ID, &game.Place.Table.Name, &game.Date, &game.IsHomeGame)
		if err != nil {
			logger.Error().Err(err).Msg("failed to postgresql.GetTeamGames")
			return nil, errors.New(pkgerr.ErrGetGameList)
		}

		games = append(games, game)
	}

	return games, nil
}

func (db *RWDBOperation) DeleteFutureGame(logger zerolog.Logger, ctx context.Context, req entities.DeleteFutureGameRequest, tx tx.ITx) error {
	timeout, cancel := context.WithTimeout(ctx, db.cfg.MaxIdleConnectionTimeout)
	defer cancel()

	res, err := poolOrTx(db.db, tx).Exec(timeout, queryDeleteFutureGame, req.ID)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to postgresql.DeleteFutureGame")
		return DecodeDatabaseError(err)
	}

	if res.RowsAffected() == 0 {
		err = pgx.ErrNoRows
		logger.Error().Stack().Err(err).Msg("failed find to postgresql.DeleteFutureGame")
		return DecodeDatabaseError(err)
	}

	return nil
}

func (db *RDBOperation) GetTournamentGameList(logger zerolog.Logger, ctx context.Context, req *entities.GetTournamentGameList) ([]entities.TournamentGame, error) {
	timeout, cancel := context.WithTimeout(ctx, db.cfg.MaxIdleConnectionTimeout)
	defer cancel()

	const query string = `
		SELECT g.id, g.city_id, g.place_id, g.date, g.team1_id, g.team2_id, g.tech_loose_team_id, g.is_tiebreak, g.stage_id
		FROM games AS g
			JOIN tournament_stages AS ts ON g.stage_id = ts.id
			JOIN tournaments AS t ON ts.tournament_id = t.id
		WHERE 
			($1::INT IS NULL OR t.id = $1)
			AND ($2::INT IS NULL OR g.city_id = $2);`

	rows, err := db.db.Query(timeout, query, req.TournamentId, req.CityId)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to postgresql.GetTournamentGames")
		return nil, DecodeDatabaseError(err)
	}
	defer rows.Close()

	var games []entities.TournamentGame

	for rows.Next() {
		var g entities.TournamentGame
		err = rows.Scan(&g.ID, &g.CityID, &g.PlaceID, &g.Date, &g.Team1ID, &g.Team2ID, &g.TechLooseTeamID, &g.IsTiebreak, &g.StageID)
		if err != nil {
			logger.Error().Stack().Err(err).Msg("failed rows.Scan postgresql.GetTournamentGames")
			return nil, DecodeDatabaseError(err)
		}
		games = append(games, g)
	}

	return games, nil
}

func (db *RWDBOperation) CreateFutureTournamentStageGames(logger zerolog.Logger, ctx context.Context, stageID, cityID int64, team1IDs, team2IDs []int64, tx tx.ITx) error {
	timeout, cancel := context.WithTimeout(ctx, db.cfg.MaxIdleConnectionTimeout)
	defer cancel()

	const query string = `
        INSERT INTO games (city_id, team1_id, team2_id, stage_id)
        VALUES ($1, $2, $3, $4);`

	exec := poolOrTx(db.db, tx)

	for i := range team1IDs {
		_, err := exec.Exec(timeout, query, cityID, team1IDs[i], team2IDs[i], stageID)
		if err != nil {
			logger.Error().Err(err).Msg("failed tx.Exec postgresql.CreateTournamentStageGames")
			return DecodeDatabaseError(err)
		}
	}

	return nil
}

func (db *RWDBOperation) CreateFutureTournamentGame(logger zerolog.Logger, ctx context.Context, request *entities.CreateFutureTournamentGameRequest, cfg *config.DBConfig) (int64, error) {
	timeout, cancel := context.WithTimeout(ctx, cfg.MaxIdleConnectionTimeout)
	defer cancel()

	const query string = `
		INSERT INTO games(city_id, place_id, date, team1_id, team2_id, is_tiebreak, stage_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id;`

	var id int64

	err := db.db.QueryRow(
		timeout,
		query,
		request.CityID,
		request.PlaceID,
		request.Date,
		request.Team1ID,
		request.Team2ID,
		request.IsTiebreak,
		request.StageID,
	).Scan(&id)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed QueryRow postgresql.CreateTournamentGame")
		return 0, DecodeDatabaseError(err)
	}

	return id, nil
}

func (db *RWDBOperation) UpdateFutureTournamentGame(logger zerolog.Logger, ctx context.Context, req *entities.UpdateFutureTournamentGameRequest, cfg *config.DBConfig) error {
	timeout, cancel := context.WithTimeout(ctx, cfg.MaxIdleConnectionTimeout)
	defer cancel()

	const query string = `
		UPDATE games
		SET
			place_id = COALESCE($2, place_id),
			date = COALESCE($3, date),
			team1_id = COALESCE($4, team1_id),
			team2_id = COALESCE($5, team2_id)
		WHERE id = $1;`

	tag, err := db.db.Exec(
		timeout,
		query,
		req.GameID,
		req.PlaceID,
		req.Date,
		req.Team1ID,
		req.Team2ID,
	)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed UpdateTournamentGame")
		return DecodeDatabaseError(err)
	}
	if tag.RowsAffected() == 0 {
		err = pgx.ErrNoRows
		logger.Error().Stack().Err(err).Msg("no rows affected UpdateTournamentGame")
		return DecodeDatabaseError(err)
	}

	return nil
}

func (db *RDBOperation) GetTournamentGame(logger zerolog.Logger, ctx context.Context, gameID int64, cfg *config.DBConfig) (entities.TournamentGame, error) {
	timeout, cancel := context.WithTimeout(ctx, cfg.MaxIdleConnectionTimeout)
	defer cancel()

	const query string = `
		SELECT g.id, g.city_id, g.place_id, g.date, g.team1_id, g.team2_id, g.tech_loose_team_id, g.is_tiebreak, g.stage_id
		FROM games AS g
		WHERE g.id = $1;`

	var g entities.TournamentGame

	err := db.db.QueryRow(timeout, query, gameID).
		Scan(&g.ID, &g.CityID, &g.PlaceID, &g.Date, &g.Team1ID, &g.Team2ID, &g.TechLooseTeamID, &g.IsTiebreak, &g.StageID)
	if err != nil {
		logger.Error().Err(err).Msg("failed QueryRow postgresql.GetTournamentGame")
		return entities.TournamentGame{}, DecodeDatabaseError(err)
	}

	return g, nil
}

func (db *RWDBOperation) CreatePlayedTournamentGame(logger zerolog.Logger, ctx context.Context, req *entities.CreatePlayedTournamentGameRequest, tx tx.ITx) (int64, error) {
	timeout, cancel := context.WithTimeout(ctx, db.cfg.MaxIdleConnectionTimeout)
	defer cancel()

	const query string = `
		INSERT INTO games (city_id, place_id, date, team1_id, team2_id, tech_loose_team_id, is_tiebreak, stage_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id;`

	var id int64

	err := poolOrTx(db.db, tx).QueryRow(
		timeout, query, req.CityID, req.PlaceID, req.Date, req.Team1ID, req.Team2ID, req.TechLooseTeamID, req.IsTiebreak, req.StageID,
	).Scan(&id)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed CreatePlayedTournamentGame")
		return 0, DecodeDatabaseError(err)
	}

	return id, nil
}

func (db *RWDBOperation) UpdatePlayedTournamentGame(logger zerolog.Logger, ctx context.Context, req *entities.UpdatePlayedTournamentGameRequest, tx tx.ITx) error {
	timeout, cancel := context.WithTimeout(ctx, db.cfg.MaxIdleConnectionTimeout)
	defer cancel()

	const query string = `UPDATE public.games
		SET 
			date = $2,
			place_id = $3,
			stage_id = $4,
			team1_id = $5,
			team2_id = $6,
			tech_loose_team_id = $7,
			updated_at = NOW()
		WHERE id = $1;`

	tag, err := poolOrTx(db.db, tx).Exec(timeout, query, req.GameID, req.Date, req.PlaceID, req.StageID, req.Team1ID, req.Team2ID, req.TechLooseTeamID)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed UpdatePlayedTournamentGame")
		return DecodeDatabaseError(err)
	}

	if tag.RowsAffected() == 0 {
		err = pgx.ErrNoRows
		logger.Error().Stack().Err(err).Msg("no rows updated")
		return DecodeDatabaseError(err)
	}

	return nil
}

func (db *RDBOperation) FetchTournamentTeamGames(logger zerolog.Logger, ctx context.Context, tournamentId, teamId int64) ([]entities.TournamentGame, error) {
	timeout, cancel := context.WithTimeout(ctx, db.cfg.MaxIdleConnectionTimeout)
	defer cancel()

	const query string = `
		SELECT g.id, g.city_id, g.place_id, g.date, g.team1_id, g.team2_id, g.tech_loose_team_id, g.is_tiebreak, g.stage_id
		FROM games AS g
			JOIN tournament_stages AS ts ON g.stage_id = ts.id
			JOIN tournaments AS t ON ts.tournament_id = t.id
		WHERE 
			t.id = $1 AND
			(g.team1_id = $2 OR g.team2_id = $2);`

	rows, err := db.db.Query(timeout, query, tournamentId, teamId)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to postgresql.FetchTournamentTeamGames")
		return nil, DecodeDatabaseError(err)
	}
	defer rows.Close()

	var games []entities.TournamentGame

	for rows.Next() {
		var g entities.TournamentGame
		err = rows.Scan(&g.ID, &g.CityID, &g.PlaceID, &g.Date, &g.Team1ID, &g.Team2ID, &g.TechLooseTeamID, &g.IsTiebreak, &g.StageID)
		if err != nil {
			logger.Error().Stack().Err(err).Msg("failed rows.Scan postgresql.FetchTournamentTeamGames")
			return nil, DecodeDatabaseError(err)
		}
		games = append(games, g)
	}

	return games, nil
}

func (db *RWDBOperation) CreateGame(logger zerolog.Logger, ctx context.Context, req entities.CreateGameRequest, tx tx.ITx) (int64, error) {
	timeout, cancel := context.WithTimeout(ctx, db.cfg.MaxIdleConnectionTimeout)
	defer cancel()

	var gameId int64

	err := poolOrTx(db.db, tx).QueryRow(
		timeout, queryInsertGame, req.CityID, req.PlaceID, req.LeagueID, req.Date, req.Team1ID, req.Team2ID, req.TechLooseTeamID, req.IsTiebreak,
	).Scan(&gameId)

	if err != nil {
		logger.Error().Err(err).Msg("failed to postgresql.CreateGame")
		return 0, err
	}

	return gameId, nil
}
