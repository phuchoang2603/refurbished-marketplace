package cartclient

import (
	"context"

	cartv1 "refurbished-marketplace/shared/proto/cart/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn   *grpc.ClientConn
	client cartv1.CartServiceClient
}

func New(addr string) (*Client, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &Client{conn: conn, client: cartv1.NewCartServiceClient(conn)}, nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) GetCart(ctx context.Context, cartID string) (*cartv1.Cart, error) {
	return c.client.GetCart(ctx, &cartv1.GetCartRequest{CartId: cartID})
}

func (c *Client) AddCartItem(ctx context.Context, cartID, productID string, quantity int32) (*cartv1.Cart, error) {
	return c.client.AddCartItem(ctx, &cartv1.AddCartItemRequest{CartId: cartID, ProductId: productID, Quantity: quantity})
}

func (c *Client) SetCartItemQuantity(ctx context.Context, cartID, productID string, quantity int32) (*cartv1.Cart, error) {
	return c.client.SetCartItemQuantity(ctx, &cartv1.SetCartItemQuantityRequest{CartId: cartID, ProductId: productID, Quantity: quantity})
}

func (c *Client) RemoveCartItem(ctx context.Context, cartID, productID string) (*cartv1.Cart, error) {
	return c.client.RemoveCartItem(ctx, &cartv1.RemoveCartItemRequest{CartId: cartID, ProductId: productID})
}

func (c *Client) ClearCart(ctx context.Context, cartID string) error {
	_, err := c.client.ClearCart(ctx, &cartv1.ClearCartRequest{CartId: cartID})
	return err
}
