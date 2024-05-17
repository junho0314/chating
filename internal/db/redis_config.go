package db

import (
	"chating_service/internal/config"
	"context"

	redis "github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

var rds *redis.Client

func InitRedisConnection(appConfig *config.AppConfig) {
	// Initialize redis connection

	rds = redis.NewClient(&redis.Options{
		Addr:     appConfig.Redis.Host + ":" + appConfig.Redis.Port,
		Password: appConfig.Redis.Password, // no password set
		DB:       0,                        // use default DB
	})

	_, err := rds.Ping(context.Background()).Result()
	if err != nil {
		log.Error().Msg("InitRedisConnection:: error connecting to the redis. " + err.Error())
		panic(err)

	}

	err = rds.Set(context.Background(), "connect", "success", 0).Err()
	if err != nil {
		log.Error().Msg("InitRedisConnection:: error setting key. " + err.Error())
		panic(err)
	}
}

func GetRedisConnection(ctx context.Context) RedisCtx {
	return RedisCtx{
		Rds: rds,
		Ctx: ctx,
	}
}
