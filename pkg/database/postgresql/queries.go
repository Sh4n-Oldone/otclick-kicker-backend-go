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
    		ru = $2
		WHERE id = $3;`

	queryDeleteCity string = `UPDATE cities SET deleted_at = NOW() WHERE id = $1;`
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

	// Table queries -->
	queryGetTableList string = `
		SELECT id, name, updated_at, deleted_at 
		FROM tables
		WHERE deleted_at IS NULL
		ORDER BY id;`

	queryGetTableListWithDeleted string = `
		SELECT id, name, updated_at, deleted_at 
		FROM tables
		ORDER BY id;`

	queryGetTableByID string = `
		SELECT id, name, updated_at, deleted_at
		FROM tables
		WHERE id = $1;`

	queryCreateTable string = `
		INSERT INTO tables (name, updated_at) 
		VALUES 
		(
			$1,
			NOW()
		) RETURNING id;`

	queryUpdateTable string = `
		UPDATE tables
		SET 
		    name = COALESCE($2, name),
			updated_at = NOW()
		WHERE id = $1;`

	queryDeleteTable string = `UPDATE tables SET deleted_at = NOW() WHERE id = $1;`
	// <--

	// Bar queries -->
	queryGetBarList string = `
		SELECT id, city_id, name, description, updated_at, deleted_at 
		FROM bars
		WHERE deleted_at IS NULL
		ORDER BY id;`

	queryGetBarListByCityID string = `
		SELECT id, city_id, name, description, updated_at, deleted_at 
		FROM bars 
		WHERE city_id = $1 AND deleted_at IS NULL
		ORDER BY id;`

	queryGetBarListWithDeleted string = `
		SELECT id, city_id, name, description, updated_at, deleted_at 
		FROM bars
		ORDER BY id;`

	queryGetBarListByCityIDWithDeleted string = `
		SELECT id, city_id, name, description, updated_at, deleted_at 
		FROM bars 
		WHERE city_id = $1
		ORDER BY id;`

	queryGetBarByID string = `
		SELECT id, city_id, name, description, updated_at, deleted_at
		FROM bars
		WHERE id = $1;`

	queryCreateBar string = `
		INSERT INTO bars (city_id, name, description, updated_at) 
		VALUES 
		(
			$1,
			$2,
			$3,
			NOW()
		) RETURNING id;`

	queryUpdateBar string = `
		UPDATE cities
		SET 
			city_id = COALESCE($2, city_id),
		    name = COALESCE($3, name),
    		description = COALESCE($4, description),
			updated_at = NOW()
		WHERE id = $1;`

	queryDeleteBar string = `UPDATE bars SET deleted_at = NOW() WHERE id = $1;`
	// <--

	// Place queries -->
	queryGetPlaceList string = `
		SELECT p.id, p.bar_id, b.bar_name, p.table_id, t.table_name, p.updated_at, p.deleted_at 
		FROM places as p
		JOIN tables as t ON places.table_id = tables.id
		JOIN bars as b ON tables.bar_id = bars.id
		WHERE p.deleted_at IS NULL
		ORDER BY p.bar_id, p.table_id, p.id;`

	queryGetPlaceListWithDeleted string = `
		SELECT p.id, p.bar_id, b.bar_name, p.table_id, t.table_name, p.updated_at, p.deleted_at 
		FROM places as p
		JOIN tables as t ON places.table_id = tables.id
		JOIN bars as b ON tables.bar_id = bars.id
		ORDER BY p.bar_id, p.table_id, p.id;`

	queryGetPlaceListByBarID string = `
		SELECT p.id, p.bar_id, b.bar_name, p.table_id, t.table_name, p.updated_at, p.deleted_at 
		FROM places as p
		JOIN tables as t ON places.table_id = tables.id
		JOIN bars as b ON tables.bar_id = bars.id
		WHERE p.bar_id = $1 AND p.deleted_at IS NULL
		ORDER BY p.bar_id, p.table_id, p.id;`

	queryGetPlaceListByBarIDWithDeleted string = `
		SELECT p.id, p.bar_id, b.bar_name, p.table_id, t.table_name, p.updated_at, p.deleted_at 
		FROM places as p
		JOIN tables as t ON places.table_id = tables.id
		JOIN bars as b ON tables.bar_id = bars.id
		WHERE p.bar_id = $1
		ORDER BY p.bar_id, p.table_id, p.id;`

	queryGetBarListByTableID string = `
		SELECT p.id, p.bar_id, b.bar_name, p.table_id, t.table_name, p.updated_at, p.deleted_at 
		FROM places as p
		JOIN tables as t ON places.table_id = tables.id
		JOIN bars as b ON tables.bar_id = bars.id
		WHERE p.table_id = $1 AND p.deleted_at IS NULL
		ORDER BY p.bar_id, p.table_id, p.id;`

	queryGetBarListByTableIDWithDeleted string = `
		SELECT p.id, p.bar_id, b.bar_name, p.table_id, t.table_name, p.updated_at, p.deleted_at 
		FROM places as p
		JOIN tables as t ON places.table_id = tables.id
		JOIN bars as b ON tables.bar_id = bars.id
		WHERE p.table_id = $1
		ORDER BY p.bar_id, p.table_id, p.id;`

	queryGetBarListByBarIDByTableID string = `
		SELECT p.id, p.bar_id, b.bar_name, p.table_id, t.table_name, p.updated_at, p.deleted_at 
		FROM places as p
		JOIN tables as t ON places.table_id = tables.id
		JOIN bars as b ON tables.bar_id = bars.id
		WHERE p.bar_id = $1 AND p.table_id = $2 AND p.deleted_at IS NULL
		ORDER BY p.bar_id, p.table_id, p.id;`

	queryGetBarListByBarIDByTableIDWithDeleted string = `
		SELECT p.id, p.bar_id, b.bar_name, p.table_id, t.table_name, p.updated_at, p.deleted_at 
		FROM places as p
		JOIN tables as t ON places.table_id = tables.id
		JOIN bars as b ON tables.bar_id = bars.id
		WHERE p.bar_id = $1 AND p.table_id = $2
		ORDER BY p.bar_id, p.table_id, p.id;`

	queryGetPlaceByID string = `
		SELECT id, bar_id, table_id, updated_at, deleted_at 
		FROM places
		WHERE id = $1 AND deleled_at IS NULL;`

	queryCreatePlace string = `
		INSERT INTO places (bar_id, table_id, updated_at) 
		VALUES 
		(
			$1,
			$2,
			NOW()
		) RETURNING id;`

	queryUpdatePlace string = `
		UPDATE places
		SET 
			bar_id = COALESCE($2, bar_id),
			table_id = COALESCE($3, table_id),
			updated_at = NOW()
		WHERE id = $1;`

	queryDeletePlace string = `UPDATE places SET deleted_at = NOW() WHERE id = $1;`
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
		WHERE id = $2;`

	queryUpdateTeamsLeagueID = `
		UPDATE teams
		SET 
		    league_id = $1
		WHERE id = $2;`

	queryGetTeamLeagueID = `
		SELECT league_id
		FROM teams
		WHERE id = $1;`

	queryDeleteTeamsLeagueID = `
		UPDATE teams 
		SET 
		    league_id = NULL
		WHERE league_id = $1;`

	queryDeleteLeague string = `DELETE FROM leagues WHERE id = $1;`
	// <--

	queryGetGamesYears string = `    
	SELECT DISTINCT EXTRACT(YEAR FROM date) AS year
    FROM games
    ORDER BY year ASC;`
)
