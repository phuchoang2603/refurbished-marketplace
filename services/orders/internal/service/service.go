// Package service provides the core business logic for managing orders in the refurbished marketplace application. It defines the Service struct, which interacts with the database to perform operations such as creating orders, listing orders by buyer, and updating order statuses.
package service

import (
	"context"
	"database/sql"
	"errors"

	"refurbished-marketplace/services/orders/internal/database"

	"github.com/google/uuid"
)

var (
	ErrInvalidBuyerID        = errors.New("invalid buyer user id")
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

type queryStore interface {
	CreateOrder(ctx context.Context, arg database.CreateOrderParams) (database.Order, error)
	CreateOrderItem(ctx context.Context, arg database.CreateOrderItemParams) (database.OrderItem, error)
	CreateOrderOutbox(ctx context.Context, arg database.CreateOrderOutboxParams) (database.OrdersOutbox, error)
	GetOrderByID(ctx context.Context, id uuid.UUID) (database.Order, error)
	ListOrdersByBuyer(ctx context.Context, arg database.ListOrdersByBuyerParams) ([]database.Order, error)
	ListOrderItemsByOrderID(ctx context.Context, orderID uuid.UUID) ([]database.OrderItem, error)
	UpdateOrderStatus(ctx context.Context, arg database.UpdateOrderStatusParams) (database.Order, error)
}

type Service struct {
	db *sql.DB
}

func New(db *sql.DB) *Service {
	return &Service{db: db}
}
