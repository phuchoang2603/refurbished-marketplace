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

	encodedItems := make([]outboxItem, 0, len(items))
	for _, item := range items {
		encodedItems = append(encodedItems, outboxItem{
			ProductID:      item.ProductID.String(),
			Quantity:       item.Quantity,
			UnitPriceCents: item.UnitPriceCents,
		})
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
