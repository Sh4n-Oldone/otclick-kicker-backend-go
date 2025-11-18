package postgresql

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/database/postgresql/tx"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/config"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/constant"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	errtmp "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/error_templates"
	pkgerr "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/errors"
)

func (db *RDBOperation) GetTeam(logger zerolog.Logger, ctx context.Context, teamID int64) (entities.GetTeamResponse, error) {
	timeout, cancel := context.WithTimeout(ctx, db.cfg.MaxIdleConnectionTimeout)
	defer cancel()

	team := entities.GetTeamResponse{}

	const queryGetTeam = `SELECT t.id, t.name, t.short_name, t.avatar, t.city_id,
	COALESCE(ARRAY_AGG(DISTINCT tll.league_id ORDER BY tll.league_id) FILTER (WHERE tll.league_id IS NOT NULL), ARRAY[]::int8[]),
	COALESCE(ARRAY_AGG(DISTINCT ptl.player_id ORDER BY ptl.player_id) FILTER (WHERE ptl.player_id IS NOT NULL), ARRAY[]::int8[])
	FROM teams t
	LEFT JOIN teams_leagues_links tll ON t.id = tll.team_id
	LEFT JOIN players_teams_links ptl ON t.id = ptl.team_id
	WHERE t.id = $1
	GROUP BY t.id`

	var leagueIDs []int64
	var playerIDs []int64
	err := db.db.QueryRow(timeout, queryGetTeam, teamID).Scan(&team.ID, &team.Name, &team.ShortName, &team.Avatar, &team.CityId, &leagueIDs, &playerIDs)
	if err != nil {
		logger.Error().Err(err).Msg("failed to GetTeam")
		return entities.GetTeamResponse{}, DecodeDatabaseError(err)
	}

	const queryGetLeagueByID = `SELECT name FROM leagues WHERE id = $1;`

	var leagues = []entities.LeagueShort{}

	for _, leagueID := range leagueIDs {
		var l entities.LeagueShort

		l.ID = leagueID

		err = db.db.QueryRow(timeout, queryGetLeagueByID, leagueID).
			Scan(&l.Name)
		if err != nil {
			logger.Error().Stack().Err(err).Msg("failed to postgresql.GetTeam/queryGetPlayerByID")
			return entities.GetTeamResponse{}, DecodeDatabaseError(errors.New(pkgerr.ErrGetPlayer))
		}

		leagues = append(leagues, l)
	}

	queryGetPlayerByID := `
		SELECT
			p.id,
			p.name,
			p.second_name,
			p.last_name,
			p.avatar,
			p.active_player,
			(p.deleted_at IS NOT NULL) AS deleted,
			p.city_id,
			t.name,
			t.short_name,
			r.value
		FROM players p
		LEFT JOIN players_teams_links ptl ON p.id = ptl.player_id
		LEFT JOIN teams t ON t.id = ptl.team_id
		LEFT JOIN rating r ON p.id = r.player_id %s
		WHERE p.id = $1;`

	// возможно, проще было сделать через слияние SQL запросов
	if len(leagueIDs) == 0 {
		queryGetPlayerByID = fmt.Sprintf(queryGetPlayerByID, "")
	} else {
		queryGetPlayerByID = fmt.Sprintf(queryGetPlayerByID, "AND r.league_id IN ("+strings.Trim(strings.ReplaceAll(fmt.Sprint(leagueIDs), " ", ", "), "[]")+")")
	}
	// альтернативные варианты перевода массива чисел в строку
	//strings.Trim(strings.Join(strings.Fields(fmt.Sprint(leagueIDs)), ", "), "[]")
	//strings.Trim(strings.Join(strings.Split(fmt.Sprint(leagueIDs), " "), ", "), "[]")
	var players = []entities.PlayerGetTeam{}
	for _, playerID := range playerIDs {
		var p entities.PlayerGetTeam

		err = db.db.QueryRow(timeout, queryGetPlayerByID, playerID).
			Scan(&p.ID,
				&p.Name,
				&p.SecondName,
				&p.LastName,
				&p.Avatar,
				&p.ActivePlayer,
				&p.Deleted,
				&p.CityID,
				&p.TeamName,
				&p.TeamShortName,
				&p.RatingNumber)
		if err != nil {
			logger.Error().Stack().Err(err).Msg("failed to postgresql.GetTeam/queryGetPlayerByID")
			return entities.GetTeamResponse{}, DecodeDatabaseError(errors.New(pkgerr.ErrGetPlayer))
		}
		players = append(players, p)
	}

	team.Leagues = leagues
	team.Players = players

	return team, nil
}

func (db *RDBOperation) GetTeams(logger zerolog.Logger, ctx context.Context, cityID int64, onlyFree bool) ([]entities.TeamShort, error) {
	query := `SELECT t.id, t.name, t.short_name FROM teams t
		WHERE t.city_id = $1`
	if onlyFree {
		query += " AND NOT EXISTS (SELECT 1 FROM teams_leagues_links tll WHERE tll.team_id = t.id)"
	}
	query += " ORDER BY id"

	rows, err := db.db.Query(ctx, query, cityID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to GetTeams")
		return nil, DecodeDatabaseError(err)
	}
	defer rows.Close()

	var teams []entities.TeamShort

	for rows.Next() {
		var team entities.TeamShort

		err = rows.Scan(&team.ID, &team.Name, &team.ShortName)
		if err != nil {
			logger.Error().Err(err).Msg("failed to scan team list")
			return nil, DecodeDatabaseError(err)
		}

		teams = append(teams, team)
	}

	return teams, nil
}

func (db *RDBOperation) GetTeamsByCity(logger zerolog.Logger, ctx context.Context, onlyFree bool, cityID int64) ([]entities.TeamShort, error) {
	const query = `SELECT t.id, t.name, t.short_name
		FROM teams t
		LEFT JOIN teams_leagues_links tll ON t.id = tll.team_id
		WHERE ($1::boolean IS NOT TRUE OR tll.league_id IS NULL)
		  AND t.city_id = $2
		ORDER BY t.id`

	rows, err := db.db.Query(ctx, query, onlyFree, cityID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to GetTeamsByCity")
		return nil, DecodeDatabaseError(err)
	}
	defer rows.Close()

	var teams []entities.TeamShort

	for rows.Next() {
		var team entities.TeamShort

		err = rows.Scan(&team.ID, &team.Name, &team.ShortName)
		if err != nil {
			logger.Error().Err(err).Msg("failed to scan team list")
			return nil, DecodeDatabaseError(err)
		}

		teams = append(teams, team)
	}

	return teams, nil
}

func (db *RDBOperation) GetTeamsByLeague(logger zerolog.Logger, ctx context.Context, leagueID int64) ([]entities.TeamByLeague, error) {
	const query = `SELECT t.id, t.name, t.short_name, t.avatar, t.city_id, COALESCE(ARRAY_AGG(ptl.player_id) FILTER (WHERE ptl.player_id IS NOT NULL), ARRAY[]::int8[])
	FROM teams t
	LEFT JOIN players_teams_links ptl ON ptl.team_id = t.id
	LEFT JOIN teams_leagues_links tll ON t.id = tll.team_id
	WHERE tll.league_id = $1
	GROUP BY t.id
	ORDER BY t.id`

	rows, err := db.db.Query(ctx, query, leagueID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to GetTeamsByLeague")
		return nil, DecodeDatabaseError(err)
	}
	defer rows.Close()

	var teams []entities.TeamByLeague

	for rows.Next() {
		var team entities.TeamByLeague

		err = rows.Scan(&team.Id, &team.Name, &team.ShortName, &team.Avatar, &team.CityId, &team.Players)
		if err != nil {
			logger.Error().Err(err).Msg("failed to scan team list")
			return nil, DecodeDatabaseError(err)
		}

		teams = append(teams, team)
	}

	return teams, nil
}

func (db *RDBOperation) FetchLeagues(logger zerolog.Logger, ctx context.Context, cityID int64, seasonID int64) ([]entities.League, error) {
	rows, err := db.db.Query(ctx, "SELECT id, name FROM leagues WHERE city_id = $1 AND season_id = $2", cityID, seasonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var leagues []entities.League
	for rows.Next() {
		var league entities.League
		err = rows.Scan(&league.ID, &league.Name)
		if err != nil {
			return nil, err
		}
		leagues = append(leagues, league)
	}

	return leagues, nil
}

func (db *RDBOperation) FetchTeams(logger zerolog.Logger, ctx context.Context, leagueID int64) ([]entities.Team, error) {
	const query = `SELECT t.id, t.short_name FROM teams t 
		LEFT JOIN teams_leagues_links tll ON t.id = tll.team_id
		WHERE tll.league_id = $1`

	rows, err := db.db.Query(ctx, query, leagueID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []entities.Team
	for rows.Next() {
		var team entities.Team
		err = rows.Scan(&team.ID, &team.ShortName)
		if err != nil {
			return nil, err
		}
		teams = append(teams, team)
	}

	return teams, nil
}

func (db *RDBOperation) FetchTeamsByTournament(logger zerolog.Logger, ctx context.Context, tournamentId, tournamentType *int64) ([]entities.Team, error) {
	const query = `SELECT t.id, t.name, t.short_name FROM teams t
		JOIN tournaments_teams_link ttl ON t.id = ttl.team_id
		JOIN tournaments ts ON ttl.tournament_id = ts.id
		WHERE
		($1::INT IS NULL OR ts.id = $1) AND
		($2::INT IS NULL OR ts.type_id = $2)`

	rows, err := db.db.Query(ctx, query, *tournamentId, *tournamentType)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed rows.Scan postgresql.FetchTeamsByTournament")
		return nil, err
	}
	defer rows.Close()

	var teams []entities.Team
	for rows.Next() {
		var team entities.Team
		err = rows.Scan(&team.ID, &team.Name, &team.ShortName)
		if err != nil {
			logger.Error().Stack().Err(err).Msg("failed rows.Scan postgresql.FetchTeamsByTournament")
			return nil, err
		}
		teams = append(teams, team)
	}

	return teams, nil
}

func (db *RDBOperation) FetchPastGames(logger zerolog.Logger, ctx context.Context, teamID1, teamID2, cityID, leagueID int64, tiebreak *bool) ([]entities.GameFetch, error) {
	const query = `SELECT id, tech_loose_team_id FROM games WHERE team1_id = $1 AND team2_id = $2 AND city_id = $3 AND league_id = $4 AND date < now()
		AND CASE
		WHEN $5 = true THEN is_tiebreak = true
		WHEN $5 = false THEN is_tiebreak = false
		WHEN $5 IS NULL THEN true
		END`

	rows, err := db.db.Query(ctx, query, teamID1, teamID2, cityID, leagueID, tiebreak)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var games []entities.GameFetch
	for rows.Next() {
		var game entities.GameFetch
		err = rows.Scan(&game.ID, &game.TechLooseTeamID)
		if err != nil {
			return nil, err
		}
		games = append(games, game)
	}

	return games, nil
}

func (db *RDBOperation) FetchPastGamesTiebreak(logger zerolog.Logger, ctx context.Context, leagueId int64) ([]entities.GameTiebreak, error) {
	const query = `
	SELECT 
		g.id,
		g.city_id,
		g.place_id,
		g.date,
		g.team1_id,
		g.team2_id,
		sum(score_team1) as score_team1,
		sum(score_team2) as score_team2,
		g.league_id,
		g.tech_loose_team_id
	FROM games g
	LEFT JOIN matches m ON g.id = m.game_id
	WHERE g.league_id = $1 AND g.date < now() AND is_tiebreak = true
	GROUP BY g.id, g.city_id, g.place_id, g.date, g.team1_id, g.team2_id, g.league_id, g.tech_loose_team_id
	ORDER BY g.id DESC;`

	rows, err := db.db.Query(ctx, query, leagueId)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to postgresql.FetchPastGamesTiebreak")
		return nil, DecodeDatabaseError(err)
	}
	defer rows.Close()

	var games []entities.GameTiebreak

	for rows.Next() {
		game := entities.GameTiebreak{}
		err = rows.Scan(
			&game.Id,
			&game.CityId,
			&game.PlaceId,
			&game.Date,
			&game.Team1Id,
			&game.Team2Id,
			&game.ScoreTeam1,
			&game.ScoreTeam2,
			&game.LeagueId,
			&game.TechLooseTeamId,
		)
		if err != nil {
			logger.Error().Stack().Err(err).Msg("failed to postgresql.FetchPastGamesTiebreak")
			return nil, DecodeDatabaseError(err)
		}

		games = append(games, game)
	}

	return games, nil
}

func (db *RDBOperation) FetchTournamentPastGamesTiebreak(logger zerolog.Logger, ctx context.Context, tournamentId int64) ([]entities.GameTiebreak, error) {
	const query = `
	SELECT 
		g.id,
		g.city_id,
		g.place_id,
		g.date,
		g.team1_id,
		g.team2_id,
		sum(score_team1) as score_team1,
		sum(score_team2) as score_team2,
		t.id,
		g.tech_loose_team_id
	FROM games g
	LEFT JOIN matches m ON g.id = m.game_id
	JOIN tournament_stages ts ON g.stage_id = ts.id
	JOIN tournaments t ON ts.tournament_id = t.id
	WHERE t.id = $1 AND g.date < now() AND is_tiebreak = true
	GROUP BY g.id, g.city_id, g.place_id, g.date, g.team1_id, g.team2_id, t.id, g.tech_loose_team_id
	ORDER BY g.id DESC;`

	rows, err := db.db.Query(ctx, query, tournamentId)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to postgresql.FetchTournamentPastGamesTiebreak")
		return nil, DecodeDatabaseError(err)
	}
	defer rows.Close()

	var games []entities.GameTiebreak

	for rows.Next() {
		game := entities.GameTiebreak{}
		err = rows.Scan(
			&game.Id,
			&game.CityId,
			&game.PlaceId,
			&game.Date,
			&game.Team1Id,
			&game.Team2Id,
			&game.ScoreTeam1,
			&game.ScoreTeam2,
			&game.TournamentId,
			&game.TechLooseTeamId,
		)
		if err != nil {
			logger.Error().Stack().Err(err).Msg("failed to postgresql.FetchTournamentPastGamesTiebreak")
			return nil, DecodeDatabaseError(err)
		}

		games = append(games, game)
	}

	return games, nil
}

func (db *RDBOperation) FetchTournamentPastGames(logger zerolog.Logger, ctx context.Context, teamId1, teamId2, cityId, tournamentId int64, tiebreak *bool) ([]entities.GameFetch, error) {
	const query = `SELECT g.id, g.tech_loose_team_id FROM games g
		JOIN tournament_stages ts ON g.stage_id = ts.id
		JOIN tournaments t ON ts.tournament_id = t.id
	WHERE ((g.team1_id = $1 AND g.team2_id = $2) OR (g.team1_id = $2 AND g.team2_id = $1))
		AND
		g.city_id = $3 AND t.id = $4 AND g.date < now()
		AND CASE
		WHEN $5 = true THEN is_tiebreak = true
		WHEN $5 = false THEN is_tiebreak = false
		WHEN $5 IS NULL THEN true
		END
	ORDER BY g.id`

	rows, err := db.db.Query(ctx, query, teamId1, teamId2, cityId, tournamentId, tiebreak)

	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed rows.Scan postgresql.FetchTournamentPastGames")
		return nil, err
	}
	defer rows.Close()

	var games []entities.GameFetch
	for rows.Next() {
		var game entities.GameFetch
		err = rows.Scan(&game.ID, &game.TechLooseTeamID)
		if err != nil {
			logger.Error().Stack().Err(err).Msg("failсed rows.Scan postgresql.FetchTournamentPastGames")
			return nil, err
		}
		games = append(games, game)
	}

	return games, nil
}

func (db *RDBOperation) FetchMatches(logger zerolog.Logger, ctx context.Context, gameID int64) ([]entities.ShortMatch, error) {
	timeout, cancel := context.WithTimeout(ctx, db.cfg.MaxIdleConnectionTimeout)
	defer cancel()

	const query = `SELECT id, team1_id, team2_id, score_team1, score_team2 FROM matches
					WHERE game_id = $1
					ORDER BY id`

	rows, err := db.db.Query(timeout, query, gameID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to postgresql.FetchMatches")
		return nil, DecodeDatabaseError(err)
	}
	defer rows.Close()

	var matches []entities.ShortMatch
	for rows.Next() {
		var match entities.ShortMatch
		err = rows.Scan(
			&match.ID,
			&match.Team1ID,
			&match.Team2ID,
			&match.ScoreTeam1,
			&match.ScoreTeam2)
		if err != nil {
			logger.Error().Err(err).Msg("failed to postgresql.FetchMatches")
			return nil, DecodeDatabaseError(err)
		}
		matches = append(matches, match)
	}

	return matches, nil
}

func (db *RDBOperation) TeamsHaveNoGames(logger zerolog.Logger, ctx context.Context, teams []entities.Team, seasonID int64) (bool, error) {
	const queryGame = `SELECT g.id
					   FROM games g
					   JOIN leagues l ON g.league_id = l.id
					   WHERE (g.team1_id = $1 OR g.team2_id = $1)
					   AND l.season_id = $2
					`
	for _, team := range teams {
		rows, err := db.db.Query(ctx, queryGame, team.ID, seasonID)
		if err != nil {
			return false, err
		}
		defer rows.Close()

		var games []entities.ComingGame
		for rows.Next() {
			var game entities.ComingGame
			err = rows.Scan(&game.ID)
			if err != nil {
				return false, err
			}
			games = append(games, game)
		}
		if len(games) >= 0 {
			return false, nil
		}

	}

	return true, nil
}

func (db *RDBOperation) TeamsHaveNoGamesTournament(logger zerolog.Logger, ctx context.Context, teams []entities.Team, seasonId int64) (bool, error) {
	const queryGame = `SELECT g.id
					   FROM games g
					   JOIN tournament_stages ts ON g.stage_id = ts.id
					   JOIN tournaments t ON ts.tournament_id = t.id
					   WHERE (g.team1_id = $1 OR g.team2_id = $1)
					   AND t.season_id = $2
					`
	for _, team := range teams {
		rows, err := db.db.Query(ctx, queryGame, team.ID, seasonId)
		if err != nil {
			logger.Error().Stack().Err(err).Msg("failed rows.Scan postgresql.TeamsHaveNoGamesTournament")
			return false, err
		}
		defer rows.Close()

		var games []entities.TournamentGame
		for rows.Next() {
			var game entities.TournamentGame
			err = rows.Scan(&game.ID)
			if err != nil {
				logger.Error().Stack().Err(err).Msg("failed rows.Scan postgresql.TeamsHaveNoGamesTournament")
				return false, err
			}
			games = append(games, game)
		}
		if len(games) >= 0 {
			return false, nil
		}

	}

	return true, nil
}

func (db *RWDBOperation) CreateTeam(logger zerolog.Logger, ctx context.Context, team entities.CreateTeamRequest, tx tx.ITx) (int64, error) {
	timeout, cancel := context.WithTimeout(ctx, db.cfg.MaxIdleConnectionTimeout)
	defer cancel()

	var teamId int64
	const queryCreateTeam = `INSERT INTO teams (name, short_name, avatar, city_id) VALUES ($1, $2, $3, $4) RETURNING id`

	// Вставка команды в таблицу `teams` и получение `id` новой команды
	err := poolOrTx(db.db, tx).QueryRow(timeout, queryCreateTeam,
		team.Name,
		team.ShortName,
		team.Avatar,
		team.CityId,
	).Scan(&teamId)
	if err != nil {
		logger.Error().Err(err).Msg("failed to create Team record")
		return 0, DecodeDatabaseError(err)
	}

	return teamId, nil
}

func (db *RWDBOperation) UpdateTeam(logger zerolog.Logger, ctx context.Context, req entities.UpdateTeamRequest, cfg *config.DBConfig) (bool, error) {
	timeout, cancel := context.WithTimeout(ctx, cfg.MaxIdleConnectionTimeout)
	defer cancel()

	const query string = `
		UPDATE teams
		SET
		name = COALESCE($1, name),
		short_name = COALESCE($2, short_name),
		city_id = COALESCE($3, city_id),
		avatar = COALESCE($4, avatar)
		WHERE id = $5;`

	tag, err := db.db.Exec(timeout, query, req.Name, req.ShortName, req.CityId, req.Avatar, req.ID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to postgresql.UpdateTeam")
		return false, DecodeDatabaseError(err)
	}
	if tag.RowsAffected() == 0 {
		err = errors.New("no rows affected")
		logger.Error().Err(err).Msg("postgresql.UpdateTeam no rows affected")
		return false, errtmp.New(err.Error(), err, codes.NotFound, http.StatusNotFound)
	}

	return true, nil
}

func (db *RWDBOperation) DeleteTeam(logger zerolog.Logger, ctx context.Context, id int64, cfg *config.DBConfig) (bool, error) {
	timeout, cancel := context.WithTimeout(ctx, cfg.MaxIdleConnectionTimeout)
	defer cancel()
	//games, matches, users : в этих таблицах тоже могут быть связи с командой (team_id)
	// Запросы для удаления связанных данных
	const queryDeleteFromPlayersTeamsLinks = `DELETE FROM players_teams_links WHERE team_id = $1`
	const queryDeleteFromTeamsLeaguesLinks = `DELETE FROM teams_leagues_links WHERE team_id = $1`
	const queryDeleteTeam = `DELETE FROM teams WHERE id = $1`

	// Начинаем транзакцию
	tx, err := db.db.Begin(timeout)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to begin transaction in DeleteTeam")
		return false, DecodeDatabaseError(errors.New(pkgerr.ErrDeleteTeam))
	}

	// Удаление из `players_teams_links`
	_, err = tx.Exec(timeout, queryDeleteFromPlayersTeamsLinks, id)
	if err != nil {
		logger.Error().Err(err).Msg("failed to delete from players_teams_links")
		_ = tx.Rollback(timeout)
		return false, DecodeDatabaseError(err)
	}

	// Удаление из `teams_leagues_links`
	_, err = tx.Exec(timeout, queryDeleteFromTeamsLeaguesLinks, id)
	if err != nil {
		logger.Error().Err(err).Msg("failed to delete from teams_leagues_links")
		_ = tx.Rollback(timeout)
		return false, DecodeDatabaseError(err)
	}

	//Удаление самой команды
	tag, err := tx.Exec(timeout, queryDeleteTeam, id)
	if err != nil {
		logger.Error().Err(err).Msg("failed to delete from teams")
		_ = tx.Rollback(timeout)
		return false, DecodeDatabaseError(err)
	}

	//Проверка, что команда была удалена
	rowsAffected := tag.RowsAffected()
	if rowsAffected == 0 {
		logger.Warn().Msg("No team was deleted; it may not exist")
		_ = tx.Rollback(timeout)
		return false, errors.New("failed to delete team; it may not exist")
	}

	// Коммит транзакции
	if err = tx.Commit(timeout); err != nil {
		logger.Error().Stack().Err(err).Msg("failed to commit transaction in DeleteTeam")
		_ = tx.Rollback(timeout)
		return false, DecodeDatabaseError(errors.New(pkgerr.ErrDeleteTeam))
	}

	return true, nil
}

func (db *RWDBOperation) AddPlayerIntoTeam(logger zerolog.Logger, ctx context.Context, playerID, teamID int64, tx tx.ITx) (bool, error) {
	timeout, cancel := context.WithTimeout(ctx, db.cfg.MaxIdleConnectionTimeout)
	defer cancel()

	exec := poolOrTx(db.db, tx)

	var exists int

	const query1 = "SELECT 1 FROM players_teams_links WHERE player_id = $1 AND team_id = $2"

	err := exec.QueryRow(timeout, query1, playerID, teamID).Scan(&exists)
	if err != nil && errors.Is(err, pgx.ErrNoRows) == false {
		logger.Error().Err(err).Msg(err.Error())
		return false, DecodeDatabaseError(err)
	}

	if exists > 0 {
		err = errors.New("player already exists in team")
		logger.Error().Err(err).Msg(err.Error())
		return false, errtmp.New(err.Error(), err, codes.AlreadyExists, http.StatusConflict)
	}

	const query2 = "INSERT INTO players_teams_links(player_id, team_id) VALUES($1, $2)"

	tag, err := exec.Exec(timeout, query2, playerID, teamID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to AddPlayerIntoTeam")
		return false, DecodeDatabaseError(err)
	}
	if tag.RowsAffected() == 0 {
		err = errors.New("no rows affected")
		logger.Error().Err(err).Msg("no rows affected in AddPlayerIntoTeam")
		return false, errtmp.New(err.Error(), err, codes.NotFound, http.StatusNotFound)
	}

	return true, nil
}

func (db *RWDBOperation) RemovePlayerFromTeam(logger zerolog.Logger, ctx context.Context, playerID, teamID int64, cfg *config.DBConfig) (bool, error) {
	timeout, cancel := context.WithTimeout(ctx, cfg.MaxIdleConnectionTimeout)
	defer cancel()

	var exists int

	const query1 = "SELECT 1 FROM players_teams_links WHERE player_id = $1 AND team_id = $2"

	err := db.db.QueryRow(timeout, query1, playerID, teamID).Scan(&exists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Error().Err(err).Msg("Player not in that team!")
			return false, DecodeDatabaseError(err)
		}
		logger.Error().Err(err).Msg("failed to RemovePlayerFromTeam")
		return false, DecodeDatabaseError(err)
	}

	const query2 = "DELETE FROM players_teams_links WHERE player_id = $1 AND team_id = $2"

	tag, err := db.db.Exec(timeout, query2, playerID, teamID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to RemovePlayerFromTeam")
		return false, DecodeDatabaseError(err)
	}
	if tag.RowsAffected() == 0 {
		err = errors.New("no rows affected")
		logger.Error().Err(err).Msg("No rows affected, failed to remove player from team")
		return false, errtmp.New(err.Error(), err, codes.NotFound, http.StatusNotFound)
	}

	return true, nil
}

func (db *RDBOperation) GetTeamsByPlayerID(logger zerolog.Logger, ctx context.Context, playerID int, tx tx.ITx) ([]entities.TeamItem, error) {
	timeout, cancel := context.WithTimeout(ctx, db.cfg.MaxIdleConnectionTimeout)
	defer cancel()

	const query = `SELECT t.id, t.name, t.short_name, t.avatar, t.city_id,
	COALESCE(
		ARRAY_AGG(
			DISTINCT tll.league_id ORDER BY tll.league_id) 
				FILTER (WHERE tll.league_id IS NOT NULL), ARRAY[]::int8[]) AS leagues,
	COALESCE(
		ARRAY_AGG(
			DISTINCT ttl.tournament_id ORDER BY ttl.tournament_id) 
				FILTER (WHERE ttl.tournament_id IS NOT NULL), ARRAY[]::int8[]) AS tournaments			
	FROM teams t
	LEFT JOIN teams_leagues_links tll ON t.id = tll.team_id
	LEFT JOIN tournaments_teams_link ttl ON t.id = ttl.team_id
	JOIN players_teams_links ptl ON t.id = ptl.team_id
	WHERE ptl.player_id = $1
	GROUP BY t.id;`

	rows, err := poolOrTx(db.db, tx).Query(timeout, query, playerID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to GetTeamsByPlayerID")
		return nil, DecodeDatabaseError(err)
	}
	defer rows.Close()

	var teams []entities.TeamItem

	for rows.Next() {
		var team entities.TeamItem

		err = rows.Scan(&team.ID, &team.Name, &team.ShortName, &team.Avatar, &team.CityID, &team.Leagues, &team.Tournaments)
		if err != nil {
			logger.Error().Err(err).Msg("failed to scan team list")
			return nil, DecodeDatabaseError(err)
		}

		teams = append(teams, team)
	}

	return teams, nil
}

func (db *RDBOperation) GetTeamExtraPointsCount(logger zerolog.Logger, ctx context.Context, teamID, leagueID int64) (int64, error) {
	const query = `SELECT COALESCE(SUM(points), 0)
	FROM team_extra_points
	WHERE team_id = $1 AND league_id = $2;`

	var extraPoints int64

	err := db.db.QueryRow(ctx, query, teamID, leagueID).Scan(&extraPoints)
	if err != nil {
		logger.Error().Err(err).Msg("failed to postgresql.GetTeamExtraPoints")
		return 0, DecodeDatabaseError(err)
	}

	return extraPoints, nil
}

func (db *RDBOperation) GetTournamentTeamExtraPointsCount(logger zerolog.Logger, ctx context.Context, teamId, tournamentId int64) (int64, error) {
	const query = `SELECT COALESCE(SUM(points), 0)
	FROM team_extra_points tep
		JOIN tournaments_teams_link ttl ON tep.id = ttl.team_id
		JOIN tournaments t ON ttl.tournament_id = t.id
	WHERE tep.team_id = $1 AND t.id = $2;`

	var extraPoints int64

	err := db.db.QueryRow(ctx, query, teamId, tournamentId).Scan(&extraPoints)
	if err != nil {
		logger.Error().Err(err).Msg("failed to postgresql.GetTournamentTeamExtraPointsCount")
		return 0, DecodeDatabaseError(err)
	}

	return extraPoints, nil
}

// CreateExtraPoints deprecated
func (db *RWDBOperation) CreateExtraPoints(logger zerolog.Logger, ctx context.Context, req *entities.CreateExtraPointsRequest) (int64, error) {
	const query = "INSERT INTO team_extra_points(team_id, league_id, reason, points) VALUES($1, $2, $3, $4) RETURNING id"

	var id int64
	err := db.db.QueryRow(ctx, query, req.TeamId, req.LeagueId, req.Reason, req.Points).Scan(&id)

	if err != nil {
		logger.Error().Err(err).Msg("failed to postgresql.CreateExtraPoints")
		return 0, DecodeDatabaseError(err)
	}

	return id, nil
}

func (db *RWDBOperation) CreateExtraPointsTournament(ctx context.Context, logger zerolog.Logger, req *entities.CreateExtraPointsTournamentRequest, tx tx.ITx) (int64, error) {
	const query = "INSERT INTO tournament_team_extra_points(team_id, tournament_id, reason, points) VALUES($1, $2, $3, $4) RETURNING id"

	var id int64

	err := poolOrTx(db.db, tx).QueryRow(ctx, query, req.TeamId, req.TournamentId, req.Reason, req.Points).Scan(&id)
	if err != nil {
		logger.Error().Err(err).Msg("failed to postgresql.CreateExtraPointsTournament")
		return 0, DecodeDatabaseError(err)
	}

	return id, nil
}

// UpdateExtraPoints deprecated
func (db *RWDBOperation) UpdateExtraPoints(logger zerolog.Logger, ctx context.Context, req *entities.UpdateExtraPointsRequest) (bool, error) {
	const query = `UPDATE team_extra_points SET
		team_id = COALESCE($2, team_id),
		league_id = COALESCE($3, league_id),
		reason = COALESCE($4, reason),
		points = COALESCE($5, points),
		updated_at = now()
		WHERE id = $1`

	tag, err := db.db.Exec(ctx, query, req.Id, req.TeamId, req.LeagueId, req.Reason, req.Points)
	if err != nil {
		logger.Error().Err(err).Msg("failed to postgresql.UpdateExtraPoints")
		return false, DecodeDatabaseError(err)
	}
	if tag.RowsAffected() == 0 {
		err = errors.New("no rows affected, failed to postgresql.UpdateExtraPoints")
		logger.Error().Msg(err.Error())
		return false, DecodeDatabaseError(err)
	}
	return true, nil
}

func (db *RWDBOperation) UpdateExtraPointsTournament(ctx context.Context, logger zerolog.Logger, req *entities.UpdateExtraPointsTournamentRequest, tx tx.ITx) (bool, error) {
	const query = `
		UPDATE tournament_team_extra_points SET
			team_id = COALESCE($2, team_id),
			tournament_id = COALESCE($3, tournament_id),
			reason = COALESCE($4, reason),
			points = COALESCE($5, points),
			updated_at = now()
		WHERE id = $1`

	tag, err := poolOrTx(db.db, tx).Exec(ctx, query, req.Id, req.TeamId, req.TournamentId, req.Reason, req.Points)
	if err != nil {
		logger.Error().Err(err).Msg("failed to postgresql.UpdateExtraPointsTournament")
		return false, DecodeDatabaseError(err)
	}
	if tag.RowsAffected() == 0 {
		err = errors.New("no rows affected, failed to postgresql.UpdateExtraPointsTournament")
		logger.Error().Msg(err.Error())

		return false, DecodeDatabaseError(err)
	}
	return true, nil
}

// DeleteExtraPoints deprecated
func (db *RWDBOperation) DeleteExtraPoints(logger zerolog.Logger, ctx context.Context, extraPointsId int64) (bool, error) {
	const query = `DELETE FROM team_extra_points WHERE id = $1`

	tag, err := db.db.Exec(ctx, query, extraPointsId)
	if err != nil {
		logger.Error().Err(err).Msg("failed to postgresql.DeleteExtraPoints")
		return false, DecodeDatabaseError(err)
	}
	if tag.RowsAffected() == 0 {
		err = errors.New("no rows affected, failed to postgresql.DeleteExtraPoints")
		logger.Error().Msg(err.Error())
		return false, DecodeDatabaseError(err)
	}
	return true, nil
}

func (db *RWDBOperation) DeleteExtraPointsTournament(ctx context.Context, logger zerolog.Logger, extraPointsId int64, tx tx.ITx) (bool, error) {
	const query = `DELETE FROM tournament_team_extra_points WHERE id = $1`

	_, err := poolOrTx(db.db, tx).Exec(ctx, query, extraPointsId)
	if err != nil {
		logger.Error().Err(err).Msg("failed to postgresql.DeleteExtraPointsTournament")
		return false, DecodeDatabaseError(err)
	}

	return true, nil
}

func (db *RWDBOperation) DeleteExtraPointsByTournamentId(ctx context.Context, logger zerolog.Logger, tournamentId int64, tx tx.ITx) error {
	const query = `DELETE FROM tournament_team_extra_points WHERE tournament_id = $1`

	_, err := poolOrTx(db.db, tx).Exec(ctx, query, tournamentId)
	if err != nil {
		logger.Error().Err(err).Msg("failed to postgresql.DeleteExtraPointsByTournamentId")
		return DecodeDatabaseError(err)
	}

	return nil
}

// GetExtraPointsListByTeamAndLeagueId deprecated
func (db *RDBOperation) GetExtraPointsListByTeamAndLeagueId(logger zerolog.Logger, ctx context.Context, teamId, leagueId int64) ([]entities.ExtraPoints, error) {
	const query = `SELECT id, team_id, league_id, reason, points
	FROM team_extra_points
	WHERE team_id = $1 AND league_id = $2;`

	rows, err := db.db.Query(ctx, query, teamId, leagueId)

	if err != nil {
		logger.Error().Err(err).Msg("failed to postgresql.GetExtraPointsListByTeamAndLeagueId")
		return nil, DecodeDatabaseError(err)
	}

	var extraPointsList []entities.ExtraPoints

	for rows.Next() {
		var extraPoints entities.ExtraPoints

		err = rows.Scan(
			&extraPoints.Id,
			&extraPoints.TeamId,
			&extraPoints.LeagueId,
			&extraPoints.Reason,
			&extraPoints.Points,
		)
		if err != nil {
			logger.Error().Err(err).Msg("failed to scan extraPoints list")
			return nil, DecodeDatabaseError(err)
		}

		extraPointsList = append(extraPointsList, extraPoints)
	}

	return extraPointsList, nil
}

func (db *RDBOperation) GetExtraPointsListByTeamAndTournamentId(ctx context.Context, logger zerolog.Logger, teamId, tournamentId int64) ([]entities.ExtraPointsTournament, error) {
	const query = `
		SELECT
		    id,
		    team_id,
		    tournament_id,
		    reason,
		    points
		FROM tournament_team_extra_points
		WHERE team_id = $1 AND tournament_id = $2;`

	rows, err := db.db.Query(ctx, query, teamId, tournamentId)
	if err != nil {
		logger.Error().Err(err).Msg("failed to postgresql.GetExtraPointsListByTeamAndTournamentId")
		return nil, DecodeDatabaseError(err)
	}

	var extraPointsList []entities.ExtraPointsTournament

	for rows.Next() {
		var extraPoints entities.ExtraPointsTournament

		err = rows.Scan(
			&extraPoints.Id,
			&extraPoints.TeamId,
			&extraPoints.TournamentId,
			&extraPoints.Reason,
			&extraPoints.Points,
		)
		if err != nil {
			logger.Error().Err(err).Msg("failed to scan extraPoints list")
			return nil, DecodeDatabaseError(err)
		}

		extraPointsList = append(extraPointsList, extraPoints)
	}

	return extraPointsList, nil
}

// GetExtraPointsById deprecated
func (db *RDBOperation) GetExtraPointsById(logger zerolog.Logger, ctx context.Context, extraPointsId int64) (entities.ExtraPoints, error) {
	const query = `SELECT id, team_id, league_id, reason, points
	FROM team_extra_points
	WHERE id = $1`

	var extraPoints entities.ExtraPoints

	err := db.db.QueryRow(ctx, query, extraPointsId).Scan(
		&extraPoints.Id,
		&extraPoints.TeamId,
		&extraPoints.LeagueId,
		&extraPoints.Reason,
		&extraPoints.Points,
	)

	if err != nil {
		logger.Error().Err(err).Msg("failed to postgresql.GetExtraPointsById")
		return entities.ExtraPoints{}, DecodeDatabaseError(err)
	}

	return extraPoints, nil
}

func (db *RDBOperation) GetExtraPointsTournamentById(ctx context.Context, logger zerolog.Logger, extraPointsId int64) (entities.ExtraPointsTournament, error) {
	const query = `
		SELECT
		    id,
		    team_id,
		    tournament_id,
		    reason,
		    points
		FROM tournament_team_extra_points
		WHERE id = $1`

	var extraPoints entities.ExtraPointsTournament

	err := db.db.QueryRow(ctx, query, extraPointsId).Scan(
		&extraPoints.Id,
		&extraPoints.TeamId,
		&extraPoints.TournamentId,
		&extraPoints.Reason,
		&extraPoints.Points,
	)
	if err != nil {
		logger.Error().Err(err).Msg("failed to postgresql.GetExtraPointsTournamentById")
		return entities.ExtraPointsTournament{}, DecodeDatabaseError(err)
	}

	return extraPoints, nil
}

func (db *RDBOperation) GetTeamById(logger zerolog.Logger, ctx context.Context, teamID int64, tx tx.ITx) (entities.TeamV2, error) {
	timeout, cancel := context.WithTimeout(ctx, db.cfg.MaxIdleConnectionTimeout)
	defer cancel()

	team := entities.TeamV2{}

	const queryGetTeam = `
	SELECT 
		t.id, 
		t.name, 
		t.short_name, 
		t.city_id,
		t.avatar, 
		COALESCE(ARRAY_AGG(DISTINCT ptl.player_id ORDER BY ptl.player_id) FILTER (WHERE ptl.player_id IS NOT NULL), ARRAY[]::BIGINT[]) AS players_ids
	FROM teams t
		LEFT JOIN players_teams_links ptl ON t.id = ptl.team_id
	WHERE t.id = $1
	GROUP BY t.id`

	err := poolOrTx(db.db, tx).QueryRow(timeout, queryGetTeam, teamID).
		Scan(&team.Id, &team.Name, &team.ShortName, &team.CityId, &team.Avatar, &team.PlayersIds)
	if err != nil {
		logger.Error().Err(err).Msg("failed to postgresql.GetTeamById")
		return entities.TeamV2{}, DecodeDatabaseError(err)
	}

	return team, nil
}

func (db *RDBOperation) GetLeagueListByTeamId(logger zerolog.Logger, ctx context.Context, teamID int64) ([]entities.LeagueShort, error) {
	const query = `
		SELECT 
			l.id,
			l.name
		FROM leagues l
			LEFT JOIN teams_leagues_links tll ON l.id = tll.league_id
		WHERE tll.team_id = $1;`

	var leagues []entities.LeagueShort

	rows, err := db.db.Query(ctx, query, teamID)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to postgresql.GetLeagueListByTeamId")
		return nil, DecodeDatabaseError(errors.New(pkgerr.ErrGetPlayer))
	}
	defer rows.Close()

	for rows.Next() {
		league := entities.LeagueShort{}
		if err = rows.Scan(&league.ID, &league.Name); err != nil {
			logger.Error().Stack().Err(err).Msg("failed to postgresql.GetLeagueListByTeamId")
			return nil, DecodeDatabaseError(err)
		}
		leagues = append(leagues, league)
	}

	return leagues, nil
}

func (db *RDBOperation) GetCaptainByTeamId(logger zerolog.Logger, ctx context.Context, teamID int64) (entities.User, error) {
	const query = `
		SELECT
			u.id,
			u.email,
			u.role_id,
			r.name,
			r.description
		FROM users u
			JOIN teams t ON t.id = u.team_id
			JOIN user_roles r ON r.id = u.role_id
		WHERE t.id = $1 AND u.role_id = $2;`

	var captain entities.User
	captain.Role = &entities.Role{}

	err := db.db.QueryRow(ctx, query, teamID, constant.CaptainRoleId).
		Scan(&captain.ID, &captain.Email, &captain.Role.ID, &captain.Role.Name, &captain.Role.Description)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to postgresql.GetCaptainByTeamId")
		return entities.User{}, DecodeDatabaseError(err)
	}

	return captain, nil
}

func (db *RDBOperation) GetTeamGamesInLeague(logger zerolog.Logger, ctx context.Context, teamId, leagueId int64, tiebreak *bool) ([]entities.Game, error) {
	const query = `
		SELECT 
			g.id,
			g.city_id,
			g.place_id,
			g.date,
			g.team1_id,
			g.team2_id,
			g.league_id,
			g.tech_loose_team_id,
			g.team1_id = $1 AS is_home_game
		FROM games g
		WHERE (g.team1_id = $1 OR g.team2_id = $1) AND g.league_id = $2
		AND CASE
		WHEN $3 = true THEN is_tiebreak = true
		WHEN $3 = false THEN is_tiebreak = false
		WHEN $3 IS NULL THEN true
		END;`

	rows, err := db.db.Query(ctx, query, teamId, leagueId, tiebreak)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to postgresql.GetTeamGamesInLeague")
		return nil, DecodeDatabaseError(err)
	}
	defer rows.Close()

	var games []entities.Game

	for rows.Next() {
		game := entities.Game{}
		err = rows.Scan(
			&game.Id,
			&game.CityId,
			&game.PlaceId,
			&game.Date,
			&game.Team1Id,
			&game.Team2Id,
			&game.LeagueId,
			&game.TechLooseTeamId,
			&game.IsHomeGame)
		if err != nil {
			logger.Error().Stack().Err(err).Msg("failed to postgresql.GetTeamGamesInLeague")
			return nil, DecodeDatabaseError(err)
		}

		games = append(games, game)
	}

	return games, nil
}

func (db *RDBOperation) GetTournamentTeamList(logger zerolog.Logger, ctx context.Context, tournamentID int64, cfg *config.DBConfig) ([]entities.TournamentTeam, error) {
	timeout, cancel := context.WithTimeout(ctx, cfg.MaxIdleConnectionTimeout)
	defer cancel()

	const query string = `
		SELECT t.id, t.name, t.short_name, t.city_id, t.avatar 
		FROM teams AS t
		JOIN tournaments_teams_link AS ttl ON t.id = ttl.team_id
		WHERE ttl.tournament_id = $1;`

	rows, err := db.db.Query(timeout, query, tournamentID)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to postgresql.GetTournamentTeamList")
		return nil, DecodeDatabaseError(err)
	}
	defer rows.Close()

	var tournamentTeams []entities.TournamentTeam

	for rows.Next() {
		tt := entities.TournamentTeam{}
		err = rows.Scan(&tt.ID, &tt.Name, &tt.ShortName, &tt.CityID, &tt.Avatar)
		if err != nil {
			logger.Error().Stack().Err(err).Msg("failed to postgresql.GetTournamentTeamList")
			return nil, DecodeDatabaseError(err)
		}
		tournamentTeams = append(tournamentTeams, tt)
	}

	return tournamentTeams, nil
}
