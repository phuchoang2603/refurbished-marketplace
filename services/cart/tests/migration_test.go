package tests

import (
	"testing"
	"time"

	"github.com/phuchoang2603/refurbished-marketplace/services/cart/internal/migration"
	testredis "github.com/phuchoang2603/refurbished-marketplace/shared/testutil/redis"
)

func TestRedisMigrationPreservesCartsAndExpiry(t *testing.T) {
	source := testredis.SetupRedisContainer(t)
	destination := testredis.SetupRedisContainer(t)
	ctx := t.Context()
	for key, ttl := range map[string]time.Duration{"cart:expiring": 2 * time.Hour, "cart:persistent": 0, "unrelated": 0} {
		if err := source.Set(ctx, key, "cart-content", ttl).Err(); err != nil {
			t.Fatal(err)
		}
	}
	copied, err := migration.CopyCarts(ctx, source, destination)
	if err != nil || copied != 2 {
		t.Fatalf("copied=%d err=%v", copied, err)
	}
	for _, key := range []string{"cart:expiring", "cart:persistent"} {
		if value, err := destination.Get(ctx, key).Result(); err != nil || value != "cart-content" {
			t.Fatalf("key=%s value=%s err=%v", key, value, err)
		}
		if exists := source.Exists(ctx, key).Val(); exists != 1 {
			t.Fatal("source cart removed")
		}
	}
	ttl := destination.PTTL(ctx, "cart:expiring").Val()
	if ttl <= time.Hour || ttl > 2*time.Hour {
		t.Fatalf("incorrect expiry: %v", ttl)
	}
	if ttl := destination.PTTL(ctx, "cart:persistent").Val(); ttl != -1 {
		t.Fatalf("persistent cart has TTL %v", ttl)
	}
	if exists := destination.Exists(ctx, "unrelated").Val(); exists != 0 {
		t.Fatal("copied unrelated key")
	}
	if _, err := migration.CopyCarts(ctx, source, destination); err == nil {
		t.Fatal("overwrote destination carts")
	}
}
