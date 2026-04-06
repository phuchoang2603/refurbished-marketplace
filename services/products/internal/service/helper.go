package service

import (
	"database/sql"
	"strings"
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
	if err == sql.ErrNoRows {
		return ErrProductNotFound
	}
	return err
}
