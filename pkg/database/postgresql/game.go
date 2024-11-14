package postgresql

import (
	"context"
	stderr "errors"
	"github.com/rs/zerolog"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/entity"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/errors"
	"time"
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

func (db *RWDBOperation) CreatePlayedGame(logger zerolog.Logger, ctx context.Context, request entities.CreateGameRequest) (entities.CreateGameResponse, error) {
	var gameId int
	matchIds := make([]int, 0)

	tx, err := db.db.Begin(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("failed to postgresql.CreatePlayedGame")
		return entities.CreateGameResponse{}, stderr.New(errors.ErrCreateGame)
	}

	err = tx.QueryRow(ctx, queryInsertGame, request.CityID, request.PlaceID, request.LeagueID, request.Date, request.Team1ID, request.Team2ID, request.TechLooseTeamID).
		Scan(&gameId)

	if err != nil {
		_ = tx.Rollback(ctx)
		logger.Error().Err(err).Msg("failed to postgresql.CreatePlayedGame")
		return entities.CreateGameResponse{}, stderr.New(errors.ErrCreateGame)
	}

	for _, match := range request.Matches {
		var matchId int

		if match.Player2Team1Id != nil && *match.Player2Team1Id == 0 {
			match.Player2Team1Id = nil
		}
		if match.Player2Team2Id != nil && *match.Player2Team2Id == 0 {
			match.Player2Team2Id = nil
		}

		err = tx.QueryRow(ctx, queryInsertMatchesToPlayedGame,
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
			match.Player2Team2RateAfter).
			Scan(&matchId)
		if err != nil {
			_ = tx.Rollback(ctx)
			logger.Error().Err(err).Msg("failed to postgresql.CreatePlayedGame")
			return entities.CreateGameResponse{}, stderr.New(errors.ErrCreateMatch)
		}
		matchIds = append(matchIds, matchId)
	}

	err = tx.Commit(ctx)
	if err != nil {
		_ = tx.Rollback(ctx)
		logger.Error().Err(err).Msg("failed to postgresql.CreatePlayedGame")
		return entities.CreateGameResponse{}, stderr.New(errors.ErrCreateGame)
	}

	return entities.CreateGameResponse{
		GameID:   gameId,
		MatchIDs: matchIds,
	}, nil
}

func (db *RWDBOperation) DeleteGame(logger zerolog.Logger, ctx context.Context, gameID int) error {
	const query1 string = `
		DELETE FROM public.matches
		WHERE game_id = $1
	`

	const query2 string = `
		DELETE FROM public.games
		WHERE id = $1
	`

	tx, err := db.db.Begin(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("failed to postgresql.DeleteGame")
		return DecodeDatabaseError(stderr.New(errors.ErrDeleteGame))
	}

	_, err = tx.Exec(ctx, query1, gameID)
	if err != nil {
		_ = tx.Rollback(ctx)
		logger.Error().Err(err).Msg("failed to postgresql.DeleteGame")
		return DecodeDatabaseError(stderr.New(errors.ErrDeleteMatch))
	}

	tag, err := tx.Exec(ctx, query2, gameID)
	if err != nil {
		_ = tx.Rollback(ctx)
		logger.Error().Err(err).Msg("failed to postgresql.DeleteGame")
		return DecodeDatabaseError(stderr.New(errors.ErrDeleteGame))
	}
	if tag.RowsAffected() == 0 {
		_ = tx.Rollback(ctx)
		err = stderr.New(errors.ErrGameNotFound)
		logger.Error().Err(err).Msg("failed to postgresql.DeleteGame")
		return DecodeDatabaseError(err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		_ = tx.Rollback(ctx)
		logger.Error().Err(err).Msg("failed to postgresql.DeleteGame")
		return DecodeDatabaseError(stderr.New(errors.ErrDeleteGame))
	}

	return nil
}

func (db *RDBOperation) GetGame(logger zerolog.Logger, ctx context.Context, gameID int) (entities.GetGameResponse, error) {
	const query string = `
	SELECT 
		g.id AS game_id,
		g.city_id,
		g.date AS game_date,
		g.place_id AS game_place_id,
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
	ORDER BY m.updated_at 
	`

	var game entities.GetGameResponse
	matches := make([]entities.FullMatch, 0)

	rows, err := db.db.Query(ctx, query, gameID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to postgresql.GetGame")
		return entities.GetGameResponse{}, DecodeDatabaseError(stderr.New(errors.ErrGetGame))
	}
	defer rows.Close()

	for rows.Next() {
		var match entities.FullMatch

		var p1t1Name, p2t1Name, p1t2Name, p2t2Name, p1t1LastName, p2t1LastName, p1t2LastName, p2t2LastName *string

		err = rows.Scan(
			&game.ID,
			&game.CityID,
			&game.Date,
			&game.PlaceID,
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
			return entities.GetGameResponse{}, DecodeDatabaseError(stderr.New(errors.ErrGetGame))
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
		PlaceID:         game.PlaceID,
		LeagueID:        game.LeagueID,
		Team1ID:         game.Team1ID,
		Team1Name:       game.Team1Name,
		Team2ID:         game.Team2ID,
		Team2Name:       game.Team2Name,
		TechLooseTeamID: game.TechLooseTeamID,
		Matches:         matches,
	}, nil
}

func (db *RWDBOperation) UpdateGame(logger zerolog.Logger, ctx context.Context, game entities.UpdateGameRequest, rates []entity.Rating) error {
	tx, err := db.db.Begin(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("failed to postgresql.UpdateGame")
		return stderr.New(errors.ErrUpdateGame)
	}

	tag, err := tx.Exec(ctx, queryUpdateGame, game.ID, game.Date, game.PlaceID, game.LeagueID, game.Team1ID, game.Team2ID, game.TechLooseTeamID)

	if err != nil {
		_ = tx.Rollback(ctx)
		logger.Error().Err(err).Msg("failed to postgresql.UpdateGame")
		return stderr.New(errors.ErrUpdateGame)
	}
	if tag.RowsAffected() == 0 {
		_ = tx.Rollback(ctx)
		err = stderr.New(errors.ErrGameNotFound)
		logger.Error().Err(err).Msg("failed to postgresql.UpdateGame")
		return err
	}

	ids := make([]int, 0)
	for _, match := range game.Matches {
		if match.ID != nil {
			ids = append(ids, *match.ID)
		}
	}

	_, err = tx.Exec(ctx, queryDeleteMatchesToUpdateGame, ids, game.ID)
	if err != nil {
		_ = tx.Rollback(ctx)
		err2 := stderr.New(errors.ErrDeleteMatch)
		logger.Error().Err(err).Msg("failed to delete match in postgresql.UpdateGame")
		return err2
	}

	for _, match := range game.Matches {

		if match.Player2Team1Id != nil && *match.Player2Team1Id == 0 {
			match.Player2Team1Id = nil
		}
		if match.Player2Team2Id != nil && *match.Player2Team2Id == 0 {
			match.Player2Team2Id = nil
		}

		//если id матча нет создаем новый
		if match.ID == nil {
			_, err = tx.Exec(ctx, queryInsertMatchesToUpdateGame,
				match.Date,
				game.ID,
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
				match.Player2Team2RateAfter)
			if err != nil {
				_ = tx.Rollback(ctx)
				logger.Error().Err(err).Msg("failed to create match in postgresql.UpdateGame")
				return stderr.New(errors.ErrUpdateMatch)
			}
			// если id матча есть то обновляем
		} else {
			tag2, err := tx.Exec(ctx, queryUpdateMatchesToUpdateGame,
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
				game.ID)
			if err != nil {
				_ = tx.Rollback(ctx)
				logger.Error().Err(err).Msg("failed to update match in postgresql.UpdateGame")
				return stderr.New(errors.ErrUpdateMatch)
			}
			if tag2.RowsAffected() == 0 {
				_ = tx.Rollback(ctx)
				logger.Error().Err(err).Msg("failed to update match in postgresql.UpdateGame")
				return stderr.New(errors.ErrMatchNotFound)
			}
		}
	}

	for _, rate := range rates {
		_, err := tx.Exec(ctx, queryCreateRatingInsertIgnore, rate.PlayerID, rate.LeagueID, rate.Value)
		if err != nil {
			_ = tx.Rollback(ctx)
			logger.Error().Err(err).Msg("failed to update match in postgresql.UpdateGame")
			return stderr.New(errors.ErrUpdateMatch)
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		_ = tx.Rollback(ctx)
		logger.Error().Err(err).Msg("failed to postgresql.UpdateGame")
		return stderr.New(errors.ErrUpdateGame)
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
		return nil, DecodeDatabaseError(stderr.New(errors.ErrGetGame))
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
			return nil, DecodeDatabaseError(stderr.New(errors.ErrGetGame))
		}

		games = append(games, game)
	}

	return games, nil
}

func (db *RWDBOperation) UpdateFutureGame(logger zerolog.Logger, ctx context.Context, request entity.UpdateFutureGameRequest) error {
	const query string = `
		UPDATE public.games
		SET 
		    league_id = $2,
			date = $3,
			place_id = $4,
			team1_id = $5,
			team2_id = $6,
			updated_at = NOW()
		WHERE id = $1;
	`

	tag, err := db.db.Exec(ctx, query, request.ID, request.LeagueID, request.Date, request.PlaceID, request.Team1ID, request.Team2ID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to postgresql.UpdateFutureGame")
		return stderr.New(errors.ErrUpdateGame)
	}
	if tag.RowsAffected() == 0 {
		err = stderr.New(errors.ErrGameNotFound)
		logger.Error().Err(err).Msg("failed to postgresql.UpdateFutureGame")
		return err
	}

	return nil
}

func (db *RDBOperation) GetGamesYears(logger zerolog.Logger, ctx context.Context) (entity.GetGamesYearsResponse, error) {
	rows, err := db.db.Query(ctx, queryGetGamesYears)
	if err != nil {
		logger.Error().Err(err).Msg("failed to get years list")
		return entity.GetGamesYearsResponse{}, DecodeDatabaseError(err)
	}
	defer rows.Close()

	var years []entity.Year

	for rows.Next() {
		var year entity.Year

		err = rows.Scan(&year.Year)
		if err != nil {
			logger.Error().Err(err).Msg("failed to scan year record")
			return entity.GetGamesYearsResponse{}, DecodeDatabaseError(err)
		}

		years = append(years, year)
	}

	return entity.GetGamesYearsResponse{Years: years}, nil
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
		return nil, DecodeDatabaseError(stderr.New(errors.ErrGetGame))
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
			return nil, DecodeDatabaseError(stderr.New(errors.ErrGetGame))
		}

		games = append(games, game)
	}

	return games, nil
}

func (db *RDBOperation) GetFutureGames(logger zerolog.Logger, ctx context.Context, cityID int) ([]entity.ShortGame, error) {
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
		LEFT JOIN places p ON g.place_id = p.id
		LEFT JOIN bars b ON p.bar_id = b.id
		LEFT JOIN tables tbl ON p.table_id = tbl.id
		JOIN teams t1 ON g.team1_id = t1.id
		JOIN teams t2 ON g.team2_id = t2.id
		JOIN teams_leagues_links tll1 ON t1.id = tll1.team_id  --связи с лигой команды 1
		JOIN teams_leagues_links tll2 ON t2.id = tll2.team_id  --связи с лигой команды 2
		JOIN leagues l ON tll1.league_id = l.id AND tll2.league_id = l.id  -- объединение по одинаковой лиге
		WHERE g.date >= CURRENT_DATE AND g.city_id = $1
		ORDER BY g.date;
	`

	rows, err := db.db.Query(ctx, query, cityID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to postgresql.GetFutureGames")
		return nil, DecodeDatabaseError(stderr.New(errors.ErrGetGameList))
	}
	defer rows.Close()

	games := make([]entity.ShortGame, 0)

	for rows.Next() {
		var g entity.ShortGame
		var t1 entity.TeamShort
		var t2 entity.TeamShort

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

		g.Teams = []entity.TeamShort{t1, t2}

		games = append(games, g)
	}

	return games, nil
}

func (db *RWDBOperation) CreateFutureGame(logger zerolog.Logger, ctx context.Context, request entity.CreateFutureGameRequest) (int, error) {
	const query string = `
		INSERT INTO games (city_id, league_id, date, place_id, team1_id, team2_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id;
	`

	var id int

	err := db.db.QueryRow(ctx, query, request.CityID, request.LeagueID, request.Date, request.PlaceID, request.Team1ID, request.Team2ID).Scan(&id)
	if err != nil {
		logger.Error().Err(err).Msg("failed to postgresql.CreateFutureGame")
		return 0, stderr.New(errors.ErrCreateGame)
	}

	return id, nil
}

func (db *RDBOperation) GetTeamGames(logger zerolog.Logger, ctx context.Context, teamID int) ([]entity.TeamGame, error) {
	const query string = `
		WITH
			league_teams AS (
				SELECT
					t.id as team_id,
					t.name as team_name,
					t.short_name as team_short_name
				FROM teams t
				JOIN teams_leagues_links tll ON t.id = tll.team_id
				WHERE tll.league_id = (SELECT league_id FROM teams_leagues_links WHERE team_id = $1)
				  AND t.id != $1
			),
			games_as_team1 AS (
				SELECT
					g.id,
					lt.team_id,
					lt.team_name,
					lt.team_short_name,
					p.id as place_id,
					b.id as bar_id,
					b.name as bar_name,
					tbl.id as table_id,
					tbl.name as table_name,
					g.date,
					true as is_home_game
				FROM league_teams lt
						 LEFT JOIN games g ON g.team1_id = $1 AND g.team2_id = lt.team_id
						 LEFT JOIN places p ON g.place_id = p.id
						 LEFT JOIN bars b ON p.bar_id = b.id
						 LEFT JOIN tables tbl ON p.table_id = tbl.id
			),
			games_as_team2 AS (
				SELECT
					g.id,
					lt.team_id,
					lt.team_name,
					lt.team_short_name,
					p.id as place_id,
					b.id as bar_id,
					b.name as bar_name,
					tbl.id as table_id,
					tbl.name as table_name,
					g.date,
					false as is_home_game
				FROM league_teams lt
						 LEFT JOIN games g ON g.team2_id = $1 AND g.team1_id = lt.team_id
						 LEFT JOIN places p ON g.place_id = p.id
						 LEFT JOIN bars b ON p.bar_id = b.id
						 LEFT JOIN tables tbl ON p.table_id = tbl.id
			)
		SELECT *
		FROM (
				 SELECT * FROM games_as_team1
				 UNION ALL
				 SELECT * FROM games_as_team2
			 ) all_games
		WHERE date >= CURRENT_DATE or date IS NULL
		ORDER BY team_id;
	`

	games := make([]entity.TeamGame, 0)

	rows, err := db.db.Query(ctx, query, teamID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to postgresql.GetTeamGames")
		return nil, stderr.New(errors.ErrGetGameList)
	}

	for rows.Next() {
		var game entity.TeamGame

		err = rows.Scan(&game.ID, &game.Team.ID, &game.Team.Name, &game.Team.ShortName, &game.Place.ID,
			&game.Place.Bar.ID, &game.Place.Bar.Name, &game.Place.Table.ID, &game.Place.Table.Name, &game.Date, &game.IsHomeGame)
		if err != nil {
			logger.Error().Err(err).Msg("failed to postgresql.GetTeamGames")
			return nil, stderr.New(errors.ErrGetGameList)
		}

		games = append(games, game)
	}

	return games, nil
}

func (db *RWDBOperation) CreateGameWithRating(
	logger zerolog.Logger, ctx context.Context, request entities.CreateGameRequest, rates map[int]int, operator *string, leagueID int64) (entities.CreateGameResponse, error) {

	var gameId int
	matchIds := make([]int, 0)

	tx, err := db.db.Begin(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("failed to postgresql.CreateGameWithRating")
		return entities.CreateGameResponse{}, stderr.New(errors.ErrCreateGame)
	}

	err = tx.QueryRow(ctx, queryInsertGame, request.CityID, request.PlaceID, request.LeagueID, request.Date, request.Team1ID, request.Team2ID, request.TechLooseTeamID).
		Scan(&gameId)

	if err != nil {
		_ = tx.Rollback(ctx)
		logger.Error().Err(err).Msg("failed to postgresql.CreateGameWithRating")
		return entities.CreateGameResponse{}, stderr.New(errors.ErrCreateGame)
	}

	for _, match := range request.Matches {
		var matchId int

		if match.Player2Team1Id != nil && *match.Player2Team1Id == 0 {
			match.Player2Team1Id = nil
		}
		if match.Player2Team2Id != nil && *match.Player2Team2Id == 0 {
			match.Player2Team2Id = nil
		}

		err = tx.QueryRow(ctx, queryInsertMatchesToPlayedGame,
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
			match.Player2Team2RateAfter).
			Scan(&matchId)
		if err != nil {
			_ = tx.Rollback(ctx)
			logger.Error().Err(err).Msg("failed to postgresql.CreateGameWithRating")
			return entities.CreateGameResponse{}, stderr.New(errors.ErrCreateMatch)
		}
		matchIds = append(matchIds, matchId)
	}

	for playerID, value := range rates {
		if playerID == 0 {
			continue
		}
		rate := &entity.Rating{
			PlayerID: int64(playerID),
			LeagueID: leagueID,
			Value:    int64(value),
		}

		query := queryCreateRating
		if operator != nil && *operator == "insertIgnore" {
			query = queryCreateRatingInsertIgnore
		}

		_, err = tx.Exec(ctx, query, rate.PlayerID, rate.LeagueID, rate.Value)
		if err != nil {
			_ = tx.Rollback(ctx)
			logger.Error().Err(err).Msg("failed to postgresql.CreateGameWithRating")
			return entities.CreateGameResponse{}, stderr.New(errors.ErrRating)
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		_ = tx.Rollback(ctx)
		logger.Error().Err(err).Msg("failed to postgresql.CreateGameWithRating")
		return entities.CreateGameResponse{}, stderr.New(errors.ErrCreateGame)
	}

	return entities.CreateGameResponse{
		GameID:   gameId,
		MatchIDs: matchIds,
	}, nil
}

func (db *RWDBOperation) UpdateGameWithRating(
	logger zerolog.Logger, ctx context.Context, game entities.UpdateGameRequest, rates map[int]int, operator *string, leagueID int64) error {

	tx, err := db.db.Begin(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("failed to postgresql.UpdateGame")
		return stderr.New(errors.ErrUpdateGame)
	}

	tag, err := tx.Exec(ctx, queryUpdateGame, game.ID, game.Date, game.PlaceID, game.LeagueID, game.Team1ID, game.Team2ID)
	if err != nil {
		_ = tx.Rollback(ctx)
		logger.Error().Err(err).Msg("failed to postgresql.UpdateGame")
		return stderr.New(errors.ErrUpdateGame)
	}
	if tag.RowsAffected() == 0 {
		_ = tx.Rollback(ctx)
		err = stderr.New(errors.ErrGameNotFound)
		logger.Error().Err(err).Msg("failed to postgresql.UpdateGame")
		return err
	}

	ids := make([]int, 0)
	for _, match := range game.Matches {
		if match.ID != nil {
			ids = append(ids, *match.ID)
		}
	}

	_, err = tx.Exec(ctx, queryDeleteMatchesToUpdateGame, ids, game.ID)
	if err != nil {
		_ = tx.Rollback(ctx)
		err2 := stderr.New(errors.ErrDeleteMatch)
		logger.Error().Err(err).Msg("failed to delete match in postgresql.UpdateGame")
		return err2
	}

	for _, match := range game.Matches {

		if match.Player2Team1Id != nil && *match.Player2Team1Id == 0 {
			match.Player2Team1Id = nil
		}
		if match.Player2Team2Id != nil && *match.Player2Team2Id == 0 {
			match.Player2Team2Id = nil
		}

		//если id матча нет создаем новый
		if match.ID == nil {
			_, err = tx.Exec(ctx, queryInsertMatchesToUpdateGame,
				match.Date,
				game.ID,
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
				match.Player2Team2RateAfter)
			if err != nil {
				_ = tx.Rollback(ctx)
				logger.Error().Err(err).Msg("failed to create match in postgresql.UpdateGame")
				return stderr.New(errors.ErrUpdateMatch)
			}
			// если id матча есть то обновляем
		} else {
			_, err = tx.Exec(ctx, queryUpdateMatchesToUpdateGame,
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
				game.ID)
			if err != nil {
				_ = tx.Rollback(ctx)
				logger.Error().Err(err).Msg("failed to update match in postgresql.UpdateGame")
				return stderr.New(errors.ErrUpdateMatch)
			}
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		_ = tx.Rollback(ctx)
		logger.Error().Err(err).Msg("failed to postgresql.UpdateGame")
		return stderr.New(errors.ErrUpdateGame)
	}

	return nil
}
