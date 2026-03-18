package database

import (
	"context"
	"log/slog"

	"github.com/redis/go-redis/v9"
)

func CreateRedisClient(redisUrl string) (*redis.Client, error) {
	opts, err := redis.ParseURL(redisUrl)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opts)

	if err := client.Ping(context.Background()).Err(); err != nil {
		client.Close()
		return nil, err
	}

	slog.Info("Successfully connected to Redis")
	return client, nil
}
