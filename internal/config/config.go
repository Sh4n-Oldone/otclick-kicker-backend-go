package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"github.com/sethvargo/go-envconfig"
)

const (
	HeaderContentTypeKey  = "Content-Type"
	HeaderContentTypeJSON = "application/json; charset=utf-8"
)

func NewConfig() (*Configuration, error) {
	var envFiles []string
	if _, err := os.Stat(".env"); err == nil {
		log.Println("found .env file, adding it to env config files list")
		envFiles = append(envFiles, ".env")
	}
	if os.Getenv("APP_ENV") != "" {
		appEnvName := fmt.Sprintf(".env.%s", os.Getenv("APP_ENV"))
		if _, err := os.Stat(appEnvName); err == nil {
			log.Println("found", appEnvName, "file, adding it to env config files list")
			envFiles = append(envFiles, appEnvName)
		}
	}
	if len(envFiles) > 0 {
		err := godotenv.Overload(envFiles...)
		if err != nil {
			return nil, errors.Wrapf(err, "error while opening env config: %s", err)
		}
	}
	cfg := &Configuration{}
	ctx := context.Background()

	err := envconfig.Process(ctx, cfg)
	if err != nil {
		return nil, errors.Wrapf(err, "error while parsing env config: %s", err)
	}
	return cfg, nil
}

type (
	// Configuration is basic structure that contains configuration
	Configuration struct {
		Log         LogConfig         `env:",prefix=LOG_"`
		Runtime     RuntimeConfig     `env:",prefix=RUNTIME_"`
		Cache       CacheConfig       `env:",prefix=CACHE_"`
		HTTP        HTTPConfig        `env:",prefix=HTTP_"`
		HealthCheck HealthCheckConfig `env:",prefix=HEALTHCHECK_"`
		Debug       DebugConfig       `env:",prefix=DEBUG_"`
		RWDB        DBConfig          `env:",prefix=RWDB_"`
		RDB         DBConfig          `env:",prefix=RDB_"`
		Version     Version           `env:",prefix=VERSION_"`
		Token       TokenConfig       `env:",prefix=TOKEN_"`
		Secret      Secret            `env:",prefix=SECRET_"`
		Redis       RedisConfig       `env:",prefix=REDIS_"`
		LifeTime    LifeTimeConfig    `env:",prefix=LIFETIME_"`
	}

	LogConfig struct {
		Level             string        `env:"LEVEL,default=info"`
		Batch             bool          `env:"BATCH,default=false"`
		BatchSize         int           `env:"BATCH_SIZE,default=1000"`
		BatchPollInterval time.Duration `env:"BATCH_POLL_INTERVAL,default=5s"`
	}

	RuntimeConfig struct {
		UseCPUs    int `env:"USE_CPUS,default=0"`
		MaxThreads int `env:"MAX_THREADS,default=0"`
	}

	CacheConfig struct {
		Type             string        `env:"TYPE,default=dummy"`
		ConnectionString string        `env:"CONNECTION_STRING"` // ex. redis://<user>:<password>@<host>:<port>/<db_number>
		TTL              time.Duration `env:"TTL,default=60s"`
		DialTimeout      time.Duration `env:"DIAL_TIMEOUT,default=2s"`
		MaxRetries       int           `env:"MAX_RETRIES,default=1"`
	}

	HTTPConfig struct {
		RequestLoggingEnabled      bool          `env:"REQUEST_LOGGING_ENABLED,default=false"`
		ResponseTimeLoggingEnabled bool          `env:"RESPONSE_TIME_LOGGING_ENABLED,default=false"`
		ReadTimeout                time.Duration `env:"READ_TIMEOUT,default=30s"`
		WriteTimeout               time.Duration `env:"WRITE_TIMEOUT,default=30s"`
		IdleTimeout                time.Duration `env:"IDLE_TIMEOUT,default=30s"`
		MaxRequestBodySize         int           `env:"MAX_REQUEST_BODY_SIZE,default=4194304"`
		Network                    string        `env:"NETWORK,default=tcp"`
		Address                    string        `env:"ADDRESS,default=:8081"`
	}

	HealthCheckConfig struct {
		GoroutineThreshold int `env:"GOROUTINE_THRESHOLD,default=20"`
	}

	DebugConfig struct {
		InsecureSkipVerify bool `env:"INSECURE_SKIP_VERIFY"`
	}

	DBConfig struct {
		ConnectionString         string        `env:"CONNECTION_STRING,required"`
		MaxOpenConnection        int32         `env:"MAX_OPEN_CONNECTION,default=25"`
		MaxIdleConnection        int32         `env:"MAX_IDLE_CONNECTION,default=10"`
		MaxIdleConnectionTimeout time.Duration `env:"MAX_IDLE_TIMEOUT,default=300s"`
	}

	Version struct {
		Number string `env:"NUMBER,default=1.0.0"`
		Build  string `env:"BUILD,default=dev"`
	}

	TokenConfig struct {
		AccessTTL  time.Duration `env:"TTL_ACCESS,default=1h"`
		RefreshTTL time.Duration `env:"TTL_REFRESH,default=720h"`
	}

	Secret struct {
		Key  string `env:"KEY, required"`
		Salt string `env:"SALT, required"`
	}

	RedisConfig struct {
		ConnectionString string        `env:"CONNECTION_STRING"` // ex. redis://<user>:<password>@<host>:<port>/<db_number>
		Timeout          time.Duration `env:"TIMEOUT,default=2s"`
	}

	LifeTimeConfig struct {
		Session  time.Duration `env:"SESSION,default=720h"`
		TempUser time.Duration `env:"TEMP_USER,default=24h"`
	}
)
