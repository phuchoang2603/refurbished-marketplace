package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/phuchoang2603/refurbished-marketplace/services/orders/internal/database"

	"github.com/google/uuid"
)

var (
	ErrInvalidBuyerID        = errors.New("invalid buyer user id")
	ErrInvalidMerchantID     = errors.New("invalid merchant id")
	ErrInvalidProductID      = errors.New("invalid product id")
	ErrInvalidQuantity       = errors.New("invalid quantity")
	ErrInvalidPagination     = errors.New("invalid pagination")
	ErrInvalidTotalCents     = errors.New("invalid total cents")
	ErrInvalidUnitPriceCents = errors.New("invalid unit price cents")
	ErrOrderNotFound         = errors.New("order not found")
	ErrInvalidStatus         = errors.New("invalid order status")
	ErrInvalidIdempotencyKey = errors.New("invalid idempotency key")
	ErrIdempotencyConflict   = errors.New("idempotency key conflict")
	ErrOrderNotPayable       = errors.New("order cannot be paid")
	ErrInsufficientStock     = errors.New("insufficient stock")
)

const (
	OrderStatusUnspecified = "ORDER_STATUS_UNSPECIFIED"
	OrderStatusPending     = "ORDER_STATUS_PENDING"
	OrderStatusPaid        = "ORDER_STATUS_PAID"
	OrderStatusFailed      = "ORDER_STATUS_FAILED"
)

type StockReserver interface {
	ReserveStock(ctx context.Context, orderID, merchantID uuid.UUID, totalCents int64, items []OrderItemInput) error
}

type Service struct {
	db      *sql.DB
	queries *database.Queries
	stock   StockReserver
}

func New(db *sql.DB, stock StockReserver) *Service {
	return &Service{db: db, queries: database.New(db), stock: stock}
}
