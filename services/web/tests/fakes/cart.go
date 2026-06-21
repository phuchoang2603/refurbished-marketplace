package fakes

import (
	"context"

	cartv1 "refurbished-marketplace/shared/proto/cart/v1"
)

type CartService struct {
	GetFn       func(context.Context, string) (*cartv1.Cart, error)
	AddFn       func(context.Context, string, string, string, int32) (*cartv1.Cart, error)
	SetQtyFn    func(context.Context, string, string, string, int32) (*cartv1.Cart, error)
	RemoveFn    func(context.Context, string, string) (*cartv1.Cart, error)
	ClearCartFn func(context.Context, string) error
}

func (f *CartService) GetCart(ctx context.Context, cartID string) (*cartv1.Cart, error) {
	if f.GetFn != nil {
		return f.GetFn(ctx, cartID)
	}
	return &cartv1.Cart{}, nil
}

func (f *CartService) AddCartItem(ctx context.Context, cartID, productID, merchantID string, quantity int32) (*cartv1.Cart, error) {
	if f.AddFn != nil {
		return f.AddFn(ctx, cartID, productID, merchantID, quantity)
	}
	return &cartv1.Cart{}, nil
}

func (f *CartService) SetCartItemQuantity(ctx context.Context, cartID, productID, merchantID string, quantity int32) (*cartv1.Cart, error) {
	if f.SetQtyFn != nil {
		return f.SetQtyFn(ctx, cartID, productID, merchantID, quantity)
	}
	return &cartv1.Cart{}, nil
}

func (f *CartService) RemoveCartItem(ctx context.Context, cartID, productID string) (*cartv1.Cart, error) {
	if f.RemoveFn != nil {
		return f.RemoveFn(ctx, cartID, productID)
	}
	return &cartv1.Cart{}, nil
}

func (f *CartService) ClearCart(ctx context.Context, cartID string) error {
	if f.ClearCartFn != nil {
		return f.ClearCartFn(ctx, cartID)
	}
	return nil
}
