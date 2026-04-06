package service

import (
	"context"
	"fmt"
	"time"

	"gapi-platform/internal/repository"
)

const (
	redemptionCacheKeyPrefix = "redeem:"
	redemptionLockKeyPrefix  = "redeem_lock:"
	redemptionCacheTTL       = 24 * time.Hour
)

type RedemptionCacheService struct {
	redis *repository.RedisClient
}

func NewRedemptionCacheService(redis *repository.RedisClient) *RedemptionCacheService {
	return &RedemptionCacheService{redis: redis}
}

func (s *RedemptionCacheService) cacheKey(codeID, userID uint) string {
	return fmt.Sprintf("%s%d:%d", redemptionCacheKeyPrefix, codeID, userID)
}

func (s *RedemptionCacheService) lockKey(codeID, userID uint) string {
	return fmt.Sprintf("%s%d:%d", redemptionLockKeyPrefix, codeID, userID)
}

func (s *RedemptionCacheService) IsRedeemed(ctx context.Context, codeID, userID uint) (bool, error) {
	if s.redis == nil {
		return false, nil
	}
	key := s.cacheKey(codeID, userID)
	exists, err := s.redis.Client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

func (s *RedemptionCacheService) MarkRedeemed(ctx context.Context, codeID, userID uint) error {
	if s.redis == nil {
		return nil
	}
	key := s.cacheKey(codeID, userID)
	return s.redis.Client.Set(ctx, key, "1", redemptionCacheTTL).Err()
}

func (s *RedemptionCacheService) AcquireLock(ctx context.Context, codeID, userID uint, ttl time.Duration) (string, error) {
	if s.redis == nil {
		return "", fmt.Errorf("redis not available")
	}
	key := s.lockKey(codeID, userID)
	return s.redis.AcquireLock(ctx, key, ttl)
}

func (s *RedemptionCacheService) ReleaseLock(ctx context.Context, codeID, userID uint, token string) error {
	if s.redis == nil {
		return nil
	}
	key := s.lockKey(codeID, userID)
	return s.redis.ReleaseLock(ctx, key, token)
}
