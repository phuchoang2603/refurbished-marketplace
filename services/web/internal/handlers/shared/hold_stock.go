package shared

import (
	"context"

	sharedlog "github.com/phuchoang2603/refurbished-marketplace/shared/observe/log"
	ordersv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/orders/v1"
	productsv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/products/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func HoldStockForOrder(ctx context.Context, products ProductsService, orders OrdersService, order *ordersv1.Order) error {
	if order == nil || order.GetStatus() != ordersv1.OrderStatus_ORDER_STATUS_PENDING {
		return nil
	}
	if products == nil {
		return status.Error(codes.Unavailable, "products unavailable")
	}

	items := make([]*productsv1.ReserveStockItem, 0, len(order.GetItems()))
	for _, item := range order.GetItems() {
		items = append(items, &productsv1.ReserveStockItem{
			ProductId: item.GetProductId(),
			Quantity:  item.GetQuantity(),
		})
	}
	if len(items) == 0 {
		return status.Error(codes.InvalidArgument, "order has no items")
	}

	err := products.ReserveStock(ctx, order.GetId(), order.GetMerchantId(), order.GetTotalCents(), items)
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
