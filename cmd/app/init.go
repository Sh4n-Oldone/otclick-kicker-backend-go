package main

import (
	"context"
	"net"
	"net/http"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/bufbuild/protovalidate-go"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/heptiolabs/healthcheck"
	"github.com/jackc/pgx/v5/pgxpool"
	goRedis "github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/config"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/endpoint"

	epCity "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/endpoint/city"
	epPlayer "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/endpoint/player"
	srvCity "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/city"
	srvPlayer "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/player"

	epMatch "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/endpoint/match"
	srvMatch "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/match"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/healthchecker"
	tpHTTP "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http"
	tpHTTPCity "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/city"
	tpHTTPMatch "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/match"
	customMiddleware "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/middleware"
	tpHTTPPlayer "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/player"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/database/postgresql"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/database/redis"
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

func initHTTPRouter(_ *config.Configuration) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.NoCache)
	router.Use(middleware.RealIP)
	router.Use(customMiddleware.RequestID)
	router.Use(middleware.Recoverer)
	router.Use(middleware.StripSlashes)

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

func initKitHTTP(appConfig *config.Configuration, endpoints endpoint.ServicesEndpoints, netLogger zerolog.Logger, listenErr chan error, router *chi.Mux) (*http.Server, net.Listener) {
	var serverOptions []kithttp.ServerOption

	router.Mount("/Kicker.v1.CityService/",
		tpHTTPCity.NewServer(
			endpoints.CityEP,
			serverOptions))

	router.Mount("/Kicker.v1.MatchService/",
		tpHTTPMatch.NewServer(
			endpoints.MatchEP,
			serverOptions))

	router.Mount("/Kicker.v1.PlayerService/",
		tpHTTPPlayer.NewServer(
			endpoints.PlayerEP,
			serverOptions))

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
	validator *protovalidate.Validator,
	rwdbOperationer postgresql.RWDBOperationer,
	rdbOperationer postgresql.RDBOperationer,
	redisDB redis.Redis,
) endpoint.ServicesEndpoints {
	citySrv := srvCity.NewService(appConfig, &apiLogger, validator, rwdbOperationer, rdbOperationer)
	matchSrv := srvMatch.NewService(appConfig, &apiLogger, validator, rwdbOperationer, rdbOperationer)
	playerSrv := srvPlayer.NewService(appConfig, &apiLogger, rwdbOperationer, rdbOperationer)

	return endpoint.ServicesEndpoints{
		CityEP:   epCity.MakeEndpoints(citySrv),
		MatchEP:  epMatch.MakeEndpoints(matchSrv),
		PlayerEP: epPlayer.MakeEndpoints(playerSrv),
	}
}
