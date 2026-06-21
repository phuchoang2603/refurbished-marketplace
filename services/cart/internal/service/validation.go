package service

import (
	"github.com/google/uuid"
)

func findCartItem(items []CartItem, productID string) int {
	for i, item := range items {
		if item.ProductID == productID {
			return i
		}
	}
	return -1
}

func validateUUID(id string, errType error) error {
	if id == "" {
		return errType
	}
	if _, err := uuid.Parse(id); err != nil {
		return errType
	}
	return nil
}
