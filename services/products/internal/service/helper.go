package service

import (
	"database/sql"
	"errors"
	"strings"

	"refurbished-marketplace/services/products/internal/database"

	"github.com/google/uuid"
)

func mapDBProduct(p database.Product) Product {
	return Product{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		PriceCents:  p.PriceCents,
		MerchantID:  p.MerchantID,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func mapDBProductRow(p database.GetProductByIDRow) Product {
	product := Product{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		PriceCents:  p.PriceCents,
		MerchantID:  p.MerchantID,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
	if p.AvailableQty.Valid {
		qty := p.AvailableQty.Int32
		product.AvailableQty = &qty
	}
	if p.ReservedQty.Valid {
		qty := p.ReservedQty.Int32
		product.ReservedQty = &qty
	}
	return product
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

func mapProductNotFound(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return ErrProductNotFound
	}
	return err
}

func normalizeProductName(name string) string {
	return strings.TrimSpace(name)
}

func normalizeProductDescription(description, fallback string) string {
	desc := strings.TrimSpace(description)
	if desc == "" {
		return fallback
	}
	return desc
}

func validateProductID(productID uuid.UUID) error {
	if productID == uuid.Nil {
		return ErrInvalidProductID
	}
	return nil
}

func validateNonNegativeQuantity(quantity int32) error {
	if quantity < 0 {
		return ErrInvalidQuantity
	}
	return nil
}

func validatePositiveQuantity(quantity int32) error {
	if quantity <= 0 {
		return ErrInvalidQuantity
	}
	return nil
}
