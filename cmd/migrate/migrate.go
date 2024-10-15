package main

import (
	"embed"
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/pkg/errors"

	"github.com/urfave/cli/v2"

	"node71.otclick.ru/backend/template/migrations"
)

var (
	m               *migrate.Migrate
	pgMigrationsDir string
	pgConnString    string
)

func main() {
	pgMigrationsDir = "database/postgres"
	pgConnString = os.Getenv("RWDB_CONNECTION_STRING")

	fmt.Println("Found migration files")
	fmt.Println("Postgres:")

	fileList, err := getAllFilenames(&migrations.Database, pgMigrationsDir)
	if err != nil {
		log.Fatalln(err)
	}
	for i := range fileList {
		fmt.Println(fileList[i])
	}

	app := cli.App{
		Name:      "migrate_tool",
		HelpName:  "migrate",
		Usage:     "tool for database migration management",
		UsageText: "migrate - tool for database migration management",
		ArgsUsage: "",
		Version:   "v0.0.1",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "conn", Aliases: []string{"c"}},
		},
		Commands: cli.Commands{
			&cli.Command{
				Name: "up",
				Action: func(context *cli.Context) error {
					err := getMigrate(context)
					if err != nil {
						return err
					}
					steps := context.Int("step")
					if steps == 0 {
						err := migrateUp(m, nil)
						if err != nil {
							return err
						}
					} else {
						err := migrateUp(m, &steps)
						if err != nil {
							return err
						}
					}
					return nil
				},
				Description: "Up command migrate up database version by all available migration if step flag not used and by \"n\" steps if step flag used. As example - \"migrate up -steps 2\"",
				Flags: []cli.Flag{
					&cli.IntFlag{Name: "step", Aliases: []string{"s"}},
				},
			},
			&cli.Command{
				Name: "down",
				Action: func(context *cli.Context) error {
					err := getMigrate(context)
					if err != nil {
						return err
					}
					steps := context.Int("step")
					steps *= -1
					if steps == 0 {
						err := migrateDown(m, nil)
						if err != nil {
							return err
						}
					} else {
						err := migrateDown(m, &steps)
						if err != nil {
							return err
						}
					}
					return nil
				},
				Description: "Down command migrate down database version by all available migration if step flag not used and by \"n\" steps if step flag used. As example - \"migrate down -steps 2\"",
				Flags: []cli.Flag{
					&cli.IntFlag{Name: "step", Aliases: []string{"s"}},
				},
			},
			&cli.Command{
				Name: "upto",
				Action: func(context *cli.Context) error {
					err := getMigrate(context)
					if err != nil {
						return err
					}
					oldVersion, _, _ := m.Version()
					newVersion := context.Int("version")
					if newVersion == 0 {
						migVer := os.Getenv("MIGRATION_VERSION")
						migVerInt, err := strconv.ParseInt(migVer, 10, 64)
						if err != nil {
							return err
						}
						newVersion = int(migVerInt)
					}
					if uint(newVersion) < oldVersion {
						return errors.New("new migration version must be greater or equal old one")
					}
					steps := 0
					for i, file := range fileList {
						if strings.Contains(file, strconv.Itoa(int(oldVersion))) {
							steps = i
						}
						if strings.Contains(file, strconv.Itoa(newVersion)) {
							steps = i - steps
						}
					}
					if steps <= 0 {
						fmt.Println("Nothing to change")
						return nil
					}
					err = migrateUp(m, &steps)
					if err != nil {
						return err
					}
					return nil
				},
				Flags: []cli.Flag{
					&cli.IntFlag{Name: "version", Aliases: []string{"v"}},
				},
			},
			&cli.Command{
				Name: "force",
				Action: func(context *cli.Context) error {
					err := getMigrate(context)
					if err != nil {
						return err
					}
					version := context.Int("version")
					err = migrateForce(m, version)
					if err != nil {
						return err
					}
					return nil
				},
				Description: "Force command force database version by \"version\" flag value",
				Flags: []cli.Flag{
					&cli.IntFlag{Name: "version", Aliases: []string{"v"}, Required: true},
				},
			},
		},
		EnableBashCompletion: true,
		CommandNotFound: func(cCtx *cli.Context, command string) {
			fmt.Fprintf(cCtx.App.Writer, "Thar be no %q here.\n", command)
		},
		Compiled: time.Now(),
		Authors: []*cli.Author{
			{Name: "Aleksey Shepelev", Email: "a.shepelev@otclick.ru"},
		},
	}

	err = app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

	ver, dirt, _ := m.Version()
	fmt.Printf("New DB Version: %v Dirty: %v\n", ver, dirt)
	os.Exit(0)
}

func getMigrate(context *cli.Context) (err error) {
	if conn := context.String("conn"); conn != "" {
		pgConnString = conn
	}

	m, err = migrate.New("file://migrations/"+pgMigrationsDir, pgConnString)
	if err != nil {
		return err
	}

	ver, dirt, _ := m.Version()
	fmt.Printf("DB Version: %v Dirty: %v\n", ver, dirt)
	return nil
}

func getAllFilenames(fs *embed.FS, dir string) (out []string, err error) {
	if len(dir) == 0 {
		dir = "."
	}
	entries, err := fs.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		fp := path.Join(dir, entry.Name())
		if entry.IsDir() {
			res, err := getAllFilenames(fs, fp)
			if err != nil {
				return nil, err
			}
			out = append(out, res...)
			continue
		}
		out = append(out, fp)
	}
	return
}

func migrateUp(m *migrate.Migrate, step *int) error {
	log.Println("Migrate up started")
	defer log.Println("Migrate up finished")

	if step == nil {
		err := m.Up()
		if err != nil {
			return err
		}
	} else {
		err := m.Steps(*step)
		if err != nil {
			return err
		}
	}

	return nil
}

func migrateDown(m *migrate.Migrate, step *int) error {
	log.Println("Migrate down started")
	defer log.Println("Migrate down finished")

	if step == nil {
		err := m.Down()
		if err != nil {
			return err
		}
	} else {
		err := m.Steps(*step)
		if err != nil {
			return err
		}
	}

	return nil
}

func migrateForce(m *migrate.Migrate, version int) error {
	log.Println("Migrate down started")
	defer log.Println("Migrate down finished")

	err := m.Force(version)
	if err != nil {
		return err
	}

	return nil
}
