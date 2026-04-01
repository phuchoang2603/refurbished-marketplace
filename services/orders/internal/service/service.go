package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"refurbished-marketplace/services/orders/internal/database"

	"github.com/google/uuid"
)

var (
	ErrInvalidBuyerID   = errors.New("invalid buyer user id")
	ErrInvalidProductID = errors.New("invalid product id")
	ErrInvalidQuantity  = errors.New("invalid quantity")
	ErrOrderNotFound    = errors.New("order not found")
	ErrInvalidStatus    = errors.New("invalid order status")
)

const (
	OrderStatusUnspecified = "ORDER_STATUS_UNSPECIFIED"
	OrderStatusPending     = "ORDER_STATUS_PENDING"
	OrderStatusPaid        = "ORDER_STATUS_PAID"
	OrderStatusFailed      = "ORDER_STATUS_FAILED"
)

type OrderItemInput struct {
	ProductID      uuid.UUID
	Quantity       int32
	UnitPriceCents int64
}

type Order struct {
	ID          uuid.UUID
	BuyerUserID uuid.UUID
	Status      string
	TotalCents  int64
	Items       []OrderItem
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type OrderItem struct {
	ID             uuid.UUID
	OrderID        uuid.UUID
	ProductID      uuid.UUID
	Quantity       int32
	UnitPriceCents int64
	LineTotalCents int64
	CreatedAt      time.Time
}

type Service struct {
	db *sql.DB
}

func New(db *sql.DB) *Service {
	return &Service{db: db}
}

func (s *Service) CreateOrder(ctx context.Context, buyerUserID uuid.UUID, items []OrderItemInput, totalCents int64) (Order, error) {
	if buyerUserID == uuid.Nil {
		return Order{}, ErrInvalidBuyerID
	}
	if len(items) == 0 {
		return Order{}, ErrInvalidProductID
	}
	if totalCents <= 0 {
		return Order{}, fmt.Errorf("invalid total cents")
	}
	for _, item := range items {
		if item.ProductID == uuid.Nil {
			return Order{}, ErrInvalidProductID
		}
		if item.Quantity <= 0 {
			return Order{}, ErrInvalidQuantity
		}
		if item.UnitPriceCents <= 0 {
			return Order{}, fmt.Errorf("invalid unit price cents")
		}
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Order{}, err
	}
	queries := database.New(tx)
	defer func() {
		_ = tx.Rollback()
	}()

	created, err := queries.CreateOrder(ctx, database.CreateOrderParams{
		ID:          uuid.New(),
		BuyerUserID: buyerUserID,
		Status:      OrderStatusPending,
		TotalCents:  totalCents,
	})
	if err != nil {
		return Order{}, err
	}

	orderItems := make([]OrderItem, 0, len(items))
	for _, item := range items {
		createdItem, err := queries.CreateOrderItem(ctx, database.CreateOrderItemParams{
			ID:             uuid.New(),
			OrderID:        created.ID,
			ProductID:      item.ProductID,
			Quantity:       item.Quantity,
			UnitPriceCents: item.UnitPriceCents,
			LineTotalCents: item.UnitPriceCents * int64(item.Quantity),
		})
		if err != nil {
			return Order{}, err
		}
		orderItems = append(orderItems, mapDBOrderItem(createdItem))
	}

	if err := tx.Commit(); err != nil {
		return Order{}, err
	}

	createdOrder := mapDBOrder(created)
	createdOrder.Items = orderItems
	return createdOrder, nil
}

func (s *Service) GetOrderByID(ctx context.Context, id uuid.UUID) (Order, error) {
	if id == uuid.Nil {
		return Order{}, ErrOrderNotFound
	}

	queries := database.New(s.db)
	got, err := queries.GetOrderByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Order{}, ErrOrderNotFound
		}
		return Order{}, err
	}

	items, err := queries.ListOrderItemsByOrderID(ctx, id)
	if err != nil {
		return Order{}, err
	}

	order := mapDBOrder(got)
	order.Items = make([]OrderItem, 0, len(items))
	for _, item := range items {
		order.Items = append(order.Items, mapDBOrderItem(item))
	}
	return order, nil
}

func (s *Service) ListOrdersByBuyer(ctx context.Context, buyerUserID uuid.UUID, limit, offset int32) ([]Order, error) {
	if buyerUserID == uuid.Nil {
		return nil, ErrInvalidBuyerID
	}
	if limit <= 0 || limit > 100 {
		return nil, ErrInvalidQuantity
	}
	if offset < 0 {
		return nil, ErrInvalidQuantity
	}

	queries := database.New(s.db)
	rows, err := queries.ListOrdersByBuyer(ctx, database.ListOrdersByBuyerParams{BuyerUserID: buyerUserID, Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}

	result := make([]Order, 0, len(rows))
	for _, row := range rows {
		order := mapDBOrder(row)
		items, err := queries.ListOrderItemsByOrderID(ctx, row.ID)
		if err != nil {
			return nil, err
		}
		order.Items = make([]OrderItem, 0, len(items))
		for _, item := range items {
			order.Items = append(order.Items, mapDBOrderItem(item))
		}
		result = append(result, order)
	}
	return result, nil
}

func (s *Service) UpdateOrderStatus(ctx context.Context, id uuid.UUID, status string) (Order, error) {
	if id == uuid.Nil {
		return Order{}, ErrOrderNotFound
	}
	status = strings.TrimSpace(strings.ToUpper(status))
	if status == "" || status == OrderStatusUnspecified {
		return Order{}, ErrInvalidStatus
	}
	if status != OrderStatusPending && status != OrderStatusPaid && status != OrderStatusFailed {
		return Order{}, ErrInvalidStatus
	}

	queries := database.New(s.db)
	updated, err := queries.UpdateOrderStatus(ctx, database.UpdateOrderStatusParams{ID: id, Status: status})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Order{}, ErrOrderNotFound
		}
		return Order{}, err
	}

	order := mapDBOrder(updated)
	items, err := queries.ListOrderItemsByOrderID(ctx, id)
	if err != nil {
		return Order{}, err
	}
	order.Items = make([]OrderItem, 0, len(items))
	for _, item := range items {
		order.Items = append(order.Items, mapDBOrderItem(item))
	}
	return order, nil
}

func mapDBOrder(o database.Order) Order {
	return Order{
		ID:          o.ID,
		BuyerUserID: o.BuyerUserID,
		Status:      o.Status,
		TotalCents:  o.TotalCents,
		CreatedAt:   o.CreatedAt,
		UpdatedAt:   o.UpdatedAt,
	}
}

func mapDBOrderItem(i database.OrderItem) OrderItem {
	return OrderItem{
		ID:             i.ID,
		OrderID:        i.OrderID,
		ProductID:      i.ProductID,
		Quantity:       i.Quantity,
		UnitPriceCents: i.UnitPriceCents,
		LineTotalCents: i.LineTotalCents,
		CreatedAt:      i.CreatedAt,
	}
}
