package testutil

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"
	rediscontainer "github.com/testcontainers/testcontainers-go/modules/redis"
)

func SetupRedisContainer(t *testing.T) *redis.Client {
	t.Helper()

	ctx := context.Background()
	container, err := rediscontainer.Run(ctx, "docker.io/valkey/valkey:7.2.5")
	if err != nil {
		t.Fatalf("start redis container: %v", err)
	}

	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("terminate redis container: %v", err)
		}
	})

	connStr, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("redis connection string: %v", err)
	}

	opt, err := redis.ParseURL(connStr)
	if err != nil {
		t.Fatalf("parse redis url: %v", err)
	}

	rdb := redis.NewClient(opt)
	t.Cleanup(func() {
		if err := rdb.Close(); err != nil {
			t.Fatalf("close redis client: %v", err)
		}
	})

	return rdb
}
