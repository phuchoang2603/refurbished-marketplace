package service

import (
	"database/sql"
	"errors"

	"refurbished-marketplace/services/products/internal/database"
)

var (
	ErrInvalidProductName = errors.New("invalid product name")
	ErrInvalidPrice       = errors.New("invalid product price")
	ErrInvalidMerchantID  = errors.New("invalid merchant id")
	ErrProductNotFound    = errors.New("product not found")
	ErrInvalidListLimit   = errors.New("invalid list limit")
	ErrInvalidListOffset  = errors.New("invalid list offset")
)

type Service struct {
	queries *database.Queries
}

func New(db *sql.DB) *Service {
	return &Service{queries: database.New(db)}
}
