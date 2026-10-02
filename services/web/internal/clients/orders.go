package clients

import (
	"context"

	ordersv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/orders/v1"

	"google.golang.org/grpc"
)

type OrdersClient struct {
	conn   *grpc.ClientConn
	client ordersv1.OrdersServiceClient
}

func newOrdersClient(addr string) (*OrdersClient, error) {
	conn, err := newConn(addr)
	if err != nil {
		return nil, err
	}
	return &OrdersClient{conn: conn, client: ordersv1.NewOrdersServiceClient(conn)}, nil
}

func (c *OrdersClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *OrdersClient) GetOrderByID(ctx context.Context, id string) (*ordersv1.Order, error) {
	return c.client.GetOrderByID(ctx, &ordersv1.GetOrderByIDRequest{Id: id})
}

func (c *OrdersClient) ListOrdersByBuyer(ctx context.Context, buyerUserID string, limit, offset int32) (*ordersv1.ListOrdersByBuyerResponse, error) {
	return c.client.ListOrdersByBuyer(ctx, &ordersv1.ListOrdersByBuyerRequest{BuyerUserId: buyerUserID, Limit: limit, Offset: offset})
}
