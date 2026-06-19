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
	cfg    Config
}

func New(client *redis.Client, cfg Config) *Service {
	return &Service{client: client, cfg: cfg}
}

func (s *Service) cartTTL() time.Duration {
	return s.cfg.CartTTL
}
