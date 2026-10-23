package fakes

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	checkoutv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/checkout/v1"
)

type CheckoutService struct {
	WebhookFn func(context.Context, *checkoutv1.HandleGatewayWebhookRequest) (*checkoutv1.HandleGatewayWebhookResponse, error)
	SubmitFn  func(context.Context, *checkoutv1.SubmitCheckoutRequest) (*checkoutv1.SubmitCheckoutResponse, error)
	GetFn     func(context.Context, *checkoutv1.GetCheckoutRequest) (*checkoutv1.CheckoutStatus, error)
}

func (service *CheckoutService) SubmitCheckout(ctx context.Context, request *checkoutv1.SubmitCheckoutRequest) (*checkoutv1.SubmitCheckoutResponse, error) {
	return service.SubmitFn(ctx, request)
}

func (service *CheckoutService) GetCheckout(ctx context.Context, request *checkoutv1.GetCheckoutRequest) (*checkoutv1.CheckoutStatus, error) {
	if service.GetFn == nil {
		return nil, status.Error(codes.NotFound, "checkout not found")
	}
	return service.GetFn(ctx, request)
}

func (service *CheckoutService) HandleGatewayWebhook(ctx context.Context, request *checkoutv1.HandleGatewayWebhookRequest) (*checkoutv1.HandleGatewayWebhookResponse, error) {
	return service.WebhookFn(ctx, request)
}
