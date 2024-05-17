package db

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCtx struct {
	Rds *redis.Client
	Ctx context.Context
}

func (s *RedisCtx) Set(key string, value string) error {
	return s.Rds.Set(s.Ctx, key, value, 0).Err()
}
func (s *RedisCtx) Get(key string) (string, error) {
	return s.Rds.Get(s.Ctx, key).Result()
}

func (s *RedisCtx) GetTtl(key string) (time.Duration, error) {
	return s.Rds.TTL(s.Ctx, key).Result()
}

func (s *RedisCtx) Del(key string) error {
	return s.Rds.Del(s.Ctx, key).Err()
}

func (s *RedisCtx) SetWithExpire(key string, value string, expiration time.Duration) error {
	return s.Rds.Set(s.Ctx, key, value, expiration).Err()
}
