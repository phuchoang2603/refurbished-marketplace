package clients

import (
	"context"

	sharedtrace "github.com/phuchoang2603/refurbished-marketplace/shared/observe/trace"
	productsv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/products/v1"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ReserveItem struct {
	ProductID uuid.UUID
	Quantity  int32
}

type Products struct {
	conn   *grpc.ClientConn
	client productsv1.ProductsServiceClient
}

func NewProducts(addr string) (*Products, error) {
	opts := append([]grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}, sharedtrace.GRPCDialOptions()...)
	conn, err := grpc.NewClient(addr, opts...)
	if err != nil {
		return nil, err
	}
	conn.Connect()
	return &Products{conn: conn, client: productsv1.NewProductsServiceClient(conn)}, nil
}

func (c *Products) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func (c *Products) ReserveStock(ctx context.Context, orderID, merchantID uuid.UUID, totalCents int64, items []ReserveItem) error {
	pbItems := make([]*productsv1.ReserveStockItem, 0, len(items))
	for _, item := range items {
		pbItems = append(pbItems, &productsv1.ReserveStockItem{
			ProductId: item.ProductID.String(),
			Quantity:  item.Quantity,
		})
	}
	_, err := c.client.ReserveStock(ctx, &productsv1.ReserveStockRequest{
		OrderId:    orderID.String(),
		MerchantId: merchantID.String(),
		TotalCents: totalCents,
		Items:      pbItems,
	})
	return err
}
