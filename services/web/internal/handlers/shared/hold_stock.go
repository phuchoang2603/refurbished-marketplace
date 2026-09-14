package shared

import (
	"context"

	sharedlog "github.com/phuchoang2603/refurbished-marketplace/shared/observe/log"
	inventoryv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/inventory/v1"
	ordersv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/orders/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func HoldStockForOrder(ctx context.Context, inventory InventoryService, orders OrdersService, order *ordersv1.Order) error {
	if order == nil || order.GetStatus() != ordersv1.OrderStatus_ORDER_STATUS_PENDING {
		return nil
	}
	if inventory == nil {
		return status.Error(codes.Unavailable, "inventory unavailable")
	}

	items := make([]*inventoryv1.ReserveStockItem, 0, len(order.GetItems()))
	for _, item := range order.GetItems() {
		items = append(items, &inventoryv1.ReserveStockItem{
			ProductId: item.GetProductId(),
			Quantity:  item.GetQuantity(),
		})
	}
	if len(items) == 0 {
		return status.Error(codes.InvalidArgument, "order has no items")
	}

	err := inventory.ReserveStock(ctx, order.GetId(), order.GetMerchantId(), order.GetTotalCents(), items)
	if err == nil {
		return nil
	}
	if orders != nil {
		if _, statusErr := orders.UpdateOrderStatus(ctx, order.GetId(), ordersv1.OrderStatus_ORDER_STATUS_FAILED); statusErr != nil {
			sharedlog.ErrorContext(ctx, "failed to mark unreserved order failed", sharedlog.KeyOrderID, order.GetId(), sharedlog.KeyErr, statusErr)
		}
	}
	return err
}
