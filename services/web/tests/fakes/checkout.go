package fakes

import (
	"context"

	checkoutv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/checkout/v1"
)

type CheckoutService struct {
	SubmitFn func(context.Context, *checkoutv1.SubmitCheckoutRequest) (*checkoutv1.SubmitCheckoutResponse, error)
	GetFn    func(context.Context, *checkoutv1.GetCheckoutRequest) (*checkoutv1.CheckoutStatus, error)
}

func (service *CheckoutService) SubmitCheckout(ctx context.Context, request *checkoutv1.SubmitCheckoutRequest) (*checkoutv1.SubmitCheckoutResponse, error) {
	return service.SubmitFn(ctx, request)
}

func (service *CheckoutService) GetCheckout(ctx context.Context, request *checkoutv1.GetCheckoutRequest) (*checkoutv1.CheckoutStatus, error) {
	return service.GetFn(ctx, request)
}
