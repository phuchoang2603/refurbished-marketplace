package clients

import (
	"context"

	cartv1 "refurbished-marketplace/shared/proto/cart/v1"

	"google.golang.org/grpc"
)

type CartClient struct {
	conn   *grpc.ClientConn
	client cartv1.CartServiceClient
}

func newCartClient(addr string) (*CartClient, error) {
	conn, err := newConn(addr)
	if err != nil {
		return nil, err
	}
	return &CartClient{conn: conn, client: cartv1.NewCartServiceClient(conn)}, nil
}

func (c *CartClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *CartClient) GetCart(ctx context.Context, cartID string) (*cartv1.Cart, error) {
	return c.client.GetCart(ctx, &cartv1.GetCartRequest{CartId: cartID})
}

func (c *CartClient) AddCartItem(ctx context.Context, cartID, productID string, quantity int32) (*cartv1.Cart, error) {
	return c.client.AddCartItem(ctx, &cartv1.AddCartItemRequest{CartId: cartID, ProductId: productID, Quantity: quantity})
}

func (c *CartClient) SetCartItemQuantity(ctx context.Context, cartID, productID string, quantity int32) (*cartv1.Cart, error) {
	return c.client.SetCartItemQuantity(ctx, &cartv1.SetCartItemQuantityRequest{CartId: cartID, ProductId: productID, Quantity: quantity})
}

func (c *CartClient) RemoveCartItem(ctx context.Context, cartID, productID string) (*cartv1.Cart, error) {
	return c.client.RemoveCartItem(ctx, &cartv1.RemoveCartItemRequest{CartId: cartID, ProductId: productID})
}

func (c *CartClient) ClearCart(ctx context.Context, cartID string) error {
	_, err := c.client.ClearCart(ctx, &cartv1.ClearCartRequest{CartId: cartID})
	return err
}
