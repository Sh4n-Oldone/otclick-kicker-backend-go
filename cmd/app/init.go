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
	"github.com/heptiolabs/healthcheck"
	"github.com/jackc/pgx/v5/pgxpool"
	goRedis "github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/config"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/endpoint"

	epCity "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/endpoint/city"
	epGame "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/endpoint/game"
	epPlayer "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/endpoint/player"
	srvCity "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/city"
	srvGame "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/game"
	srvPlayer "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/player"

	epLeague "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/endpoint/league"
	srvLeague "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/league"

	epUser "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/endpoint/user"
	srvUser "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/user"

	epMatch "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/endpoint/match"
	srvMatch "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/match"

	epRole "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/endpoint/role"
	srvRole "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/role"

	epTeam "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/endpoint/team"
	srvTeam "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/team"

	epPlace "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/endpoint/place"
	srvPlace "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/place"

	epTable "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/endpoint/table"
	srvTable "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/table"

	epBar "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/endpoint/bar"
	srvBar "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/bar"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/healthchecker"

	tpHTTP "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http"
	tpHTTPBar "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/bar"
	tpHTTPCity "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/city"
	tpHTTPGame "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/game"
	tpHTTPLeague "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/league"
	tpHTTPMatch "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/match"
	tpHTTPPlace "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/place"
	tpHTTPPlayer "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/player"
	tpHTTPRole "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/role"
	tpHTTPTable "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/table"
	tpHTTPTeam "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/team"
	tpHTTPUser "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/user"

	customMiddleware "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/middleware"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/database/postgresql"
	// "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/database/redis"
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

func initRedisConnection(appConfig *config.Configuration) (*goRedis.Client, error) {
	opts, err := goRedis.ParseURL(appConfig.Redis.ConnectionString)
	if err != nil {
		return nil, err
	}
	rds := goRedis.NewClient(opts)

	_, err = rds.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	return rds, nil
}

/*func initCache(config *config.CacheConfig) (cache.ICache, error) {
	return connector.NewCache(config.Type, config.ConnectionString, config.DialTimeout, config.MaxRetries)
}

func initServices(config *config.Configuration, cache cache.ICache, baseLogger zerolog.Logger,
	rwdbOperationer operations.RWDBOperationer, rdbOperationer operations.RDBOperationer) serviceStruct.ServicesEndpoints {
	return serviceStruct.ServicesEndpoints{
		OperationsEP: epOperations.MakeEndpoints(svcOperations.NewOperationsService(
			config, logger.NewComponentLogger(baseLogger, "api_operations"), cache, rwdbOperationer, rdbOperationer)),
	}
}*/

/*func initSystemServiceEndpoint(config *config.Configuration, _ cache.ICache, apiLogger zerolog.Logger) epSystem.Endpoints {
	svcs := svcSystem.NewService(apiLogger, config)
	return epSystem.MakeEndpoints(svcs)
}*/

func initEndpoints(
	appConfig *config.Configuration,
	apiLogger zerolog.Logger,
	rwdbOperationer postgresql.RWDBOperationer,
	rdbOperationer postgresql.RDBOperationer,
) endpoint.ServicesEndpoints {
	citySrv := srvCity.NewService(appConfig, &apiLogger, rwdbOperationer, rdbOperationer)
	userSrv := srvUser.NewService(appConfig, &apiLogger, rwdbOperationer, rdbOperationer)
	matchSrv := srvMatch.NewService(appConfig, &apiLogger, rwdbOperationer, rdbOperationer)
	roleSrv := srvRole.NewService(appConfig, &apiLogger, rwdbOperationer, rdbOperationer)
	playerSrv := srvPlayer.NewService(appConfig, &apiLogger, rwdbOperationer, rdbOperationer)
	teamSrv := srvTeam.NewService(appConfig, &apiLogger, rwdbOperationer, rdbOperationer)
	leagueSrv := srvLeague.NewService(appConfig, &apiLogger, rwdbOperationer, rdbOperationer)
	gameSrv := srvGame.NewService(appConfig, &apiLogger, rwdbOperationer, rdbOperationer)
	tableSrv := srvTable.NewService(appConfig, &apiLogger, rwdbOperationer, rdbOperationer)
	barSrv := srvBar.NewService(appConfig, &apiLogger, rwdbOperationer, rdbOperationer)
	placeSrv := srvPlace.NewService(appConfig, &apiLogger, rwdbOperationer, rdbOperationer)

	return endpoint.ServicesEndpoints{
		CityEP:   epCity.MakeEndpoints(citySrv),
		UserEP:   epUser.MakeEndpoints(userSrv),
		MatchEP:  epMatch.MakeEndpoints(matchSrv),
		RoleEP:   epRole.MakeEndpoints(roleSrv),
		PlayerEP: epPlayer.MakeEndpoints(playerSrv),
		TeamEP:   epTeam.MakeEndpoints(teamSrv),
		LeagueEP: epLeague.MakeEndpoints(leagueSrv),
		GameEP:   epGame.MakeEndpoints(gameSrv),
		TableEP:  epTable.MakeEndpoints(tableSrv),
		BarEP:    epBar.MakeEndpoints(barSrv),
		PlaceEP:  epPlace.MakeEndpoints(placeSrv),
	}
}
