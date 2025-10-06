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

	queryCreateUserWithTeamID string = `
		INSERT INTO users (email, password, role_id, team_id)
		VALUES
		(
			$1,
			$2,
			$3,
			$4
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
		score_team2,
		updated_at
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
		$10,
		NOW()
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
			score_team2 = COALESCE($10, score_team2),
			updated_at = NOW()
		WHERE id = $11;`

	queryUpdateMatchWithRatings string = `
		UPDATE matches 
		SET
			date = COALESCE($2, date),
			game_id = COALESCE($3, game_id),
			team1_id = COALESCE($4, team1_id),
			team2_id = COALESCE($5, team2_id),
			player1_team1_id = COALESCE($6, player1_team1_id),
			player2_team1_id = COALESCE($7, player2_team1_id),
			player1_team2_id = COALESCE($8, player1_team2_id),
			player2_team2_id = COALESCE($9, player2_team2_id),
			score_team1 = COALESCE($10, score_team1),
			score_team2 = COALESCE($11, score_team2),
			player1_team1_rate_before = $12,
			player1_team2_rate_before = $13,
			player2_team1_rate_before = $14,
			player2_team2_rate_before = $15,
			player1_team1_rate_after = $16,
			player1_team2_rate_after = $17,
			player2_team1_rate_after = $18,
			player2_team2_rate_after = $19,
			updated_at = NOW()
		WHERE id = $1;`

	queryDeleteMatch string = `DELETE FROM matches WHERE id = $1`

	queryInsertMatchesToPlayedGame string = `
		INSERT INTO public.matches (date,
		                            game_id,
		                            team1_id,
		                            team2_id,
		                            player1_team1_id,
		                            player2_team1_id,
		                            player1_team2_id,
		                            player2_team2_id,
		                            score_team1,
		                            score_team2,
		                            player1_team1_rate_before,
		                            player1_team2_rate_before,
		                            player2_team1_rate_before,
		                            player2_team2_rate_before,
		                            player1_team1_rate_after,
		                            player1_team2_rate_after,
		                            player2_team1_rate_after,
		                            player2_team2_rate_after,
		                            sort)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
		RETURNING id;`

	queryGetMatchListByGameID string = `
		SELECT id,
			date,
			game_id,
			team1_id,
			team2_id,
			player1_team1_id,
			player2_team1_id,
			player1_team2_id,
			player2_team2_id,
			score_team1,
			score_team2,
			player1_team1_rate_before,
			player1_team2_rate_before,
			player2_team1_rate_before,
			player2_team2_rate_before,
			player1_team1_rate_after,
			player1_team2_rate_after,
			player2_team1_rate_after,
			player2_team2_rate_after 
		FROM matches 
		WHERE game_id = $1;`

	queryUpdateMatchesToUpdateGame string = `
		UPDATE public.matches m
		SET 
			date = $2,
			team1_id = $3,
			team2_id = $4,
			player1_team1_id = $5,
			player2_team1_id = $6,
			player1_team2_id = $7,
			player2_team2_id = $8,
			score_team1 = $9,
			score_team2 = $10,
			player1_team1_rate_before = $11,
			player1_team2_rate_before = $12,
			player2_team1_rate_before = $13,
			player2_team2_rate_before = $14,
			player1_team1_rate_after = $15,
			player1_team2_rate_after = $16,
			player2_team1_rate_after = $17,
			player2_team2_rate_after = $18,
			sort = $19,
			updated_at = NOW()
		WHERE id = $1 AND game_id = $20;`

	queryInsertMatchesToUpdateGame string = `
		INSERT INTO public.matches 
		    (
		     date,
		     game_id,
		     team1_id,
		     team2_id,
		     player1_team1_id,
		     player2_team1_id,
		     player1_team2_id,
		     player2_team2_id,
		     score_team1,
		     score_team2,
   			 player1_team1_rate_before,
			 player1_team2_rate_before,
			 player2_team1_rate_before,
			 player2_team2_rate_before,
			 player1_team1_rate_after,
			 player1_team2_rate_after,
			 player2_team1_rate_after,
			 player2_team2_rate_after,
		     sort,
		     updated_at
		     )
		VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, NOW());`

	queryDeleteMatchesToUpdateGame string = `
		DELETE FROM public.matches m
    	WHERE m.game_id = $2 AND m.id NOT IN (SELECT unnest($1::int[]));`
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
		UPDATE bars
		SET 
		    name = COALESCE($2, name),
    		description = COALESCE($3, description),
			updated_at = NOW()
		WHERE id = $1;`

	queryDeleteBar string = `UPDATE bars SET deleted_at = NOW() WHERE id = $1;`
	// <--

	// Place queries -->
	queryGetPlaceList string = `
		SELECT p.id, p.bar_id, b.name, p.table_id, t.name, p.updated_at, p.deleted_at 
		FROM places as p
		JOIN tables as t ON p.table_id = t.id
		JOIN bars as b ON p.bar_id = b.id
		WHERE p.deleted_at IS NULL
		ORDER BY p.bar_id, p.table_id, p.id;`

	queryGetPlaceListWithDeleted string = `
		SELECT p.id, p.bar_id, b.name, p.table_id, t.name, p.updated_at, p.deleted_at 
		FROM places as p
		JOIN tables as t ON p.table_id = t.id
		JOIN bars as b ON p.bar_id = b.id
		ORDER BY p.bar_id, p.table_id, p.id;`

	queryGetPlaceListByBarID string = `
		SELECT p.id, p.bar_id, b.name, p.table_id, t.name, p.updated_at, p.deleted_at 
		FROM places as p
		JOIN tables as t ON p.table_id = t.id
		JOIN bars as b ON p.bar_id = b.id
		WHERE p.bar_id = $1 AND p.deleted_at IS NULL
		ORDER BY p.bar_id, p.table_id, p.id;`

	queryGetPlaceListByBarIDWithDeleted string = `
		SELECT p.id, p.bar_id, b.name, p.table_id, t.name, p.updated_at, p.deleted_at 
		FROM places as p
		JOIN tables as t ON p.table_id = t.id
		JOIN bars as b ON p.bar_id = b.id
		WHERE p.bar_id = $1
		ORDER BY p.bar_id, p.table_id, p.id;`

	queryGetPlaceListByTableID string = `
		SELECT p.id, p.bar_id, b.name, p.table_id, t.name, p.updated_at, p.deleted_at 
		FROM places as p
		JOIN tables as t ON p.table_id = t.id
		JOIN bars as b ON p.bar_id = b.id
		WHERE p.table_id = $1 AND p.deleted_at IS NULL
		ORDER BY p.bar_id, p.table_id, p.id;`

	queryGetPlaceListByTableIDWithDeleted string = `
		SELECT p.id, p.bar_id, b.name, p.table_id, t.name, p.updated_at, p.deleted_at 
		FROM places as p
		JOIN tables as t ON p.table_id = t.id
		JOIN bars as b ON p.bar_id = b.id
		WHERE p.table_id = $1
		ORDER BY p.bar_id, p.table_id, p.id;`

	queryGetPlaceListByBarIDByTableID string = `
		SELECT p.id, p.bar_id, b.name, p.table_id, t.name, p.updated_at, p.deleted_at 
		FROM places as p
		JOIN tables as t ON p.table_id = t.id
		JOIN bars as b ON p.bar_id = b.id
		WHERE p.bar_id = $1 AND p.table_id = $2 AND p.deleted_at IS NULL
		ORDER BY p.bar_id, p.table_id, p.id;`

	queryGetPlaceListByBarIDByTableIDWithDeleted string = `
		SELECT p.id, p.bar_id, b.name, p.table_id, t.name, p.updated_at, p.deleted_at 
		FROM places as p
		JOIN tables as t ON p.table_id = t.id
		JOIN bars as b ON p.bar_id = b.id
		WHERE p.bar_id = $1 AND p.table_id = $2
		ORDER BY p.bar_id, p.table_id, p.id;`

	queryGetPlaceListByCityID string = `
		SELECT p.id, p.bar_id, b.name, p.table_id, t.name, p.updated_at, p.deleted_at 
		FROM places as p
		JOIN bars as b ON p.bar_id = b.id
		JOIN tables as t ON p.table_id = t.id
		WHERE b.city_id = $1 AND p.deleted_at IS NULL
		ORDER BY p.bar_id, p.table_id, p.id;`

	queryGetPlaceListByCityIDWithDeleted string = `
		SELECT p.id, p.bar_id, b.name, p.table_id, t.name, p.updated_at, p.deleted_at 
		FROM places as p
		JOIN bars as b ON p.bar_id = b.id
		JOIN tables as t ON p.table_id = t.id
		WHERE b.city_id = $1
		ORDER BY p.bar_id, p.table_id, p.id;`

	queryGetPlaceListByCityIDByTableID string = `
		SELECT p.id, p.bar_id, b.name, p.table_id, t.name, p.updated_at, p.deleted_at 
		FROM places as p
		JOIN bars as b ON p.bar_id = b.id
		JOIN tables as t ON p.table_id = t.id
		WHERE b.city_id = $1 AND p.table_id = $2 AND p.deleted_at IS NULL
		ORDER BY p.bar_id, p.table_id, p.id;`

	queryGetPlaceListByCityIDByTableIDWithDeleted string = `
		SELECT p.id, p.bar_id, b.name, p.table_id, t.name, p.updated_at, p.deleted_at 
		FROM places as p
		JOIN bars as b ON p.bar_id = b.id
		JOIN tables as t ON p.table_id = t.id
		WHERE b.city_id = $1 AND p.table_id = $2
		ORDER BY p.bar_id, p.table_id, p.id;`

	queryGetPlaceByID string = `
		SELECT id, bar_id, table_id, updated_at, deleted_at 
		FROM places
		WHERE id = $1 AND deleted_at IS NULL;`

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
		WHERE id = $1 AND deleted_at IS NULL;`

	queryDeletePlace string = `UPDATE places SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL;`
	// <--

	// League queries -->
	queryGetLeagueList string = `
		SELECT id, name, city_id 
		FROM leagues
		WHERE city_id = $1
		ORDER BY id;`

	queryCreateLeague string = `
		INSERT INTO leagues (name, city_id, season_id) 
		VALUES 
		(
			$1,
			$2,
			$3
		) RETURNING id;`

	queryUpdateLeague string = `
		UPDATE leagues
		SET 
		    name = COALESCE($2, name),
		    season_id = COALESCE($3, season_id)
		WHERE id = $1;`

	queryUpdateTeamsLeagueID = `
		INSERT INTO teams_leagues_links (team_id, league_id)
		VALUES ($1, $2);`

	queryDeleteTeamsLeagueID = `
		DELETE FROM teams_leagues_links
		WHERE league_id = $1;`

	queryDeleteLeague string = `DELETE FROM leagues WHERE id = $1;`
	// <--

	// Rating queries -->
	queryGetRatingList                    = `SELECT player_id, league_id, value FROM rating;`
	queryGetRatingListByPlayerID          = `SELECT player_id, league_id, value FROM rating WHERE player_id = $1;`
	queryGetRatingListByLeagueID          = `SELECT player_id, league_id, value FROM rating WHERE league_id = $1;`
	queryGetRatingByPlayerIDAndByLeagueID = `SELECT value FROM rating WHERE player_id = $1 AND league_id = $2;`
	queryCreateRating                     = `INSERT INTO rating (player_id, league_id, value) VALUES ($1, $2, $3);`
	queryCreateRatingInsertIgnore         = `INSERT INTO rating (player_id, league_id, value) VALUES ($1, $2, $3) ON CONFLICT (player_id, league_id) DO UPDATE SET value = $3;`
	queryUpdateRating                     = `UPDATE rating SET value = $3, updated_at = NOW() WHERE player_id = $1 AND league_id = $2;`

	queryDeleteRatingByLeagueID = `DELETE FROM rating WHERE league_id = $1;`
	// <--

	// Games queries -->
	queryGetGamesYears string = `    
	SELECT DISTINCT EXTRACT(YEAR FROM date) AS year
    FROM games
	WHERE date IS NOT NULL
    ORDER BY year ASC;`

	queryInsertGame string = `
		INSERT INTO public.games (city_id, place_id, league_id, date, team1_id, team2_id, tech_loose_team_id, is_tiebreak)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id;`

	queryUpdateGame string = `
		UPDATE public.games
		SET 
			date = $2,
			place_id = $3,
			league_id = $4,
			team1_id = $5,
			team2_id = $6,
			tech_loose_team_id = $7
		WHERE id = $1;`

	queryCreateSeason string = `
		INSERT INTO seasons (name, description) 
		VALUES 
		(
			$1,
			$2
		) RETURNING id;`

	queryUpdateSeason string = `
		UPDATE seasons
		SET
		    name = COALESCE($2, name),
    		description = COALESCE($3, description)
		WHERE id = $1;`

	queryDeleteSeason string = `DELETE FROM seasons WHERE id = $1`

	queryDeleteLeagueSeasonId string = `
		UPDATE leagues
		SET
		    season_id = null
		WHERE season_id = $1;`

	queryGetSeasonList string = `
		SELECT DISTINCT s.id, s.name, s.description 
		FROM seasons s
		WHERE 
			($1::INT IS NULL OR s.id = $1::INT)
			AND (
				$2::INT IS NULL 
				OR EXISTS (
					SELECT 1 FROM tournaments t 
					WHERE t.season_id = s.id AND t.city_id = $2::INT
				)
				OR EXISTS (
					SELECT 1 FROM leagues l 
					WHERE l.season_id = s.id AND l.city_id = $2::INT
				)
			);`

	queryDeleteFutureGame string = `DELETE FROM public.games WHERE id = $1 AND date > NOW();`
)
