package repository

import (
	"context"
	"fmt"
	"time"

	"gapi-platform/internal/config"
	"github.com/go-redis/redis/v8"
)

// RedisClient wraps redis client
type RedisClient struct {
	Client *redis.Client
}

// NewRedis creates a new Redis connection
func NewRedis(cfg *config.RedisConfig) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return &RedisClient{Client: client}, nil
}

// Close closes the Redis connection
func (r *RedisClient) Close() error {
	return r.Client.Close()
}
