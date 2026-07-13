package service

import (
	"context"
	"fmt"

	"refurbished-marketplace/services/products/internal/database"
	"refurbished-marketplace/shared/messaging"
	ordersv1 "refurbished-marketplace/shared/proto/orders/v1"
	productsv1 "refurbished-marketplace/shared/proto/products/v1"
	sharedtrace "refurbished-marketplace/shared/trace"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

type ReservationItemInput struct {
	ProductID uuid.UUID
	Quantity  int32
}

func parseOrderCreatedReservation(msg *ordersv1.OrderCreated) (uuid.UUID, uuid.UUID, int64, []ReservationItemInput, error) {
	orderID, err := uuid.Parse(msg.GetOrderId())
	if err != nil {
		return uuid.Nil, uuid.Nil, 0, nil, fmt.Errorf("order_id: %w", err)
	}
	merchantID, err := uuid.Parse(msg.GetMerchantId())
	if err != nil {
		return uuid.Nil, uuid.Nil, 0, nil, fmt.Errorf("merchant_id: %w", err)
	}
	if msg.GetTotalCents() <= 0 {
		return uuid.Nil, uuid.Nil, 0, nil, fmt.Errorf("total_cents: invalid")
	}
	if len(msg.GetItems()) == 0 {
		return uuid.Nil, uuid.Nil, 0, nil, fmt.Errorf("items: missing")
	}

	aggregated := make(map[uuid.UUID]int32, len(msg.GetItems()))
	for _, item := range msg.GetItems() {
		productID, err := uuid.Parse(item.GetProductId())
		if err != nil {
			return uuid.Nil, uuid.Nil, 0, nil, fmt.Errorf("product_id: %w", err)
		}
		if err := validatePositiveQuantity(item.GetQuantity()); err != nil {
			return uuid.Nil, uuid.Nil, 0, nil, err
		}
		aggregated[productID] += item.GetQuantity()
	}

	items := make([]ReservationItemInput, 0, len(aggregated))
	for productID, quantity := range aggregated {
		items = append(items, ReservationItemInput{ProductID: productID, Quantity: quantity})
	}

	return orderID, merchantID, msg.GetTotalCents(), items, nil
}

func reserveOrderItems(ctx context.Context, q *database.Queries, orderID uuid.UUID, items []ReservationItemInput) error {
	productIDs := make([]uuid.UUID, 0, len(items))
	for _, item := range items {
		productIDs = append(productIDs, item.ProductID)
	}

	inventories, err := q.GetInventoriesByProductIDsForUpdate(ctx, productIDs)
	if err != nil {
		return err
	}

	inventoryByProductID := make(map[uuid.UUID]database.Inventory, len(inventories))
	for _, inv := range inventories {
		inventoryByProductID[inv.ProductID] = inv
	}

	for _, item := range items {
		inv, ok := inventoryByProductID[item.ProductID]
		if !ok {
			return ErrInventoryNotFound
		}
		if inv.AvailableQty < item.Quantity {
			return ErrInsufficientStock
		}
	}

	for _, item := range items {
		if _, err := q.ReserveInventoryStock(ctx, database.ReserveInventoryStockParams{ProductID: item.ProductID, Quantity: item.Quantity}); err != nil {
			return err
		}
		if _, err := q.CreateInventoryReservation(ctx, database.CreateInventoryReservationParams{OrderID: orderID, ProductID: item.ProductID, Quantity: item.Quantity, Status: ReservationStatusReserved}); err != nil {
			return err
		}
	}

	return nil
}

func createInventoryReservedOutbox(ctx context.Context, q *database.Queries, orderID, merchantID uuid.UUID, totalCents int64) error {
	payload, err := proto.Marshal(&productsv1.InventoryReserved{OrderId: orderID.String(), MerchantId: merchantID.String(), TotalCents: totalCents})
	if err != nil {
		return err
	}

	_, err = q.CreateInventoryOutbox(ctx, database.CreateInventoryOutboxParams{
		ID: uuid.New(), AggregateID: orderID, EventType: messaging.EventTypeInventoryReserved, Payload: payload,
		Tracingspancontext: sharedtrace.SerializeContext(ctx),
	})
	return err
}

func createInventoryReservationFailedOutbox(ctx context.Context, q *database.Queries, orderID uuid.UUID) error {
	payload, err := proto.Marshal(&productsv1.InventoryReservationFailed{OrderId: orderID.String()})
	if err != nil {
		return err
	}

	_, err = q.CreateInventoryOutbox(ctx, database.CreateInventoryOutboxParams{
		ID: uuid.New(), AggregateID: orderID, EventType: messaging.EventTypeInventoryReservationFailed, Payload: payload,
		Tracingspancontext: sharedtrace.SerializeContext(ctx),
	})
	return err
}
