package ordersclient

import (
	"context"

	ordersv1 "refurbished-marketplace/shared/proto/orders/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn   *grpc.ClientConn
	client ordersv1.OrdersServiceClient
}

func New(addr string) (*Client, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &Client{conn: conn, client: ordersv1.NewOrdersServiceClient(conn)}, nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) CreateOrder(ctx context.Context, buyerUserID, productID string, quantity int32, totalCents int64) (*ordersv1.Order, error) {
	return c.client.CreateOrder(ctx, &ordersv1.CreateOrderRequest{
		BuyerUserId: buyerUserID,
		ProductId:   productID,
		Quantity:    quantity,
		TotalCents:  totalCents,
	})
}

func (c *Client) GetOrderByID(ctx context.Context, id string) (*ordersv1.Order, error) {
	return c.client.GetOrderByID(ctx, &ordersv1.GetOrderByIDRequest{Id: id})
}

func (c *Client) ListOrdersByBuyer(ctx context.Context, buyerUserID string, limit, offset int32) (*ordersv1.ListOrdersByBuyerResponse, error) {
	return c.client.ListOrdersByBuyer(ctx, &ordersv1.ListOrdersByBuyerRequest{BuyerUserId: buyerUserID, Limit: limit, Offset: offset})
}

func (c *Client) UpdateOrderStatus(ctx context.Context, id, status string) (*ordersv1.Order, error) {
	return c.client.UpdateOrderStatus(ctx, &ordersv1.UpdateOrderStatusRequest{Id: id, Status: status})
}
