package postgresql

const (
	queryGetCityList string = `
		SELECT * 
		FROM cities
		WHERE deleted = false
		ORDER BY id;`

	queryGetCityListWithDeleted string = `
		SELECT * 
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
)
