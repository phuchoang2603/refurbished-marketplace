package service

import (
	"context"
	"time"
)

type Cart struct {
	CartID    string
	Items     []CartItem
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CartItem struct {
	ProductID string
	Quantity  int32
}

type cartState struct {
	CartID    string           `json:"cart_id"`
	Items     map[string]int32 `json:"items"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

func (s *Service) GetCart(ctx context.Context, cartID string) (Cart, error) {
	if err := validate(cartID, ErrInvalidCartID); err != nil {
		return Cart{}, err
	}

	var state cartState
	ok, err := s.store.Load(ctx, cartID, &state)
	if err != nil {
		return Cart{}, err
	}
	if !ok {
		state = newCartState(cartID)
	}

	if state.Items == nil {
		state.Items = map[string]int32{}
	}
	return toCart(state), nil
}

func (s *Service) AddCartItem(ctx context.Context, cartID, productID string, quantity int32) (Cart, error) {
	if quantity <= 0 {
		return Cart{}, ErrInvalidQuantity
	}
	if err := validate(cartID, ErrInvalidCartID); err != nil {
		return Cart{}, err
	}
	if err := validate(productID, ErrInvalidProductID); err != nil {
		return Cart{}, err
	}

	var state cartState
	if err := s.store.Update(ctx, cartID, &state, func(exists bool) error {
		if !exists {
			state = newCartState(cartID)
		}
		if state.Items == nil {
			state.Items = map[string]int32{}
		}
		state.Items[productID] += quantity
		state.UpdatedAt = time.Now().UTC()
		return nil
	}); err != nil {
		return Cart{}, err
	}
	return toCart(state), nil
}

func (s *Service) SetCartItemQuantity(ctx context.Context, cartID, productID string, quantity int32) (Cart, error) {
	if quantity <= 0 {
		return s.RemoveCartItem(ctx, cartID, productID)
	}
	if err := validate(cartID, ErrInvalidCartID); err != nil {
		return Cart{}, err
	}
	if err := validate(productID, ErrInvalidProductID); err != nil {
		return Cart{}, err
	}

	var state cartState
	if err := s.store.Update(ctx, cartID, &state, func(exists bool) error {
		if !exists {
			state = newCartState(cartID)
		}
		if state.Items == nil {
			state.Items = map[string]int32{}
		}
		state.Items[productID] = quantity
		state.UpdatedAt = time.Now().UTC()
		return nil
	}); err != nil {
		return Cart{}, err
	}
	return toCart(state), nil
}

func (s *Service) RemoveCartItem(ctx context.Context, cartID, productID string) (Cart, error) {
	if err := validate(cartID, ErrInvalidCartID); err != nil {
		return Cart{}, err
	}
	if err := validate(productID, ErrInvalidProductID); err != nil {
		return Cart{}, err
	}

	var state cartState
	if err := s.store.Update(ctx, cartID, &state, func(exists bool) error {
		if !exists {
			return ErrItemNotFound
		}
		if _, exists := state.Items[productID]; !exists {
			return ErrItemNotFound
		}
		delete(state.Items, productID)
		state.UpdatedAt = time.Now().UTC()
		return nil
	}); err != nil {
		return Cart{}, err
	}
	return toCart(state), nil
}

func (s *Service) ClearCart(ctx context.Context, cartID string) error {
	if err := validate(cartID, ErrInvalidCartID); err != nil {
		return err
	}
	return s.store.Delete(ctx, cartID)
}
