package service

import (
	"context"
	"testing"
	"time"
)

func TestRedemptionCacheService_KeyFormats(t *testing.T) {
	service := &RedemptionCacheService{}

	tests := []struct {
		name   string
		codeID uint
		userID uint
	}{
		{"basic ids", 1, 1},
		{"large ids", 99999, 88888},
		{"zero values", 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cacheKey := service.cacheKey(tt.codeID, tt.userID)
			lockKey := service.lockKey(tt.codeID, tt.userID)

			if cacheKey == "" {
				t.Error("cacheKey should not be empty")
			}
			if lockKey == "" {
				t.Error("lockKey should not be empty")
			}
			if cacheKey == lockKey {
				t.Error("cacheKey and lockKey should be different")
			}
		})
	}
}

func TestRedemptionCacheService_NilRedis(t *testing.T) {
	service := &RedemptionCacheService{redis: nil}
	ctx := context.Background()

	t.Run("IsRedeemed returns false when redis is nil", func(t *testing.T) {
		result, err := service.IsRedeemed(ctx, 1, 1)
		if err != nil {
			t.Errorf("IsRedeemed() error = %v", err)
		}
		if result != false {
			t.Errorf("IsRedeemed() = %v, want false", result)
		}
	})

	t.Run("MarkRedeemed returns nil when redis is nil", func(t *testing.T) {
		err := service.MarkRedeemed(ctx, 1, 1)
		if err != nil {
			t.Errorf("MarkRedeemed() error = %v", err)
		}
	})

	t.Run("AcquireLock returns error when redis is nil", func(t *testing.T) {
		token, err := service.AcquireLock(ctx, 1, 1, 10*time.Second)
		if err == nil {
			t.Error("AcquireLock() should return error when redis is nil")
		}
		if token != "" {
			t.Errorf("AcquireLock() token = %v, want empty", token)
		}
	})

	t.Run("ReleaseLock returns nil when redis is nil", func(t *testing.T) {
		err := service.ReleaseLock(ctx, 1, 1, "token")
		if err != nil {
			t.Errorf("ReleaseLock() error = %v", err)
		}
	})
}

func TestRedemptionCacheService_KeyPrefix(t *testing.T) {
	service := &RedemptionCacheService{}

	key := service.cacheKey(123, 456)

	expected := "redeem:123:456"
	if key != expected {
		t.Errorf("cacheKey() = %v, want %v", key, expected)
	}
}

func TestRedemptionCacheService_LockKeyPrefix(t *testing.T) {
	service := &RedemptionCacheService{}

	key := service.lockKey(123, 456)

	expected := "redeem_lock:123:456"
	if key != expected {
		t.Errorf("lockKey() = %v, want %v", key, expected)
	}
}
