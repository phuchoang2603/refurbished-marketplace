package clients

import (
	"context"

	searchv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/search/v1"

	"google.golang.org/grpc"
)

type SearchClient struct {
	conn   *grpc.ClientConn
	client searchv1.SearchServiceClient
}

func newSearchClient(addr string) (*SearchClient, error) {
	conn, err := newConn(addr)
	if err != nil {
		return nil, err
	}
	return &SearchClient{conn: conn, client: searchv1.NewSearchServiceClient(conn)}, nil
}

func (c *SearchClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *SearchClient) SearchProducts(ctx context.Context, query, merchantID string, limit, offset int32) (*searchv1.SearchProductsResponse, error) {
	return c.client.SearchProducts(ctx, &searchv1.SearchProductsRequest{
		Query:      query,
		MerchantId: merchantID,
		Limit:      limit,
		Offset:     offset,
	})
}
