package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"refurbished-marketplace/services/orders/internal/database"
	"refurbished-marketplace/shared/messaging"

	"github.com/google/uuid"
)

func (s *Service) CreateOrder(ctx context.Context, buyerUserID uuid.UUID, items []OrderItemInput, totalCents int64) (Order, error) {
	if err := validateCreateOrderInput(buyerUserID, items, totalCents); err != nil {
		return Order{}, err
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

	orderItems, encodedItems, err := createOrderItems(ctx, queries, created.ID, items)
	if err != nil {
		return Order{}, err
	}

	payload, err := json.Marshal(outboxPayload{
		OrderID:     created.ID.String(),
		BuyerUserID: buyerUserID.String(),
		TotalCents:  totalCents,
		Items:       encodedItems,
	})
	if err != nil {
		return Order{}, err
	}

	if _, err := queries.CreateOrderOutbox(ctx, database.CreateOrderOutboxParams{
		ID:          uuid.New(),
		AggregateID: created.ID,
		EventType:   string(messaging.EventTypeOrderCreated),
		Payload:     payload,
	}); err != nil {
		return Order{}, err
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

	return loadOrderWithItems(ctx, queries, got)
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

	return loadOrdersWithItems(ctx, queries, rows)
}

func (s *Service) UpdateOrderStatus(ctx context.Context, id uuid.UUID, status string) (Order, error) {
	if id == uuid.Nil {
		return Order{}, ErrOrderNotFound
	}
	normalizedStatus, err := validateOrderStatus(status)
	if err != nil {
		return Order{}, ErrInvalidStatus
	}

	queries := database.New(s.db)
	updated, err := queries.UpdateOrderStatus(ctx, database.UpdateOrderStatusParams{ID: id, Status: normalizedStatus})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Order{}, ErrOrderNotFound
		}
		return Order{}, err
	}

	return loadOrderWithItems(ctx, queries, updated)
}
