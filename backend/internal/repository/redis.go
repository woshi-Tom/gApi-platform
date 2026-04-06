package repository

import (
	"context"
	"fmt"
	"time"

	"gapi-platform/internal/config"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
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
// Returns the lock token if acquired, empty string otherwise.
func (r *RedisClient) AcquireLock(ctx context.Context, key string, ttl time.Duration) (string, error) {
	token := uuid.New().String()
	ok, err := r.Client.SetNX(ctx, key, token, ttl).Result()
	if err != nil {
		return "", fmt.Errorf("acquire lock: %w", err)
	}
	if !ok {
		return "", nil
	}
	return token, nil
}

// ReleaseLock releases a distributed lock using ownership token.
// Uses Lua script to ensure atomic check-and-delete.
func (r *RedisClient) ReleaseLock(ctx context.Context, key string, token string) error {
	script := redis.NewScript(`
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`)
	_, err := script.Run(ctx, r.Client, []string{key}, token).Result()
	return err
}
