package main

import (
	"context"
	"net"
	"net/http"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/go-playground/validator/v10"
	"github.com/heptiolabs/healthcheck"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/config"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/endpoint"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/healthchecker"
	customMiddleware "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/middleware"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/database/postgresql"

	srvBar "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/bar"
	srvCity "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/city"
	srvGame "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/game"
	srvLeague "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/league"
	srvMatch "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/match"
	srvPlace "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/place"
	srvPlayer "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/player"
	srvRole "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/role"
	srvSeason "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/season"
	srvTable "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/table"
	srvTeam "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/team"
	srvTournament "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/tournament"
	srvUser "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/user"

	epBar "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/endpoint/bar"
	epCity "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/endpoint/city"
	epGame "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/endpoint/game"
	epLeague "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/endpoint/league"
	epMatch "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/endpoint/match"
	epPlace "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/endpoint/place"
	epPlayer "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/endpoint/player"
	epRole "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/endpoint/role"
	epSeason "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/endpoint/season"
	epTable "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/endpoint/table"
	epTeam "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/endpoint/team"
	epTournament "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/endpoint/tournament"
	epUser "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/endpoint/user"

	tpHTTP "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http"
	tpHTTPBar "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/bar"
	tpHTTPCity "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/city"
	tpHTTPGame "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/game"
	tpHTTPLeague "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/league"
	tpHTTPMatch "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/match"
	tpHTTPPlace "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/place"
	tpHTTPPlayer "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/player"
	tpHTTPRole "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/role"
	tpHTTPSeason "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/season"
	tpHTTPTable "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/table"
	tpHTTPTeam "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/team"
	tpHTTPTournament "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/tournament"
	tpHTTPUser "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/user"
)

func initRuntime(cpu, threads int, logger zerolog.Logger) {
	if cpu == 0 {
		cpu = runtime.NumCPU()
		runtime.GOMAXPROCS(runtime.NumCPU())
	} else {
		runtime.GOMAXPROCS(cpu)
	}
	logger.Info().Msgf("set to use %d CPUs", cpu)
	if threads == 0 {
		threads = 10000
	} else {
		debug.SetMaxThreads(threads)
	}
	logger.Info().Msgf("set to use maximum %d threads", threads)
}

func initHTTPRouter(config *config.Configuration) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.NoCache)
	router.Use(middleware.RealIP)
	router.Use(customMiddleware.RequestID)
	router.Use(middleware.Recoverer)
	router.Use(middleware.StripSlashes)
	if config.HTTP.CorsEnabled == false {
		router.Use(cors.Handler(cors.Options{
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   []string{"HEAD", "GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
			AllowedHeaders:   []string{"*"},
			AllowCredentials: true,
		}))
	}

	pongResponse := []byte("pong")
	router.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write(pongResponse)
		return
	})

	return router
}

func initHealthChecker(config *config.Configuration, router *chi.Mux) {
	healthChecker := healthchecker.NewHealthChecker()
	healthChecker.GetHealthChecker().AddLivenessCheck("HTTP Net Listener Started",
		healthcheck.TCPDialCheck(config.HTTP.Address, 50*time.Millisecond))
	healthChecker.GetHealthChecker().AddLivenessCheck("Goroutine Threshold",
		healthcheck.GoroutineCountCheck(25))

	router.Mount("/", healthChecker.Handler())
}

func initKitHTTP(appConfig *config.Configuration, service srvUser.IService, endpoints endpoint.ServicesEndpoints, netLogger zerolog.Logger, listenErr chan error, router *chi.Mux) (*http.Server, net.Listener) {
	var serverOptions []kithttp.ServerOption

	router.Mount("/Kicker.v1.CityService/",
		tpHTTPCity.NewServer(
			endpoints.CityEP,
			serverOptions,
			appConfig,
			service))

	router.Mount("/Kicker.v1.MatchService/",
		tpHTTPMatch.NewServer(
			endpoints.MatchEP,
			serverOptions,
			appConfig,
			service))

	router.Mount("/Kicker.v1.PlayerService/",
		tpHTTPPlayer.NewServer(
			endpoints.PlayerEP,
			serverOptions,
			appConfig,
			service))

	router.Mount("/Kicker.v1.TeamService/",
		tpHTTPTeam.NewServer(
			endpoints.TeamEP,
			serverOptions,
			appConfig,
			service))

	router.Mount("/Kicker.v1.UserService/",
		tpHTTPUser.NewServer(
			endpoints.UserEP,
			serverOptions,
			appConfig,
			service))

	router.Mount("/Kicker.v1.RoleService/",
		tpHTTPRole.NewServer(
			endpoints.RoleEP,
			serverOptions,
			appConfig,
			service))

	router.Mount("/Kicker.v1.LeagueService/",
		tpHTTPLeague.NewServer(
			endpoints.LeagueEP,
			serverOptions,
			appConfig,
			service))

	router.Mount("/Kicker.v1.GameService/",
		tpHTTPGame.NewServer(
			endpoints.GameEP,
			serverOptions,
			appConfig,
			service))

	router.Mount("/Kicker.v1.TableService/",
		tpHTTPTable.NewServer(
			endpoints.TableEP,
			serverOptions,
			appConfig,
			service))

	router.Mount("/Kicker.v1.BarService/",
		tpHTTPBar.NewServer(
			endpoints.BarEP,
			serverOptions,
			appConfig,
			service))

	router.Mount("/Kicker.v1.PlaceService/",
		tpHTTPPlace.NewServer(
			endpoints.PlaceEP,
			serverOptions,
			appConfig,
			service))

	router.Mount("/Kicker.v1.SeasonService/",
		tpHTTPSeason.NewServer(
			endpoints.SeasonEP,
			serverOptions,
			appConfig,
			service))

	router.Mount("/Kicker.v1.TournamentService/",
		tpHTTPTournament.NewServer(
			endpoints.TournamentEP,
			serverOptions,
			appConfig,
			service))

	if webDebugEnabled {
		router.Mount("/dbg", ProfilerHandler())
	}

	httpServer := &http.Server{
		Handler:      router,
		TLSConfig:    nil,
		ReadTimeout:  appConfig.HTTP.ReadTimeout,
		WriteTimeout: appConfig.HTTP.WriteTimeout,
		IdleTimeout:  appConfig.HTTP.IdleTimeout,
	}

	l, err := net.Listen(appConfig.HTTP.Network, appConfig.HTTP.Address)
	if err != nil {
		netLogger.Fatal().Err(err).Msg("failed to init net.Listen for http")
	} else {
		netLogger.Info().Msg("successful net.Listen for http init")
	}

	go tpHTTP.RunHTTPServer(httpServer, l, netLogger, listenErr)
	time.Sleep(10 * time.Millisecond)
	return httpServer, l
}

func initDBConnection(dbConfig *config.DBConfig) (*pgxpool.Pool, error) {
	postgresConfig, err := pgxpool.ParseConfig(dbConfig.ConnectionString)
	if err != nil {
		return nil, err
	}

	postgresConfig.MaxConnIdleTime = dbConfig.MaxIdleConnectionTimeout
	postgresConfig.MaxConns = dbConfig.MaxOpenConnection
	postgresConfig.MinConns = dbConfig.MaxIdleConnection

	pool, err := pgxpool.NewWithConfig(context.Background(), postgresConfig)
	if err != nil {
		return nil, err
	}

	err = pool.Ping(context.Background())
	if err != nil {
		return nil, err
	}

	return pool, nil
}

func initEndpoints(
	appConfig *config.Configuration,
	apiLogger zerolog.Logger,
	validator *validator.Validate,
	rwdbOperationer postgresql.RWDBOperationer,
	rdbOperationer postgresql.RDBOperationer,
) endpoint.ServicesEndpoints {
	citySrv := srvCity.NewService(appConfig, &apiLogger, rwdbOperationer, rdbOperationer)
	userSrv := srvUser.NewService(appConfig, &apiLogger, validator, rwdbOperationer, rdbOperationer)
	matchSrv := srvMatch.NewService(appConfig, &apiLogger, rwdbOperationer, rdbOperationer)
	roleSrv := srvRole.NewService(appConfig, &apiLogger, rwdbOperationer, rdbOperationer)
	playerSrv := srvPlayer.NewService(appConfig, &apiLogger, validator, rwdbOperationer, rdbOperationer)
	teamSrv := srvTeam.NewService(appConfig, &apiLogger, validator, rwdbOperationer, rdbOperationer, playerSrv)
	leagueSrv := srvLeague.NewService(appConfig, &apiLogger, rwdbOperationer, rdbOperationer)
	gameSrv := srvGame.NewService(appConfig, &apiLogger, validator, rwdbOperationer, rdbOperationer)
	tableSrv := srvTable.NewService(appConfig, &apiLogger, rwdbOperationer, rdbOperationer)
	barSrv := srvBar.NewService(appConfig, &apiLogger, rwdbOperationer, rdbOperationer)
	placeSrv := srvPlace.NewService(appConfig, &apiLogger, validator, rwdbOperationer, rdbOperationer)
	seasonSrv := srvSeason.NewService(appConfig, &apiLogger, rwdbOperationer, rdbOperationer)
	tournamentSrv := srvTournament.NewService(appConfig, &apiLogger, validator, rwdbOperationer, rdbOperationer)

	return endpoint.ServicesEndpoints{
		CityEP:       epCity.MakeEndpoints(citySrv),
		UserEP:       epUser.MakeEndpoints(userSrv),
		MatchEP:      epMatch.MakeEndpoints(matchSrv),
		RoleEP:       epRole.MakeEndpoints(roleSrv),
		PlayerEP:     epPlayer.MakeEndpoints(playerSrv),
		TeamEP:       epTeam.MakeEndpoints(teamSrv),
		LeagueEP:     epLeague.MakeEndpoints(leagueSrv),
		GameEP:       epGame.MakeEndpoints(gameSrv),
		TableEP:      epTable.MakeEndpoints(tableSrv),
		BarEP:        epBar.MakeEndpoints(barSrv),
		PlaceEP:      epPlace.MakeEndpoints(placeSrv),
		SeasonEP:     epSeason.MakeEndpoints(seasonSrv),
		TournamentEP: epTournament.MakeEndpoints(tournamentSrv),
	}
}
