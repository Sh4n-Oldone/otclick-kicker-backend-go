package postgresql

import (
	"context"
	stderr "errors"

	"github.com/rs/zerolog"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/entity"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/errors"

	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

func (db *RDBOperation) GetTeam(logger zerolog.Logger, ctx context.Context, teamID int64) (entity.GetTeamResponse, error) {
	team := entity.GetTeamResponse{}

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
	err := db.db.QueryRow(ctx, queryGetTeam, teamID).Scan(&team.ID, &team.Name, &team.ShortName, &team.Avatar, &team.CityId, &leagueIDs, &playerIDs)
	if err != nil {
		logger.Error().Err(err).Msg("failed to GetTeam")
		return entity.GetTeamResponse{}, DecodeDatabaseError(err)
	}
	// ///////////////////////////////////////////////////////////////////////////////////////////////////////////////
	const queryGetLeagueByID = `SELECT name FROM leagues WHERE id = $1;`
	var leagues = []entity.LeagueShort{}
	for _, leagueID := range leagueIDs {
		var l entity.LeagueShort

		l.ID = leagueID

		err := db.db.QueryRow(ctx, queryGetLeagueByID, leagueID).
			Scan(&l.Name)
		if err != nil {
			logger.Error().Stack().Err(err).Msg("failed to postgresql.GetTeam/queryGetPlayerByID")
			return entity.GetTeamResponse{}, DecodeDatabaseError(stderr.New(errors.ErrGetPlayer))
		}

		leagues = append(leagues, l)
	}
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////
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
    WHERE p.id = $1;
	`
	// возможно, проще было сделать через слияние SQL запросов
	if len(leagueIDs) == 0 {
		queryGetPlayerByID = fmt.Sprintf(queryGetPlayerByID, "")
	} else {
		queryGetPlayerByID = fmt.Sprintf(queryGetPlayerByID, "AND r.league_id IN ("+strings.Trim(strings.ReplaceAll(fmt.Sprint(leagueIDs), " ", ", "), "[]")+")")
	}
	// альтернативные варианты перевода массива чисел в строку
	//strings.Trim(strings.Join(strings.Fields(fmt.Sprint(leagueIDs)), ", "), "[]")
	//strings.Trim(strings.Join(strings.Split(fmt.Sprint(leagueIDs), " "), ", "), "[]")
	var players = []entity.PlayerGetTeam{}
	for _, playerID := range playerIDs {
		var p entity.PlayerGetTeam

		err := db.db.QueryRow(ctx, queryGetPlayerByID, playerID).
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
			return entity.GetTeamResponse{}, DecodeDatabaseError(stderr.New(errors.ErrGetPlayer))
		}
		players = append(players, p)
	}

	team.Leagues = leagues
	team.Players = players

	return team, nil
}

func (db *RDBOperation) GetTeams(logger zerolog.Logger, ctx context.Context, cityID int64, onlyFree bool) ([]entity.TeamShort, error) {
	query := `SELECT t.id, t.name, t.short_name FROM teams t
		LEFT JOIN teams_leagues_links tll ON t.id = tll.team_id
		WHERE t.city_id = $1`
	if onlyFree {
		query += " AND tll.league_id IS NULL"
	}
	query += " ORDER BY id"

	rows, err := db.db.Query(ctx, query, cityID) // , onlyFree
	if err != nil {
		logger.Error().Err(err).Msg("failed to GetTeams")
		return nil, DecodeDatabaseError(err)
	}
	defer rows.Close()

	var teams []entity.TeamShort

	for rows.Next() {
		var team entity.TeamShort

		err = rows.Scan(&team.ID, &team.Name, &team.ShortName)
		if err != nil {
			logger.Error().Err(err).Msg("failed to scan team list")
			return nil, DecodeDatabaseError(err)
		}

		teams = append(teams, team)
	}

	return teams, nil
}

func (db *RDBOperation) GetTeamsByCity(logger zerolog.Logger, ctx context.Context, onlyFree bool, cityID int64) ([]entity.TeamShort, error) {
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

	var teams []entity.TeamShort

	for rows.Next() {
		var team entity.TeamShort

		err = rows.Scan(&team.ID, &team.Name, &team.ShortName)
		if err != nil {
			logger.Error().Err(err).Msg("failed to scan team list")
			return nil, DecodeDatabaseError(err)
		}

		teams = append(teams, team)
	}

	return teams, nil
}

func (db *RDBOperation) GetTeamsByLeague(logger zerolog.Logger, ctx context.Context, leagueID int64) ([]entity.TeamByLeague, error) {
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

	var teams []entity.TeamByLeague

	for rows.Next() {
		var team entity.TeamByLeague

		err = rows.Scan(&team.Id, &team.Name, &team.ShortName, &team.Avatar, &team.CityId, &team.Players)
		if err != nil {
			logger.Error().Err(err).Msg("failed to scan team list")
			return nil, DecodeDatabaseError(err)
		}

		teams = append(teams, team)
	}

	return teams, nil
}

// ///////////////////////////////////////////////////////////////////////////////////
// ///////////////////////////////////////////////////////////////////////////////////
// ///////////////////////////////////////////////////////////////////////////////////

func (db *RDBOperation) FetchLeagues(logger zerolog.Logger, ctx context.Context, cityID int64, seasonID int64) ([]entity.League, error) {
	rows, err := db.db.Query(ctx, "SELECT id, name FROM leagues WHERE city_id = $1 AND season_id = $2", cityID, seasonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var leagues []entity.League
	for rows.Next() {
		var league entity.League
		err = rows.Scan(&league.ID, &league.Name)
		if err != nil {
			return nil, err
		}
		leagues = append(leagues, league)
	}

	return leagues, nil
}

func (db *RDBOperation) FetchTeams(logger zerolog.Logger, ctx context.Context, leagueID int64) ([]entity.Team, error) {
	const query = `SELECT t.id, t.short_name FROM teams t 
		LEFT JOIN teams_leagues_links tll ON t.id = tll.team_id
		WHERE tll.league_id = $1`

	rows, err := db.db.Query(ctx, query, leagueID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []entity.Team
	for rows.Next() {
		var team entity.Team
		err = rows.Scan(&team.ID, &team.ShortName)
		if err != nil {
			return nil, err
		}
		teams = append(teams, team)
	}

	return teams, nil
}

func (db *RDBOperation) FetchPastGames(logger zerolog.Logger, ctx context.Context, teamID1, teamID2, cityID, leagueID int64) ([]entity.GameFetch, error) {
	const query = `SELECT id, tech_loose_team_id FROM games WHERE team1_id = $1 AND team2_id = $2 AND city_id = $3 AND league_id = $4 AND date < now()`

	rows, err := db.db.Query(ctx, query, teamID1, teamID2, cityID, leagueID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var games []entity.GameFetch
	for rows.Next() {
		var game entity.GameFetch
		err = rows.Scan(&game.ID, &game.TechLooseTeamID)
		if err != nil {
			return nil, err
		}
		games = append(games, game)
	}

	return games, nil
}

func (db *RDBOperation) FetchMatches(logger zerolog.Logger, ctx context.Context, gameID int64) ([]entities.Match, error) {
	const query = `SELECT team1_id, team2_id, score_team1, score_team2 FROM matches
					WHERE game_id = $1
					ORDER BY id`

	rows, err := db.db.Query(ctx, query, gameID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var matches []entities.Match
	for rows.Next() {
		var match entities.Match
		err = rows.Scan(
			&match.Team1ID,
			&match.Team2ID,
			&match.ScoreTeam1,
			&match.ScoreTeam2)
		if err != nil {
			return nil, err
		}
		matches = append(matches, match)
	}

	return matches, nil
}

// /////////////////////////
func (db *RDBOperation) TeamsHaveNoGames(logger zerolog.Logger, ctx context.Context, teams []entity.Team, seasonID int64) (bool, error) {
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

// //////////////////////////////////////////////////////////////////////////////////
// //////////////////////////////////////////////////////////////////////////////////
// //////////////////////////////////////////////////////////////////////////////////
func (db *RWDBOperation) CreateTeam(logger zerolog.Logger, ctx context.Context, team entity.CreateTeamRequest) (int64, error) {
	var teamId int64
	const queryCreateTeam = `INSERT INTO teams (name, short_name, avatar, city_id) VALUES ($1, $2, $3, $4) RETURNING id`

	tx, err := db.db.Begin(ctx)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to postgresql.CreateTeam")
		return 0, DecodeDatabaseError(stderr.New(errors.ErrCreateTeam))
	}

	// Вставка команды в таблицу `teams` и получение `id` новой команды
	err = tx.QueryRow(ctx, queryCreateTeam,
		team.Name,
		team.ShortName,
		team.Avatar,
		team.CityId,
	).Scan(&teamId)
	if err != nil {
		logger.Error().Err(err).Msg("failed to create Team record")
		return 0, DecodeDatabaseError(err)
	}

	if err = tx.Commit(ctx); err != nil {
		logger.Error().Stack().Err(err).Msg("failed to postgresql.CreateTeam")
		_ = tx.Rollback(ctx)
		return 0, DecodeDatabaseError(stderr.New(errors.ErrCreateTeam))
	}

	return teamId, nil
}

func (db *RWDBOperation) UpdateTeam(logger zerolog.Logger, ctx context.Context, team entity.UpdateTeamRequest) (bool, error) {
	var fields []string
	var values []interface{}
	index := 1

	// Проверяем и добавляем поля для обновления
	if team.Name != nil {
		fields = append(fields, fmt.Sprintf("name = $%d", index))
		values = append(values, *team.Name)
		index++
	}
	if team.ShortName != nil {
		fields = append(fields, fmt.Sprintf("short_name = $%d", index))
		values = append(values, *team.ShortName)
		index++
	}
	if team.Avatar != nil {
		fields = append(fields, fmt.Sprintf("avatar = $%d", index))
		values = append(values, team.Avatar)
		index++
	}
	if team.CityId != nil {
		fields = append(fields, fmt.Sprintf("city_id = $%d", index))
		values = append(values, *team.CityId)
		index++
	}

	// Если нет полей для обновления, возвращаем предупреждение
	if len(fields) == 0 {
		logger.Warn().Msg("No fields to update")
		return false, nil
	}

	// Создание запроса для обновления полей таблицы `teams`
	if len(fields) > 0 {
		fieldsQuery := strings.Join(fields, ", ")
		queryUpdateTeam := fmt.Sprintf("UPDATE teams SET %s WHERE id = $%d", fieldsQuery, index)
		values = append(values, team.ID)

		// Выполняем запрос обновления команды
		result, err := db.db.Exec(ctx, queryUpdateTeam, values...)
		if err != nil {
			logger.Error().Err(err).Msg("failed to update Team record")
			return false, DecodeDatabaseError(err)
		}

		rowsAffected := result.RowsAffected()
		if rowsAffected == 0 {
			logger.Error().Err(err).Msg("no rows were affected for the team update")
			return false, stderr.New("Failed to Update Team, it does not exist")
		}
	}

	return true, nil
}

func (db *RWDBOperation) DeleteTeam(logger zerolog.Logger, ctx context.Context, id int64) (bool, error) {
	//games, matches, users : в этих таблицах тоже могут быть связи с командой (team_id)
	// Запросы для удаления связанных данных
	const queryDeleteFromPlayersTeamsLinks = `DELETE FROM players_teams_links WHERE team_id = $1`
	const queryDeleteFromTeamsLeaguesLinks = `DELETE FROM teams_leagues_links WHERE team_id = $1`
	const queryDeleteTeam = `DELETE FROM teams WHERE id = $1`

	// Начинаем транзакцию
	tx, err := db.db.Begin(ctx)
	if err != nil {
		logger.Error().Stack().Err(err).Msg("failed to begin transaction in DeleteTeam")
		return false, DecodeDatabaseError(stderr.New(errors.ErrDeleteTeam))
	}

	// Удаление из `players_teams_links`
	_, err = tx.Exec(ctx, queryDeleteFromPlayersTeamsLinks, id)
	if err != nil {
		logger.Error().Err(err).Msg("failed to delete from players_teams_links")
		_ = tx.Rollback(ctx)
		return false, DecodeDatabaseError(err)
	}

	// Удаление из `teams_leagues_links`
	_, err = tx.Exec(ctx, queryDeleteFromTeamsLeaguesLinks, id)
	if err != nil {
		logger.Error().Err(err).Msg("failed to delete from teams_leagues_links")
		_ = tx.Rollback(ctx)
		return false, DecodeDatabaseError(err)
	}

	//Удаление самой команды
	tag, err := tx.Exec(ctx, queryDeleteTeam, id)
	if err != nil {
		logger.Error().Err(err).Msg("failed to delete from teams")
		_ = tx.Rollback(ctx)
		return false, DecodeDatabaseError(err)
	}

	//Проверка, что команда была удалена
	rowsAffected := tag.RowsAffected()
	if rowsAffected == 0 {
		logger.Warn().Msg("No team was deleted; it may not exist")
		_ = tx.Rollback(ctx)
		return false, stderr.New("Failed to delete team; it may not exist")
	}

	// Коммит транзакции
	if err = tx.Commit(ctx); err != nil {
		logger.Error().Stack().Err(err).Msg("failed to commit transaction in DeleteTeam")
		_ = tx.Rollback(ctx)
		return false, DecodeDatabaseError(stderr.New(errors.ErrDeleteTeam))
	}

	return true, nil
}

// ////////////////////////////////////////////////////////////////////////////////
func (db *RWDBOperation) AddPlayerIntoTeam(logger zerolog.Logger, ctx context.Context, playerID, teamID int64) (bool, error) {
	var exists int
	const query1 = "SELECT 1 FROM players_teams_links WHERE player_id = $1 AND team_id = $2"
	err := db.db.QueryRow(ctx, query1, playerID, teamID).Scan(&exists)
	if err == nil {
		logger.Error().Err(err).Msg("Player already in that team!")
		return false, stderr.New("Player already in that team!")
	}
	if err != pgx.ErrNoRows {
		logger.Error().Err(err).Msg("Team or player not found")
		return false, DecodeDatabaseError(err)
	}

	const query2 = "INSERT INTO players_teams_links(player_id, team_id) VALUES($1, $2)"
	tag, err := db.db.Exec(ctx, query2, playerID, teamID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to AddPlayerIntoTeam")
		return false, DecodeDatabaseError(err)
	}
	if tag.RowsAffected() == 0 {
		logger.Error().Msg("No rows affected, failed to insert player into team")
		return false, stderr.New("Failed to add player into team")
	}

	return true, nil
}

func (db *RWDBOperation) RemovePlayerFromTeam(logger zerolog.Logger, ctx context.Context, playerID, teamID int64) (bool, error) {
	var exists int
	const query1 = "SELECT 1 FROM players_teams_links WHERE player_id = $1 AND team_id = $2"
	err := db.db.QueryRow(ctx, query1, playerID, teamID).Scan(&exists)
	if err == pgx.ErrNoRows {
		logger.Error().Msg("Player not in that team!")
		return false, stderr.New("Player not in that team!")
	}
	if err != nil {
		logger.Error().Err(err).Msg("Team or player not found")
		return false, DecodeDatabaseError(err)
	}

	const query2 = "DELETE FROM players_teams_links WHERE player_id = $1 AND team_id = $2"
	tag, err := db.db.Exec(ctx, query2, playerID, teamID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to RemovePlayerFromTeam")
		return false, DecodeDatabaseError(err)
	}
	if tag.RowsAffected() == 0 {
		logger.Error().Msg("No rows affected, failed to remove player from team")
		return false, stderr.New("failed to remove player from team")
	}

	return true, nil
}

func (db *RDBOperation) GetTeamsByPlayerID(logger zerolog.Logger, ctx context.Context, playerID int) ([]entities.TeamItem, error) {
	const query = `SELECT t.id, t.name, t.short_name, t.avatar, t.city_id,
	COALESCE(
		ARRAY_AGG(
			DISTINCT tll.league_id ORDER BY tll.league_id) 
				FILTER (WHERE tll.league_id IS NOT NULL), ARRAY[]::int8[]) AS leagues
	FROM teams t
	LEFT JOIN teams_leagues_links tll ON t.id = tll.team_id
	JOIN players_teams_links ptl ON t.id = ptl.team_id
	WHERE ptl.player_id = $1
	GROUP BY t.id;`

	rows, err := db.db.Query(ctx, query, playerID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to GetTeamsByPlayerID")
		return nil, DecodeDatabaseError(err)
	}
	defer rows.Close()

	var teams []entities.TeamItem

	for rows.Next() {
		var team entities.TeamItem

		err = rows.Scan(&team.ID, &team.Name, &team.ShortName, &team.Avatar, &team.CityID, &team.Leagues)
		if err != nil {
			logger.Error().Err(err).Msg("failed to scan team list")
			return nil, DecodeDatabaseError(err)
		}

		teams = append(teams, team)
	}

	return teams, nil
}
