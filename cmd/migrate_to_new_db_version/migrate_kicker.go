package main

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

import (
	"context"
)

var (
	oldConnectionStringDB       = ""
	newConnectionStringDB       = ""
	maxOpenConnection     int32 = 5
	maxIdleConnection     int32 = 1
)

func main() {
	oldDB, err := initDBConnection(oldConnectionStringDB)
	if err != nil {
		println(err)
		return
	}
	defer oldDB.Close()

	newDB, err := initDBConnection(newConnectionStringDB)
	if err != nil {
		println(err)
		return
	}
	defer newDB.Close()

	ctx := context.Background()

	err = migrateCities(ctx, oldDB, newDB)
	if err != nil {
		println(err)
		return
	}
	err = migratePlayers(ctx, oldDB, newDB)
	if err != nil {
		println(err)
		return
	}
	err = migrateTeams(ctx, oldDB, newDB)
	if err != nil {
		println(err)
		return
	}
	err = migrateLeagues(ctx, oldDB, newDB)
	if err != nil {
		println(err)
		return
	}
	/*	err = migrateGames(ctx, oldDB, newDB)
		if err != nil {
			println(err)
			return
		}
		err = migrateMatches(ctx, oldDB, newDB)
		if err != nil {
			println(err)
			return
		}*/

	return
}

func migrateCities(ctx context.Context, oldDB, newDB *pgxpool.Pool) error {
	getCities := `SELECT
		id,
		name, 
		ru
	FROM cities
	WHERE deleted = false;`

	rows, err := oldDB.Query(ctx, getCities)
	if err != nil {
		println(err)
		return err
	}
	defer rows.Close()

	type oldCity struct {
		ID      int64  `db:"id"`
		Name    string `db:"name"`
		Ru      string `db:"ru"`
		Deleted bool   `db:"deleted"`
	}
	var oldCities []oldCity
	for rows.Next() {
		var city oldCity

		err = rows.Scan(&city.ID, &city.Name, &city.Ru)
		if err != nil {
			println(err)
			return err
		}

		oldCities = append(oldCities, city)
	}

	for _, city := range oldCities {
		_, err = newDB.Exec(ctx, `INSERT INTO cities (id, name, ru) VALUES ($1, $2, $3)`, city.ID, city.Name, city.Ru)
		if err != nil {
			println(err)
			return err
		}
	}

	_, err = newDB.Exec(ctx, `DO $$
	DECLARE
		maxid bigint;
	BEGIN
		SELECT MAX(id) INTO maxid FROM cities;
		EXECUTE 'ALTER SEQUENCE cities_id_seq RESTART WITH ' || maxid+1;
	END;
	$$ language plpgsql;`)

	if err != nil {
		println(err)
		return err
	}

	return nil
}

func migratePlayers(ctx context.Context, oldDB, newDB *pgxpool.Pool) error {
	getPlayers := `SELECT
		id,
		name, 
		second_name, 
		last_name, 
		active_player,
		deleted,
		city_id
	FROM players`

	rows, err := oldDB.Query(ctx, getPlayers)
	if err != nil {
		println(err)
		return err
	}
	defer rows.Close()

	type oldPlayer struct {
		ID           int64   `db:"id"`
		Name         string  `db:"name"`
		SecondName   *string `db:"second_name"`
		LastName     string  `db:"last_name"`
		ActivePlayer bool    `db:"active_player"`
		Deleted      bool    `db:"deleted"`
		CityID       int64   `db:"city_id"`
	}
	var oldPlayers []oldPlayer
	for rows.Next() {
		var player oldPlayer

		err = rows.Scan(&player.ID, &player.Name, &player.SecondName, &player.LastName, &player.ActivePlayer, &player.Deleted, &player.CityID)
		if err != nil {
			println(err)
			return err
		}

		oldPlayers = append(oldPlayers, player)
	}

	for _, player := range oldPlayers {
		if player.Deleted {
			_, err = newDB.Exec(ctx, `INSERT INTO players (id, name, second_name, last_name, active_player, city_id, deleted_at) VALUES ($1, $2, $3, $4, $5, $6, NOW())`, player.ID, player.Name, player.SecondName, player.LastName, player.ActivePlayer, player.CityID)
			if err != nil {
				println(err)
				return err
			}
		} else {
			_, err = newDB.Exec(ctx, `INSERT INTO players (id, name, second_name, last_name, active_player, city_id, deleted_at) VALUES ($1, $2, $3, $4, $5, $6, NULL)`, player.ID, player.Name, player.SecondName, player.LastName, player.ActivePlayer, player.CityID)
			if err != nil {
				println(err)
				return err
			}
		}

	}

	_, err = newDB.Exec(ctx, `DO $$
	DECLARE
		maxid bigint;
	BEGIN
		SELECT MAX(id) INTO maxid FROM players;
		EXECUTE 'ALTER SEQUENCE players_id_seq RESTART WITH ' || maxid+1;
	END;
	$$ language plpgsql;`)

	return nil
}

func migrateTeams(ctx context.Context, oldDB, newDB *pgxpool.Pool) error {
	getTeams := `SELECT
		id,
		name, 
		short_name,
		city_id,
		players
	FROM teams`

	rows, err := oldDB.Query(ctx, getTeams)
	if err != nil {
		println(err)
		return err
	}
	defer rows.Close()

	type oldTeam struct {
		ID        int64    `db:"id"`
		Name      string   `db:"name"`
		ShortName *string  `db:"short_name"`
		CityID    int64    `db:"city_id"`
		Players   *[]int64 `db:"players"`
	}
	var oldTeams []oldTeam
	for rows.Next() {
		var team oldTeam

		err = rows.Scan(&team.ID, &team.Name, &team.ShortName, &team.CityID, &team.Players)
		if err != nil {
			println(err)
			return err
		}

		oldTeams = append(oldTeams, team)
	}

	for _, team := range oldTeams {
		_, err = newDB.Exec(ctx, `INSERT INTO teams (id, name, short_name, city_id) VALUES ($1, $2, $3, $4)`, team.ID, team.Name, team.ShortName, team.CityID)
		if err != nil {
			println(err)
			return err
		}

		for _, player := range *team.Players {
			_, err = newDB.Exec(ctx, `INSERT INTO players_teams_links (team_id, player_id) VALUES ($1, $2)`, team.ID, player)
			if err != nil {
				println(err)
				return err
			}
		}
	}

	_, err = newDB.Exec(ctx, `DO $$
	DECLARE
		maxid bigint;
	BEGIN
		SELECT MAX(id) INTO maxid FROM teams;
		EXECUTE 'ALTER SEQUENCE teams_id_seq RESTART WITH ' || maxid+1;
	END;
	$$ language plpgsql;`)

	return nil
}

func migrateLeagues(ctx context.Context, oldDB, newDB *pgxpool.Pool) error {
	getLeagues := `SELECT
		id,
		name, 
		teams,
		city_id
	FROM leagues`

	rows, err := oldDB.Query(ctx, getLeagues)
	if err != nil {
		println(err)
		return err
	}
	defer rows.Close()

	type oldLeague struct {
		ID     int64   `db:"id"`
		Name   string  `db:"name"`
		Teams  []int64 `db:"teams"`
		CityID int64   `db:"city_id"`
	}
	var oldLeagues []oldLeague
	for rows.Next() {
		var league oldLeague

		err = rows.Scan(&league.ID, &league.Name, &league.Teams, &league.CityID)
		if err != nil {
			println(err)
			return err
		}

		oldLeagues = append(oldLeagues, league)
	}

	for _, league := range oldLeagues {
		_, err = newDB.Exec(ctx, `INSERT INTO leagues (id, name, city_id) VALUES ($1, $2, $3)`, league.ID, league.Name, league.CityID)
		if err != nil {
			println(err)
			return err
		}

		for _, team := range league.Teams {
			_, err = newDB.Exec(ctx, `UPDATE teams SET league_id = $1 WHERE id = $2`, league.ID, team)
			if err != nil {
				println(err)
				return err
			}
		}
	}

	_, err = newDB.Exec(ctx, `DO $$
	DECLARE
		maxid bigint;
	BEGIN
		SELECT MAX(id) INTO maxid FROM leagues;
		EXECUTE 'ALTER SEQUENCE leagues_id_seq RESTART WITH ' || maxid+1;
	END;
	$$ language plpgsql;`)

	return nil
}

func initDBConnection(connectionString string) (*pgxpool.Pool, error) {
	dbConnection, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		return nil, err
	}

	dbConnection.MaxConnIdleTime, err = time.ParseDuration("30s")
	if err != nil {
		return nil, err
	}
	dbConnection.MaxConns = maxOpenConnection
	dbConnection.MinConns = maxIdleConnection

	pool, err := pgxpool.NewWithConfig(context.Background(), dbConnection)
	if err != nil {
		return nil, err
	}

	err = pool.Ping(context.Background())
	if err != nil {
		return nil, err
	}

	return pool, nil
}
