package service

import (
	"database/sql"
	"errors"

	"refurbished-marketplace/services/inventory/internal/database"
)

var (
	ErrInvalidProductID  = errors.New("invalid product id")
	ErrInvalidQuantity   = errors.New("invalid quantity")
	ErrInventoryNotFound = errors.New("inventory not found")
	ErrInsufficientStock = errors.New("insufficient stock")
)

type Service struct {
	queries *database.Queries
}

func New(db *sql.DB) *Service {
	return &Service{queries: database.New(db)}
}
