package productsclient

import (
	"context"

	productsv1 "refurbished-marketplace/shared/proto/products/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn   *grpc.ClientConn
	client productsv1.ProductsServiceClient
}

func New(addr string) (*Client, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &Client{conn: conn, client: productsv1.NewProductsServiceClient(conn)}, nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) CreateProduct(ctx context.Context, name, description string, priceCents int64, stock int32, terminalID string, xPos, yPos float64) (*productsv1.Product, error) {
	return c.client.CreateProduct(ctx, &productsv1.CreateProductRequest{
		Name:        name,
		Description: description,
		PriceCents:  priceCents,
		Stock:       stock,
		TerminalId:  terminalID,
		XPos:        xPos,
		YPos:        yPos,
	})
}

func (c *Client) GetProductByID(ctx context.Context, id string) (*productsv1.Product, error) {
	return c.client.GetProductByID(ctx, &productsv1.GetProductByIDRequest{Id: id})
}

func (c *Client) ListProducts(ctx context.Context, limit, offset int32) (*productsv1.ListProductsResponse, error) {
	return c.client.ListProducts(ctx, &productsv1.ListProductsRequest{Limit: limit, Offset: offset})
}
