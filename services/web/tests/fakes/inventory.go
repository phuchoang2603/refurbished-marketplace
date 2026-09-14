package fakes

import (
	"context"

	inventoryv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/inventory/v1"
)

type InventoryService struct {
	GetFn      func(context.Context, string) (*inventoryv1.Stock, error)
	GetByIDsFn func(context.Context, []string) (*inventoryv1.GetStocksByIDsResponse, error)
	ReserveFn  func(context.Context, string, string, int64, []*inventoryv1.ReserveStockItem) error
}

func (f *InventoryService) GetStock(ctx context.Context, productID string) (*inventoryv1.Stock, error) {
	if f.GetFn != nil {
		return f.GetFn(ctx, productID)
	}
	return &inventoryv1.Stock{ProductId: productID}, nil
}

func (f *InventoryService) GetStocksByIDs(ctx context.Context, productIDs []string) (*inventoryv1.GetStocksByIDsResponse, error) {
	if f.GetByIDsFn != nil {
		return f.GetByIDsFn(ctx, productIDs)
	}
	return &inventoryv1.GetStocksByIDsResponse{}, nil
}

func (f *InventoryService) ReserveStock(ctx context.Context, orderID, merchantID string, totalCents int64, items []*inventoryv1.ReserveStockItem) error {
	if f.ReserveFn != nil {
		return f.ReserveFn(ctx, orderID, merchantID, totalCents, items)
	}
	return nil
}
