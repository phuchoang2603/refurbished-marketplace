package service

import (
	"context"
	"strings"

	"refurbished-marketplace/services/orders/internal/database"
	"refurbished-marketplace/shared/messaging"

	ordersv1 "refurbished-marketplace/shared/proto/orders/v1"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

func mapDBOrder(i database.Order) Order {
	return Order{
		ID:          i.ID,
		BuyerUserID: i.BuyerUserID,
		MerchantID:  i.MerchantID,
		Status:      i.Status,
		TotalCents:  i.TotalCents,
		CreatedAt:   i.CreatedAt,
		UpdatedAt:   i.UpdatedAt,
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

func validateCreateOrderInput(buyerUserID, merchantID uuid.UUID, items []OrderItemInput, totalCents int64) error {
	if buyerUserID == uuid.Nil {
		return ErrInvalidBuyerID
	}
	if merchantID == uuid.Nil {
		return ErrInvalidMerchantID
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

func createOrderOutbox(ctx context.Context, queries *database.Queries, order Order) error {
	items := make([]*ordersv1.OrderCreatedItem, 0, len(order.Items))
	for _, item := range order.Items {
		items = append(items, &ordersv1.OrderCreatedItem{
			ProductId: item.ProductID.String(),
			Quantity:  item.Quantity,
		})
	}

	payloadBytes, err := proto.Marshal(&ordersv1.OrderCreated{
		OrderId:     order.ID.String(),
		BuyerUserId: order.BuyerUserID.String(),
		MerchantId:  order.MerchantID.String(),
		TotalCents:  order.TotalCents,
		Items:       items,
	})
	if err != nil {
		return err
	}

	_, err = queries.CreateOrderOutbox(ctx, database.CreateOrderOutboxParams{
		ID:          uuid.New(),
		AggregateID: order.ID,
		EventType:   messaging.EventTypeOrderCreated,
		Payload:     payloadBytes,
	})
	return err
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
