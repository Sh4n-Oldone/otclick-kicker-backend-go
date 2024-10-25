package postgresql

const (
	// City queries -->
	queryGetCityList string = `
		SELECT id, name, ru, deleted_at 
		FROM cities
		WHERE deleted_at IS NULL
		ORDER BY id;`

	queryGetCityListWithDeleted string = `
		SELECT id, name, ru, deleted_at 
		FROM cities
		ORDER BY id;`

	queryCreateCity string = `
		INSERT INTO cities (name, ru) 
		VALUES 
		(
			$1,
			$2
		) RETURNING id;`

	queryUpdateCity string = `
		UPDATE cities
		SET 
		    name = $1,
    		ru = $2,
			deleted = $3
		WHERE id = $4;`

	queryDeleteCity string = `UPDATE cities SET deleted = TRUE WHERE id = $1;`
	// <--

	// User queries
	queryGetUserByID string = `
		SELECT 
			u.id, 
			u.email,
			u.password,
			coalesce(r.id, 0),
			r.name,
			coalesce(t.id, 0)
		FROM users as u
			LEFT JOIN user_roles as r ON u.role_id = r.id
			LEFT JOIN teams as t ON u.team_id = t.id
		WHERE u.id = $1;`

	queryGetUserByEmail string = `
		SELECT 
			u.id, 
			u.email,
			u.password,
			coalesce(r.id, 0),
			r.name,
			coalesce(t.id, 0)
		FROM users as u
			LEFT JOIN user_roles as r ON u.role_id = r.id
			LEFT JOIN teams as t ON u.team_id = t.id
		WHERE u.email = $1;`

	queryCreateUser string = `
		INSERT INTO users (email, password, role_id)
		VALUES
		(
			$1,
			$2,
			$3
		) RETURNING id;`

	queryUpdateUser string = `
		UPDATE users 
		SET 
			email = COALESCE($1, email), 
			password = COALESCE($2, password), 
			role_id = COALESCE($3, role_id), 
			team_id = COALESCE($4, team_id)
		WHERE id = $5;`
	// <--

	// Role queries -->
	queryGetRoleList string = `SELECT id, name, description, updated_at FROM user_roles;`

	queryGetRoleByID string = `SELECT id, name, description, updated_at FROM user_roles WHERE id = $1;`

	queryGetRoleByName string = `SELECT id, name, description, updated_at FROM user_roles WHERE name = $1;`
	// <--

	// Match queries -->
	queryCreateMatch string = `
	INSERT INTO matches (
		date,
		game_id,
		team1_id,
		team2_id,
		player1_team1_id,
		player2_team1_id,
		player1_team2_id,
		player2_team2_id,
		score_team1,
		score_team2
	)
	VALUES (
		$1,
		$2,
		$3,
		$4,
		$5,
		$6,
		$7,
		$8,
		$9,
		$10
		) RETURNING id`

	queryUpdateMatch string = `
		UPDATE matches 
		SET
			date = COALESCE($1, date),
			game_id = COALESCE($2, game_id),
			team1_id = COALESCE($3, team1_id),
			team2_id = COALESCE($4, team2_id),
			player1_team1_id = COALESCE($5, player1_team1_id),
			player2_team1_id = COALESCE($6, player2_team1_id),
			player1_team2_id = COALESCE($7, player1_team2_id),
			player2_team2_id = COALESCE($8, player2_team2_id),
			score_team1 = COALESCE($9, score_team1),
			score_team2 = COALESCE($10, score_team2)
		WHERE id = $11;`

	queryDeleteMatch string = `DELETE FROM matches WHERE id = $1`
	// <--

	// League queries -->
	queryGetLeagueList string = `
		SELECT id, name, city_id 
		FROM leagues
		WHERE city_id = $1
		ORDER BY id;`

	queryCreateLeague string = `
		INSERT INTO leagues (name, city_id) 
		VALUES 
		(
			$1,
			$2
		) RETURNING id;`

	queryUpdateLeague string = `
		UPDATE leagues
		SET 
		    name = $1
		WHERE id = $2;
`

	queryUpdateTeamsLeagueID = `
		UPDATE teams
		SET 
		    league_id = $1
		WHERE id = $2;
`

	queryGetTeamLeagueID = `
		SELECT league_id
		FROM teams
		WHERE id = $1;
`

	queryDeleteTeamsLeagueID = `
		UPDATE teams 
		SET 
		    league_id = NULL
		WHERE league_id = $1;
`

	queryDeleteLeague string = `DELETE FROM leagues WHERE id = $1;`
	// <--

	queryGetGamesYears string = `    
	SELECT DISTINCT EXTRACT(YEAR FROM date) AS year
    FROM games
    ORDER BY year ASC;
`
)
