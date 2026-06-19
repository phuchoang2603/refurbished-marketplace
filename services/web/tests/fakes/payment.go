package fakes

import (
	"context"

	paymentv1 "refurbished-marketplace/shared/proto/payment/v1"
)

type PaymentService struct {
	CreateSessionFn func(context.Context, *paymentv1.CreateHostedPaymentSessionRequest) (*paymentv1.CreateHostedPaymentSessionResponse, error)
	GetSessionFn    func(context.Context, string) (*paymentv1.HostedPaymentSession, error)
	HandleWebhookFn func(context.Context, *paymentv1.HandleGatewayWebhookRequest) (*paymentv1.HandleGatewayWebhookResponse, error)
}

func (f *PaymentService) CreateHostedPaymentSession(ctx context.Context, req *paymentv1.CreateHostedPaymentSessionRequest) (*paymentv1.CreateHostedPaymentSessionResponse, error) {
	if f.CreateSessionFn != nil {
		return f.CreateSessionFn(ctx, req)
	}
	return &paymentv1.CreateHostedPaymentSessionResponse{}, nil
}

func (f *PaymentService) GetHostedPaymentSessionByOrder(ctx context.Context, orderID string) (*paymentv1.HostedPaymentSession, error) {
	if f.GetSessionFn != nil {
		return f.GetSessionFn(ctx, orderID)
	}
	return nil, nil
}

func (f *PaymentService) HandleGatewayWebhook(ctx context.Context, req *paymentv1.HandleGatewayWebhookRequest) (*paymentv1.HandleGatewayWebhookResponse, error) {
	if f.HandleWebhookFn != nil {
		return f.HandleWebhookFn(ctx, req)
	}
	return &paymentv1.HandleGatewayWebhookResponse{}, nil
}
