package service

import (
	"time"

	"github.com/google/uuid"
)

func toCart(state cartState) Cart {
	items := make([]CartItem, 0, len(state.Items))
	for productID, quantity := range state.Items {
		items = append(items, CartItem{ProductID: productID, Quantity: quantity})
	}
	return Cart{CartID: state.CartID, Items: items, CreatedAt: state.CreatedAt, UpdatedAt: state.UpdatedAt}
}

func newCartState(cartID string) cartState {
	now := time.Now().UTC()
	return cartState{
		CartID:    cartID,
		Items:     map[string]int32{},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func validate(id string, errType error) error {
	if id == "" {
		return errType
	}
	if _, err := uuid.Parse(id); err != nil {
		return errType
	}
	return nil
}
