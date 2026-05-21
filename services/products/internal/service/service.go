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
	ErrInvalidProductID   = errors.New("invalid product id")
	ErrInvalidQuantity    = errors.New("invalid quantity")
	ErrInventoryNotFound  = errors.New("inventory not found")
	ErrInsufficientStock  = errors.New("insufficient stock")
)

const (
	ReservationStatusReserved  = "RESERVED"
	ReservationStatusCommitted = "COMMITTED"
	ReservationStatusReleased  = "RELEASED"
)

type Service struct {
	db      *sql.DB
	queries *database.Queries
}

func New(db *sql.DB) *Service {
	return &Service{db: db, queries: database.New(db)}
}
