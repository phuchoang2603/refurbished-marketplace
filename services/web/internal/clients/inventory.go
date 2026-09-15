package clients

import (
	"context"

	inventoryv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/inventory/v1"

	"google.golang.org/grpc"
)

type InventoryClient struct {
	conn   *grpc.ClientConn
	client inventoryv1.InventoryServiceClient
}

func newInventoryClient(addr string) (*InventoryClient, error) {
	conn, err := newConn(addr)
	if err != nil {
		return nil, err
	}
	return &InventoryClient{conn: conn, client: inventoryv1.NewInventoryServiceClient(conn)}, nil
}

func (c *InventoryClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *InventoryClient) GetStock(ctx context.Context, productID string) (*inventoryv1.Stock, error) {
	return c.client.GetStock(ctx, &inventoryv1.GetStockRequest{ProductId: productID})
}

func (c *InventoryClient) GetStocksByIDs(ctx context.Context, productIDs []string) (*inventoryv1.GetStocksByIDsResponse, error) {
	return c.client.GetStocksByIDs(ctx, &inventoryv1.GetStocksByIDsRequest{ProductIds: productIDs})
}

func (c *InventoryClient) ReserveStock(ctx context.Context, orderID, merchantID string, totalCents int64, items []*inventoryv1.ReserveStockItem) error {
	_, err := c.client.ReserveStock(ctx, &inventoryv1.ReserveStockRequest{
		OrderId:    orderID,
		MerchantId: merchantID,
		TotalCents: totalCents,
		Items:      items,
	})
	return err
}
