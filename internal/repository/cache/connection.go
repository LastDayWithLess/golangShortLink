package cache

import (
	"context"
	"fmt"
	"short_link/config"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	rdb *redis.Client
}

func NewRedisConnect(cfg config.RedisConfig) (*RedisClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := rdb.Ping(ctx).Result()

	if err != nil {
		rdb.Close()
		return nil, fmt.Errorf("Error connecting to Redis: %s", err)
	}

	return &RedisClient{rdb: rdb}, nil
}

func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return r.rdb.Get(ctx, key).Result()
}

func (r *RedisClient) Set(ctx context.Context, key, value string, expiration time.Duration) error {
	return r.rdb.Set(ctx, key, value, expiration).Err()
}

func (r *RedisClient) Close() error {
	return r.rdb.Close()
}
