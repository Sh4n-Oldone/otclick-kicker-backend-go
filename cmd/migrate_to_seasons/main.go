package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

const (
	oldConnectionStringDB       = "" //"postgres://alexthecreator:anarchyintheuk@139.45.228.149:5515/kickerdb?sslmode=disable"
	maxOpenConnection     int32 = 5
	maxIdleConnection     int32 = 1

	defaultYears string = "2024-2025"
)

type city struct {
	id   int64
	name string
}

type game struct {
	leagueId int
	id       int
	cityId   int
	cityName string
}

func main() {
	db, err := initDBConnection(oldConnectionStringDB)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	ctx := context.Background()

	// Начинаем транзакцию
	tx, err := db.Begin(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
			fmt.Println(err)
		}
	}()

	// берем города
	cities, err := getCities(tx, ctx)
	if err != nil {
		fmt.Println(err)
		return
	}

	// группировка игр по городам
	var cityGames [][]game

	// мапа для лиг, ключ - id лиги, значение - id сезона (вдруг у сезона две или более лиг)
	leagueSeason := make(map[int]int)

	// получаю игры по городам
	for _, c := range cities {
		games, err := getCityGames(tx, ctx, c)
		if err != nil {
			fmt.Println(err)
			return
		}
		cityGames = append(cityGames, games)
	}

	// сезон = город
	for _, _games := range cityGames {
		if len(_games) == 0 {
			continue
		}

		firstGame := _games[0]
		id, err := createSeason(tx, ctx, firstGame.cityName)
		if err != nil {
			fmt.Println(err)
			return
		}

		for _, __games := range _games {
			// для каждой лиги ставим id сезона(дубли перезатрутся)
			leagueSeason[__games.leagueId] = id
		}
	}

	fmt.Println(leagueSeason)

	// в лигу надо добавить созданый id сезона
	for leagueId, seasonId := range leagueSeason {
		err := updateLeague(tx, ctx, leagueId, seasonId)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	// Коммитим транзакцию если все успешно
	err = tx.Commit(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("миграция прошла успешно")
}

func getCities(tx pgx.Tx, ctx context.Context) ([]city, error) {
	var cities []city

	queryCites := `select id, name from cities where deleted_at is null`
	rows, err := tx.Query(ctx, queryCites)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var c city
		if err = rows.Scan(&c.id, &c.name); err != nil {
			return nil, err
		}
		cities = append(cities, c)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return cities, nil
}

func getCityGames(tx pgx.Tx, ctx context.Context, city city) ([]game, error) {
	query := `
		select 
			league_id, 
			id,
			city_id 
		from games 
		where games.city_id = $1 and league_id is not null;`

	var games []game

	rows, err := tx.Query(ctx, query, city.id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var g game
		if err = rows.Scan(&g.leagueId, &g.id, &g.cityId); err != nil {
			return nil, fmt.Errorf("error scanning game: %w", err)
		}
		g.cityName = city.name
		games = append(games, g)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return games, nil
}

func createSeason(tx pgx.Tx, ctx context.Context, cityName string) (int, error) {
	var id int
	seasonName := cityName + " " + defaultYears
	fmt.Printf("Creating season: %s\n", seasonName)

	query := `insert into seasons(name, description) values($1, 'migrate from cities/years') returning id`

	err := tx.QueryRow(ctx, query, seasonName).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func updateLeague(tx pgx.Tx, ctx context.Context, leagueId, seasonId int) error {
	query := `update leagues set season_id = $1 where id = $2`

	tag, err := tx.Exec(ctx, query, seasonId, leagueId)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return errors.New("league not found")
	}

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
