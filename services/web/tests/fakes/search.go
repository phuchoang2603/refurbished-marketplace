package fakes

import (
	"context"

	searchv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/search/v1"
)

type SearchService struct {
	SearchFn func(context.Context, string, string, int32, int32) (*searchv1.SearchProductsResponse, error)
}

func (f *SearchService) SearchProducts(ctx context.Context, query, merchantID string, limit, offset int32) (*searchv1.SearchProductsResponse, error) {
	if f.SearchFn != nil {
		return f.SearchFn(ctx, query, merchantID, limit, offset)
	}
	return &searchv1.SearchProductsResponse{}, nil
}
