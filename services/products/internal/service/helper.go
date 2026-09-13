package service

import (
	"strings"

	"github.com/phuchoang2603/refurbished-marketplace/services/products/internal/database"

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

func validateListPagination(limit, offset int32) error {
	if limit <= 0 || limit > 100 {
		return ErrInvalidListLimit
	}
	if offset < 0 {
		return ErrInvalidListOffset
	}
	return nil
}
