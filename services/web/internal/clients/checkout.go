package clients

import (
	"context"

	checkoutv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/checkout/v1"
	"google.golang.org/grpc"
)

type CheckoutClient struct {
	conn   *grpc.ClientConn
	client checkoutv1.CheckoutServiceClient
}

func newCheckoutClient(addr string) (*CheckoutClient, error) {
	conn, err := newConn(addr)
	if err != nil {
		return nil, err
	}
	return &CheckoutClient{conn: conn, client: checkoutv1.NewCheckoutServiceClient(conn)}, nil
}

func (client *CheckoutClient) Close() error {
	return client.conn.Close()
}

func (client *CheckoutClient) SubmitCheckout(ctx context.Context, request *checkoutv1.SubmitCheckoutRequest) (*checkoutv1.SubmitCheckoutResponse, error) {
	return client.client.SubmitCheckout(ctx, request)
}

func (client *CheckoutClient) GetCheckout(ctx context.Context, request *checkoutv1.GetCheckoutRequest) (*checkoutv1.CheckoutStatus, error) {
	return client.client.GetCheckout(ctx, request)
}
