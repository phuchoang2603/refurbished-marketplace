package migration

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const readCart = `
local value = redis.call('DUMP', KEYS[1])
if not value then return {} end
return {value, redis.call('PTTL', KEYS[1])}
`

func CopyCarts(ctx context.Context, source, destination *redis.Client) (int, error) {
	if source.Options().Addr == destination.Options().Addr && source.Options().DB == destination.Options().DB {
		return 0, fmt.Errorf("source and destination must differ")
	}
	copied := 0
	seen := make(map[string]bool)
	var cursor uint64
	for {
		keys, next, err := source.Scan(ctx, cursor, "cart:*", 100).Result()
		if err != nil {
			return copied, fmt.Errorf("scan source: %w", err)
		}
		for _, key := range keys {
			if seen[key] {
				continue
			}
			seen[key] = true
			started := time.Now()
			values, err := source.Eval(ctx, readCart, []string{key}).Slice()
			if err != nil {
				return copied, fmt.Errorf("read source cart: %w", err)
			}
			if len(values) == 0 {
				continue
			}
			payload, ok := values[0].(string)
			if !ok || len(values) != 2 {
				return copied, fmt.Errorf("invalid source cart snapshot")
			}
			ttl, err := strconv.ParseInt(fmt.Sprint(values[1]), 10, 64)
			if err != nil || ttl < -1 {
				return copied, fmt.Errorf("invalid source cart TTL")
			}
			var remaining time.Duration
			if ttl >= 0 {
				remaining = time.Duration(ttl)*time.Millisecond - time.Since(started)
				if remaining < time.Millisecond {
					continue
				}
			}
			if err := destination.Restore(ctx, key, remaining, payload).Err(); err != nil {
				return copied, fmt.Errorf("restore cart (destination must have no conflicting carts): %w", err)
			}
			copied++
		}
		if next == 0 {
			return copied, nil
		}
		cursor = next
	}
}
