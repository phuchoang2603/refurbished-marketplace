package service

import (
	"database/sql"
	"errors"

	"refurbished-marketplace/services/orders/internal/database"
)

var (
	ErrInvalidBuyerID        = errors.New("invalid buyer user id")
	ErrInvalidMerchantID     = errors.New("invalid merchant id")
	ErrInvalidProductID      = errors.New("invalid product id")
	ErrInvalidQuantity       = errors.New("invalid quantity")
	ErrInvalidTotalCents     = errors.New("invalid total cents")
	ErrInvalidUnitPriceCents = errors.New("invalid unit price cents")
	ErrOrderNotFound         = errors.New("order not found")
	ErrInvalidStatus         = errors.New("invalid order status")
)

const (
	OrderStatusUnspecified = "ORDER_STATUS_UNSPECIFIED"
	OrderStatusPending     = "ORDER_STATUS_PENDING"
	OrderStatusPaid        = "ORDER_STATUS_PAID"
	OrderStatusFailed      = "ORDER_STATUS_FAILED"
)

type Service struct {
	db      *sql.DB
	queries *database.Queries
}

func New(db *sql.DB) *Service {
	return &Service{db: db, queries: database.New(db)}
}
