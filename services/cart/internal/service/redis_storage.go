package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

func cartKey(cartID string) string {
	return "cart:" + cartID
}

func newCart(cartID string) Cart {
	now := time.Now().UTC()
	return Cart{
		CartID:    cartID,
		Items:     []CartItem{},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (s *Service) loadCart(ctx context.Context, cartID string) (Cart, bool, error) {
	val, err := s.client.Get(ctx, cartKey(cartID)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return Cart{}, false, nil
		}
		return Cart{}, false, err
	}

	var cart Cart
	if err := json.Unmarshal(val, &cart); err != nil {
		return Cart{}, false, err
	}
	if cart.Items == nil {
		cart.Items = []CartItem{}
	}
	return cart, true, nil
}

func (s *Service) saveCart(ctx context.Context, cart Cart) error {
	buf, err := json.Marshal(cart)
	if err != nil {
		return err
	}
	return s.client.Set(ctx, cartKey(cart.CartID), buf, s.cartTTL()).Err()
}

func (s *Service) deleteCart(ctx context.Context, cartID string) error {
	return s.client.Del(ctx, cartKey(cartID)).Err()
}
