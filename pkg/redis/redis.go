package redis

import (
	"context"

	"github.com/jphhofmann/redship/package/config"

	"github.com/go-redis/redis/v8"
)

/* Defaults */
var Ctx = context.Background()
var Client *redis.Client

/* Connect to database */
func Connect() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     config.Cfg.Redis.Server,
		Password: config.Cfg.Redis.Password,
		DB:       config.Cfg.Redis.Database,
	})
}
