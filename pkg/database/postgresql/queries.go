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
)
