package service

import (
	"context"
	"time"

	"refurbished-marketplace/services/orders/internal/database"
	"refurbished-marketplace/shared/dberrors"

	"github.com/google/uuid"
)

type OrderItemInput struct {
	ProductID      uuid.UUID
	Quantity       int32
	UnitPriceCents int64
}

type Order struct {
	ID          uuid.UUID
	BuyerUserID uuid.UUID
	MerchantID  uuid.UUID
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

func (s *Service) CreateOrder(ctx context.Context, buyerUserID, merchantID uuid.UUID, items []OrderItemInput, totalCents int64) (Order, error) {
	if err := validateCreateOrderInput(buyerUserID, merchantID, items, totalCents); err != nil {
		return Order{}, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Order{}, err
	}
	q := database.New(tx)
	defer func() {
		_ = tx.Rollback()
	}()

	created, err := q.CreateOrder(ctx, database.CreateOrderParams{
		ID:          uuid.New(),
		BuyerUserID: buyerUserID,
		MerchantID:  merchantID,
		Status:      OrderStatusPending,
		TotalCents:  totalCents,
	})
	if err != nil {
		return Order{}, err
	}

	orderItems, err := createOrderItems(ctx, q, created.ID, items)
	if err != nil {
		return Order{}, err
	}

	createdOrder := mapDBOrder(created)
	createdOrder.Items = orderItems
	if err := createOrderOutbox(ctx, q, createdOrder); err != nil {
		return Order{}, err
	}

	if err := tx.Commit(); err != nil {
		return Order{}, err
	}

	return createdOrder, nil
}

func (s *Service) GetOrderByID(ctx context.Context, id uuid.UUID) (Order, error) {
	if id == uuid.Nil {
		return Order{}, ErrOrderNotFound
	}

	got, err := s.queries.GetOrderByID(ctx, id)
	if err != nil {
		if dberrors.IsNoRows(err) {
			return Order{}, ErrOrderNotFound
		}
		return Order{}, err
	}

	orders, err := loadOrdersWithItems(ctx, s.queries, []Order{mapDBOrder(got)})
	if err != nil {
		return Order{}, err
	}
	if len(orders) == 0 {
		return Order{}, ErrOrderNotFound
	}
	return orders[0], nil
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

	orders := make([]Order, 0, len(rows))
	for _, row := range rows {
		orders = append(orders, mapDBOrder(row))
	}
	return loadOrdersWithItems(ctx, s.queries, orders)
}

func (s *Service) UpdateOrderStatus(ctx context.Context, id uuid.UUID, status string) (Order, error) {
	if id == uuid.Nil {
		return Order{}, ErrOrderNotFound
	}
	normalizedStatus, err := validateOrderStatus(status)
	if err != nil {
		return Order{}, ErrInvalidStatus
	}

	updated, err := s.queries.UpdateOrderStatus(ctx, database.UpdateOrderStatusParams{ID: id, Status: normalizedStatus})
	if err != nil {
		if dberrors.IsNoRows(err) {
			return Order{}, ErrOrderNotFound
		}
		return Order{}, err
	}

	orders, err := loadOrdersWithItems(ctx, s.queries, []Order{mapDBOrder(updated)})
	if err != nil {
		return Order{}, err
	}
	if len(orders) == 0 {
		return Order{}, ErrOrderNotFound
	}
	return orders[0], nil
}
