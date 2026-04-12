package paymentclient

import (
	"context"

	paymentv1 "refurbished-marketplace/shared/proto/payment/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn   *grpc.ClientConn
	client paymentv1.PaymentServiceClient
}

func New(addr string) (*Client, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &Client{conn: conn, client: paymentv1.NewPaymentServiceClient(conn)}, nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) InitiatePayment(ctx context.Context, req *paymentv1.InitiatePaymentRequest) (*paymentv1.InitiatePaymentResponse, error) {
	return c.client.InitiatePayment(ctx, req)
}

func (c *Client) HandleGatewayWebhook(ctx context.Context, req *paymentv1.HandleGatewayWebhookRequest) (*paymentv1.HandleGatewayWebhookResponse, error) {
	return c.client.HandleGatewayWebhook(ctx, req)
}

func (c *Client) GetTransaction(ctx context.Context, paymentTransactionID string) (*paymentv1.PaymentTransaction, error) {
	return c.client.GetTransaction(ctx, &paymentv1.GetTransactionRequest{PaymentTransactionId: paymentTransactionID})
}

