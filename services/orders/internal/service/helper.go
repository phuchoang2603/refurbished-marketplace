package service

import (
	"context"
	"encoding/json"
	"refurbished-marketplace/services/orders/internal/database"
	"refurbished-marketplace/shared/messaging"
	"strings"

	"github.com/google/uuid"
)

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
		MerchantID:     i.MerchantID,
		Quantity:       i.Quantity,
		UnitPriceCents: i.UnitPriceCents,
		LineTotalCents: i.LineTotalCents,
		CreatedAt:      i.CreatedAt,
	}
}

func validateCreateOrderInput(buyerUserID uuid.UUID, items []OrderItemInput, totalCents int64) error {
	if buyerUserID == uuid.Nil {
		return ErrInvalidBuyerID
	}
	if len(items) == 0 {
		return ErrInvalidProductID
	}
	if totalCents <= 0 {
		return ErrInvalidTotalCents
	}
	for _, item := range items {
		if item.ProductID == uuid.Nil {
			return ErrInvalidProductID
		}
		if item.Quantity <= 0 {
			return ErrInvalidQuantity
		}
		if item.UnitPriceCents <= 0 {
			return ErrInvalidUnitPriceCents
		}
	}
	return nil
}

func validateOrderStatus(status string) (string, error) {
	status = strings.TrimSpace(strings.ToUpper(status))
	if status == "" || status == OrderStatusUnspecified {
		return "", ErrInvalidStatus
	}
	if status != OrderStatusPending && status != OrderStatusPaid && status != OrderStatusFailed {
		return "", ErrInvalidStatus
	}
	return status, nil
}

func createOrderItems(ctx context.Context, queries *database.Queries, orderID, buyerUserID uuid.UUID, items []OrderItemInput) ([]OrderItem, error) {
	orderItems := make([]OrderItem, 0, len(items))

	for _, item := range items {
		itemID := uuid.New()
		lineTotal := item.UnitPriceCents * int64(item.Quantity)
		createdItem, err := queries.CreateOrderItem(ctx, database.CreateOrderItemParams{
			ID:             itemID,
			OrderID:        orderID,
			ProductID:      item.ProductID,
			MerchantID:     item.MerchantID,
			Quantity:       item.Quantity,
			UnitPriceCents: item.UnitPriceCents,
			LineTotalCents: lineTotal,
		})
		if err != nil {
			return nil, err
		}

		payload := outboxItem{
			OrderID:        orderID.String(),
			OrderItemID:    itemID.String(),
			BuyerUserID:    buyerUserID.String(),
			ProductID:      item.ProductID.String(),
			MerchantID:     item.MerchantID.String(),
			Quantity:       item.Quantity,
			UnitPriceCents: item.UnitPriceCents,
			LineTotalCents: lineTotal,
		}
		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}

		if _, err := queries.CreateOrderOutbox(ctx, database.CreateOrderOutboxParams{
			ID:          uuid.New(),
			AggregateID: item.ProductID,
			EventType:   messaging.EventTypeOrderItemCreated,
			Payload:     payloadBytes,
		}); err != nil {
			return nil, err
		}

		orderItems = append(orderItems, mapDBOrderItem(createdItem))
	}

	return orderItems, nil
}

func loadOrderWithItems(ctx context.Context, queries *database.Queries, dbOrder database.Order) (Order, error) {
	orders, err := loadOrdersWithItems(ctx, queries, []database.Order{dbOrder})
	if err != nil {
		return Order{}, err
	}
	if len(orders) == 0 {
		return Order{}, ErrOrderNotFound
	}
	return orders[0], nil
}

func loadOrdersWithItems(ctx context.Context, queries *database.Queries, rows []database.Order) ([]Order, error) {
	if len(rows) == 0 {
		return []Order{}, nil
	}

	ids := make([]uuid.UUID, len(rows))
	results := make([]Order, len(rows))
	orderIndex := make(map[uuid.UUID]int, len(rows))

	for i, row := range rows {
		ids[i] = row.ID
		orderIndex[row.ID] = i
		results[i] = mapDBOrder(row)
		results[i].Items = []OrderItem{}
	}

	dbItems, err := queries.ListOrderItemsByOrderIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	for _, dbItem := range dbItems {
		if idx, ok := orderIndex[dbItem.OrderID]; ok {
			results[idx].Items = append(results[idx].Items, mapDBOrderItem(dbItem))
		}
	}

	return results, nil
}
