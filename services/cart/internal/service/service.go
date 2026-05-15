package service

import (
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ErrInvalidCartID     = errors.New("invalid cart id")
	ErrInvalidProductID  = errors.New("invalid product id")
	ErrInvalidMerchantID = errors.New("invalid merchant id")
	ErrInvalidQuantity   = errors.New("invalid quantity")
	ErrItemNotFound      = errors.New("cart item not found")
)

type Service struct {
	client *redis.Client
	ttl    time.Duration
}

func New(rdb *redis.Client, ttl time.Duration) *Service {
	return &Service{client: rdb, ttl: ttl}
}
