package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type JSONStore struct {
	client *redis.Client
	prefix string
	ttl    time.Duration
}

func NewClient(addr string) *redis.Client {
	return redis.NewClient(&redis.Options{Addr: addr})
}

func NewJSONStore(client *redis.Client, prefix string, ttl time.Duration) *JSONStore {
	return &JSONStore{client: client, prefix: prefix, ttl: ttl}
}

func (s *JSONStore) Set(ctx context.Context, key string, value any) error {
	buf, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return s.client.Set(ctx, s.key(key), buf, s.ttl).Err()
}

func (s *JSONStore) Load(ctx context.Context, key string, dest any) (bool, error) {
	val, err := s.client.Get(ctx, s.key(key)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, err
	}
	return true, json.Unmarshal(val, dest)
}

func (s *JSONStore) Delete(ctx context.Context, key string) error {
	return s.client.Del(ctx, s.key(key)).Err()
}

func (s *JSONStore) key(key string) string {
	if s.prefix == "" {
		return key
	}
	return s.prefix + ":" + key
}
