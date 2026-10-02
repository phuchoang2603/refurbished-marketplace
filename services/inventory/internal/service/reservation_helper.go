package service

import (
	"context"

	"github.com/phuchoang2603/refurbished-marketplace/services/inventory/internal/database"
	"github.com/phuchoang2603/refurbished-marketplace/shared/messaging"
	sharedtrace "github.com/phuchoang2603/refurbished-marketplace/shared/observe/trace"
	productsv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/products/v1"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

type ReservationItemInput struct {
	ProductID uuid.UUID
	Quantity  int32
}

func commandReserveInboxID(orderID uuid.UUID) string {
	return "inventory.reserve-command/" + orderID.String()
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
