package service

import (
	"refurbished-marketplace/services/inventory/internal/database"

	"github.com/google/uuid"
)

func validateProductID(productID uuid.UUID) error {
	if productID == uuid.Nil {
		return ErrInvalidProductID
	}
	return nil
}

func validatePositiveQuantity(quantity int32) error {
	if quantity <= 0 {
		return ErrInvalidQuantity
	}
	return nil
}

func mapDBInventory(i database.Inventory) Inventory {
	return Inventory{
		ProductID:    i.ProductID,
		AvailableQty: i.AvailableQty,
		ReservedQty:  i.ReservedQty,
		CreatedAt:    i.CreatedAt,
		UpdatedAt:    i.UpdatedAt,
	}
}
