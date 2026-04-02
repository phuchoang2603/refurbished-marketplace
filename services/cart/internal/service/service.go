package service

import (
	"errors"
	"time"

	cache "refurbished-marketplace/shared/cache"

	"github.com/redis/go-redis/v9"
)

var (
	ErrInvalidCartID    = errors.New("invalid cart id")
	ErrInvalidProductID = errors.New("invalid product id")
	ErrInvalidQuantity  = errors.New("invalid quantity")
	ErrItemNotFound     = errors.New("cart item not found")
)

type Service struct {
	store *cache.JSONStore
}

func New(rdb *redis.Client, ttl time.Duration) *Service {
	return &Service{store: cache.NewJSONStore(rdb, "cart", ttl)}
}
