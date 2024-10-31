package redis

import (
	"github.com/redis/go-redis/v9"
)

const (
	// refreshTokenField = "refresh_token"
	// expiresAtField    = "expires_at"
	// passHashField     = "pass_hash"
	// nicknameField     = "nickname"
	// apiKeyField       = "api_key"
)

type redisDB struct {
	dbRedis *redis.Client
}

type Redis interface {
}

func New(connStr *redis.Client) (Redis, error) {
	r := redisDB{dbRedis: connStr}

	return &r, nil
}
