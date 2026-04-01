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

type queryStore interface {
	CreateOrder(ctx context.Context, arg database.CreateOrderParams) (database.Order, error)
	GetOrderByID(ctx context.Context, id uuid.UUID) (database.Order, error)
	ListOrdersByBuyer(ctx context.Context, arg database.ListOrdersByBuyerParams) ([]database.Order, error)
	UpdateOrderStatus(ctx context.Context, arg database.UpdateOrderStatusParams) (database.Order, error)
}

var (
	ErrInvalidBuyerID   = errors.New("invalid buyer user id")
	ErrInvalidProductID = errors.New("invalid product id")
	ErrInvalidQuantity  = errors.New("invalid quantity")
	ErrOrderNotFound    = errors.New("order not found")
	ErrInvalidStatus    = errors.New("invalid order status")
)

type Order struct {
	ID          uuid.UUID
	BuyerUserID uuid.UUID
	ProductID   uuid.UUID
	Quantity    int32
	Status      string
	TotalCents  int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Service struct {
	queries queryStore
}

func New(queries queryStore) *Service {
	return &Service{queries: queries}
}

func (s *Service) CreateOrder(ctx context.Context, buyerUserID, productID uuid.UUID, quantity int32, totalCents int64) (Order, error) {
	if buyerUserID == uuid.Nil {
		return Order{}, ErrInvalidBuyerID
	}
	if productID == uuid.Nil {
		return Order{}, ErrInvalidProductID
	}
	if quantity <= 0 {
		return Order{}, ErrInvalidQuantity
	}
	if totalCents <= 0 {
		return Order{}, fmt.Errorf("invalid total cents")
	}

	created, err := s.queries.CreateOrder(ctx, database.CreateOrderParams{
		ID:          uuid.New(),
		BuyerUserID: buyerUserID,
		ProductID:   productID,
		Quantity:    quantity,
		Status:      "PENDING",
		TotalCents:  totalCents,
	})
	if err != nil {
		return Order{}, err
	}

	return mapDBOrder(created), nil
}

func (s *Service) GetOrderByID(ctx context.Context, id uuid.UUID) (Order, error) {
	if id == uuid.Nil {
		return Order{}, ErrOrderNotFound
	}

	got, err := s.queries.GetOrderByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Order{}, ErrOrderNotFound
		}
		return Order{}, err
	}

	return mapDBOrder(got), nil
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

	rows, err := s.queries.ListOrdersByBuyer(ctx, database.ListOrdersByBuyerParams{BuyerUserID: buyerUserID, Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}

	result := make([]Order, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapDBOrder(row))
	}
	return result, nil
}

func (s *Service) UpdateOrderStatus(ctx context.Context, id uuid.UUID, status string) (Order, error) {
	if id == uuid.Nil {
		return Order{}, ErrOrderNotFound
	}
	status = strings.TrimSpace(strings.ToUpper(status))
	if status == "" {
		return Order{}, ErrInvalidStatus
	}

	updated, err := s.queries.UpdateOrderStatus(ctx, database.UpdateOrderStatusParams{ID: id, Status: status})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Order{}, ErrOrderNotFound
		}
		return Order{}, err
	}

	return mapDBOrder(updated), nil
}

func mapDBOrder(o database.Order) Order {
	return Order{
		ID:          o.ID,
		BuyerUserID: o.BuyerUserID,
		ProductID:   o.ProductID,
		Quantity:    o.Quantity,
		Status:      o.Status,
		TotalCents:  o.TotalCents,
		CreatedAt:   o.CreatedAt,
		UpdatedAt:   o.UpdatedAt,
	}
}
