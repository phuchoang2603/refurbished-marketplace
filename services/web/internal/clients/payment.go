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

func (c *PaymentClient) InitiatePayment(ctx context.Context, req *paymentv1.InitiatePaymentRequest) (*paymentv1.InitiatePaymentResponse, error) {
	return c.client.InitiatePayment(ctx, req)
}

func (c *PaymentClient) HandleGatewayWebhook(ctx context.Context, req *paymentv1.HandleGatewayWebhookRequest) (*paymentv1.HandleGatewayWebhookResponse, error) {
	return c.client.HandleGatewayWebhook(ctx, req)
}

func (c *PaymentClient) GetTransaction(ctx context.Context, paymentTransactionID string) (*paymentv1.PaymentTransaction, error) {
	return c.client.GetTransaction(ctx, &paymentv1.GetTransactionRequest{PaymentTransactionId: paymentTransactionID})
}
