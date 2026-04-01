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

// AcquireLock tries to acquire a distributed lock with the given key and TTL.
// Returns true if lock was acquired, false otherwise.
func (r *RedisClient) AcquireLock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	// Use SET NX EX for atomic lock acquisition
	result, err := r.Client.SetNX(ctx, key, "1", ttl).Result()
	if err != nil {
		return false, fmt.Errorf("acquire lock: %w", err)
	}
	return result, nil
}

// ReleaseLock releases a distributed lock.
func (r *RedisClient) ReleaseLock(ctx context.Context, key string) error {
	return r.Client.Del(ctx, key).Err()
}
