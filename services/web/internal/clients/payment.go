package clients

import (
	"context"

	paymentv1 "refurbished-marketplace/shared/proto/payment/v1"

	"google.golang.org/grpc"
)

type PaymentClient struct {
	conn   *grpc.ClientConn
	client paymentv1.PaymentServiceClient
}

func newPaymentClient(addr string) (*PaymentClient, error) {
	conn, err := newConn(addr)
	if err != nil {
		return nil, err
	}
	return &PaymentClient{conn: conn, client: paymentv1.NewPaymentServiceClient(conn)}, nil
}

func (c *PaymentClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *PaymentClient) CreateHostedPaymentSession(ctx context.Context, req *paymentv1.CreateHostedPaymentSessionRequest) (*paymentv1.CreateHostedPaymentSessionResponse, error) {
	return c.client.CreateHostedPaymentSession(ctx, req)
}

func (c *PaymentClient) GetHostedPaymentSessionByOrder(ctx context.Context, orderID string) (*paymentv1.HostedPaymentSession, error) {
	return c.client.GetHostedPaymentSessionByOrder(ctx, &paymentv1.GetHostedPaymentSessionByOrderRequest{OrderId: orderID})
}

func (c *PaymentClient) HandleGatewayWebhook(ctx context.Context, req *paymentv1.HandleGatewayWebhookRequest) (*paymentv1.HandleGatewayWebhookResponse, error) {
	return c.client.HandleGatewayWebhook(ctx, req)
}
