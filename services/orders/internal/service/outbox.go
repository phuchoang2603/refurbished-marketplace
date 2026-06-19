package service

import (
	"context"

	"refurbished-marketplace/services/orders/internal/database"
	"refurbished-marketplace/shared/messaging"

	ordersv1 "refurbished-marketplace/shared/proto/orders/v1"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

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
