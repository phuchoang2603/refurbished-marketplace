package service

import (
	"context"

	"refurbished-marketplace/services/orders/internal/database"

	"github.com/google/uuid"
)

func createOrderItems(ctx context.Context, queries *database.Queries, orderID uuid.UUID, items []OrderItemInput) ([]OrderItem, error) {
	orderItems := make([]OrderItem, 0, len(items))

	for _, item := range items {
		itemID := uuid.New()
		lineTotal := item.UnitPriceCents * int64(item.Quantity)
		createdItem, err := queries.CreateOrderItem(ctx, database.CreateOrderItemParams{
			ID:             itemID,
			OrderID:        orderID,
			ProductID:      item.ProductID,
			Quantity:       item.Quantity,
			UnitPriceCents: item.UnitPriceCents,
			LineTotalCents: lineTotal,
		})
		if err != nil {
			return nil, err
		}

		orderItems = append(orderItems, mapDBOrderItem(createdItem))
	}

	return orderItems, nil
}

func loadOrdersWithItems(ctx context.Context, queries *database.Queries, orders []Order) ([]Order, error) {
	if len(orders) == 0 {
		return []Order{}, nil
	}

	ids := make([]uuid.UUID, len(orders))
	results := make([]Order, len(orders))
	orderIndex := make(map[uuid.UUID]int, len(orders))

	for i, order := range orders {
		ids[i] = order.ID
		orderIndex[order.ID] = i
		results[i] = order
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
