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

type UpdateProductInput struct {
	Name        *string
	Description *string
	PriceCents  *int64
	Stock       *int32
}

func (c *Client) CreateProduct(ctx context.Context, ownerUserID, name, description string, priceCents int64, stock int32) (*productsv1.Product, error) {
	return c.client.CreateProduct(ctx, &productsv1.CreateProductRequest{
		OwnerUserId: ownerUserID,
		Name:        name,
		Description: description,
		PriceCents:  priceCents,
		Stock:       stock,
	})
}

func (c *Client) GetProductByID(ctx context.Context, id string) (*productsv1.Product, error) {
	return c.client.GetProductByID(ctx, &productsv1.GetProductByIDRequest{Id: id})
}

func (c *Client) ListProducts(ctx context.Context, limit, offset int32) (*productsv1.ListProductsResponse, error) {
	return c.client.ListProducts(ctx, &productsv1.ListProductsRequest{Limit: limit, Offset: offset})
}

func (c *Client) UpdateProduct(ctx context.Context, id, ownerUserID string, in UpdateProductInput) (*productsv1.Product, error) {
	req := &productsv1.UpdateProductRequest{Id: id, OwnerUserId: ownerUserID}
	if in.Name != nil {
		req.Name = in.Name
	}
	if in.Description != nil {
		req.Description = in.Description
	}
	if in.PriceCents != nil {
		req.PriceCents = in.PriceCents
	}
	if in.Stock != nil {
		req.Stock = in.Stock
	}
	return c.client.UpdateProduct(ctx, req)
}

func (c *Client) DeleteProduct(ctx context.Context, id, ownerUserID string) (*productsv1.DeleteProductResponse, error) {
	return c.client.DeleteProduct(ctx, &productsv1.DeleteProductRequest{Id: id, OwnerUserId: ownerUserID})
}
