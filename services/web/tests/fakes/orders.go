package fakes

import (
	"context"

	ordersv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/orders/v1"
)

type OrdersService struct {
	CreateFn       func(context.Context, string, string, []*ordersv1.CreateOrderItem, int64, string) (*ordersv1.Order, error)
	GetFn          func(context.Context, string) (*ordersv1.Order, error)
	ListFn         func(context.Context, string, int32, int32) (*ordersv1.ListOrdersByBuyerResponse, error)
	UpdateStatusFn func(context.Context, string, ordersv1.OrderStatus) (*ordersv1.Order, error)
}

func (f *OrdersService) CreateOrder(ctx context.Context, buyerUserID, merchantID string, items []*ordersv1.CreateOrderItem, totalCents int64, idempotencyKey string) (*ordersv1.Order, error) {
	if f.CreateFn != nil {
		return f.CreateFn(ctx, buyerUserID, merchantID, items, totalCents, idempotencyKey)
	}
	return nil, nil
}

func (f *OrdersService) GetOrderByID(ctx context.Context, id string) (*ordersv1.Order, error) {
	if f == nil {
		return nil, nil
	}
	if f.GetFn != nil {
		return f.GetFn(ctx, id)
	}
	return nil, nil
}

func (f *OrdersService) ListOrdersByBuyer(ctx context.Context, buyerUserID string, limit, offset int32) (*ordersv1.ListOrdersByBuyerResponse, error) {
	if f.ListFn != nil {
		return f.ListFn(ctx, buyerUserID, limit, offset)
	}
	return &ordersv1.ListOrdersByBuyerResponse{}, nil
}

func (f *OrdersService) UpdateOrderStatus(ctx context.Context, id string, status ordersv1.OrderStatus) (*ordersv1.Order, error) {
	if f.UpdateStatusFn != nil {
		return f.UpdateStatusFn(ctx, id, status)
	}
	return &ordersv1.Order{Id: id, Status: status}, nil
}
