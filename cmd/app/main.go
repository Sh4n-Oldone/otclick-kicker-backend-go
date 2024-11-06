package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/config"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/user"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/database/postgresql"
	// "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/database/redis"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/logger"
)

func main() {
	// config
	appConfig, err := config.NewConfig() //todo переписать конфиг - должен быть более читаемым
	if err != nil {
		log.Fatalln(err)
	}

	// profiler
	profiler := initDebugger()
	defer func() {
		stopDebugger(profiler)
	}()
	if webDebugEnabled {
		appConfig.Log.Level = logger.Debug
	}

	// loggers
	var baseLogger zerolog.Logger
	var loggerCloser io.WriteCloser
	if appConfig.Log.Batch {
		baseLogger, loggerCloser, err = logger.NewDiodeLogger(os.Stdout, appConfig.Log.Level, appConfig.Log.BatchSize, appConfig.Log.BatchPollInterval)
	} else {
		baseLogger, loggerCloser, err = logger.NewLogger(os.Stdout, appConfig.Log.Level)
	}
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		if loggerCloser != nil {
			err = loggerCloser.Close()
			if err != nil {
				log.Fatalf("error acquired while closing log writer: %+v", err)
			}
		}
	}()
	baseLogger = baseLogger.With().
		Str("app_version", appConfig.Version.Number).
		Str("app_build", appConfig.Version.Build).
		CallerWithSkipFrameCount(2).
		Logger()
	apiLogger := logger.NewComponentLogger(baseLogger, "api")
	coreLogger := logger.NewComponentLogger(baseLogger, "core")
	netLogger := logger.NewComponentLogger(baseLogger, "net")

	defer func() {
		coreLogger.Info().Msg("application stopped")
	}()

	coreLogger.Info().Msg("system initialization started")

	initRuntime(appConfig.Runtime.UseCPUs, appConfig.Runtime.MaxThreads, coreLogger)

	listenErr := make(chan error, 1)

	rwdb, err := initDBConnection(&appConfig.RWDB)
	if err != nil {
		coreLogger.Fatal().Err(err).Msg("failed to establish a connection with the Read/Write database")
	} else {
		coreLogger.Info().Msg("successful connection with the Read/Write database")
	}
	defer rwdb.Close()

	rdb, err := initDBConnection(&appConfig.RDB)
	if err != nil {
		coreLogger.Fatal().Err(err).Msg("failed to establish a connection with the Read-only database")
	} else {
		coreLogger.Info().Msg("successful connection with the Read-only database")
	}
	defer rdb.Close()

	rwdbOperationer, rdbOperationer := postgresql.NewOperationer(rwdb, rdb)

	//rds, err := initRedisConnection(appConfig)
	//if err != nil {
	//	coreLogger.Fatal().Err(err).Msg("failed to establish a connection with the redis")
	//} else {
	//	coreLogger.Info().Msg("successful connection with the redis")
	//}
	//defer func(rds *goRedis.Client) {
	//	err = rds.Close()
	//	if err != nil {
	//		coreLogger.Error().Msg("failed to close the redis connection")
	//	}
	//}(rds)

	// redisDB, err := redis.New(rds)

	userService := user.NewService(appConfig, &apiLogger, rwdbOperationer, rdbOperationer)

	serviceEndpoints := initEndpoints(appConfig, apiLogger, rwdbOperationer, rdbOperationer)
	chiRouter := initHTTPRouter(appConfig)

	initHealthChecker(appConfig, chiRouter)

	httpServer, httpListener := initKitHTTP(appConfig, userService, serviceEndpoints, netLogger, listenErr, chiRouter)
	defer func() {
		err = httpListener.Close()
		if err != nil {
			// we don't really need it because we already closed it by cmux.Close()
			// netLogger.Warn().Err(err).Msgf("failed to close net.Listen %+w - %+v", err, err)
		}
	}()

	runApp(httpServer, coreLogger, listenErr)
}

func runApp(httpServer *http.Server, coreLogger zerolog.Logger, listenErr chan error) {

	var shutdownCh = make(chan os.Signal, 1)
	signal.Notify(shutdownCh, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	var err error
	var runningApp = true

	for runningApp {
		select {
		// handle error channel
		case err = <-listenErr:
			if err != nil {
				coreLogger.Error().Err(err).Msg("received listener error")
				shutdownCh <- os.Kill
			}
		// handle os system signal
		case sig := <-shutdownCh:
			coreLogger.Info().Msgf("shutdown signal received: %s", sig.String())
			ctxTimeout, timeoutCancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
			err = httpServer.Shutdown(ctxTimeout) // may return ErrServerClosed
			defer timeoutCancelFunc()
			if err != nil {
				coreLogger.Error().Err(err).Msg("received http Shutdown error")
			}
			coreLogger.Info().Msg("server loop stopped")
			runningApp = false
		}
	}
}
