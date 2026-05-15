package service

import (
	"strings"

	"refurbished-marketplace/services/products/internal/database"
	"refurbished-marketplace/shared/dberrors"
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

func mapProductNotFound(err error) error {
	if err == nil {
		return nil
	}
	if dberrors.IsNoRows(err) {
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
