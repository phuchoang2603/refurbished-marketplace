package clients

import (
	"context"

	productsv1 "refurbished-marketplace/shared/proto/products/v1"

	"google.golang.org/grpc"
)

type ProductsClient struct {
	conn   *grpc.ClientConn
	client productsv1.ProductsServiceClient
}

func newProductsClient(addr string) (*ProductsClient, error) {
	conn, err := newConn(addr)
	if err != nil {
		return nil, err
	}
	return &ProductsClient{conn: conn, client: productsv1.NewProductsServiceClient(conn)}, nil
}

func (c *ProductsClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *ProductsClient) CreateProduct(ctx context.Context, name, description string, priceCents int64, merchantID string, initialStock int32) (*productsv1.Product, error) {
	return c.client.CreateProduct(ctx, &productsv1.CreateProductRequest{
		Name:         name,
		Description:  description,
		PriceCents:   priceCents,
		MerchantId:   merchantID,
		InitialStock: &initialStock,
	})
}

func (c *ProductsClient) GetProductByID(ctx context.Context, id string) (*productsv1.Product, error) {
	return c.client.GetProductByID(ctx, &productsv1.GetProductByIDRequest{Id: id})
}

func (c *ProductsClient) ListProducts(ctx context.Context, limit, offset int32) (*productsv1.ListProductsResponse, error) {
	return c.client.ListProducts(ctx, &productsv1.ListProductsRequest{Limit: limit, Offset: offset})
}
