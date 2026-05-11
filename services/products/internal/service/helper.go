package service

import (
	"strings"

	"refurbished-marketplace/shared/dberrors"
)

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

func mapProductNotFound(err error) error {
	if err == nil {
		return nil
	}
	if dberrors.IsNoRows(err) {
		return ErrProductNotFound
	}
	return err
}
