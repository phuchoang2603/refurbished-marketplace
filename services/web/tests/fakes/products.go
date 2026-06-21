package fakes

import (
	"context"

	productsv1 "refurbished-marketplace/shared/proto/products/v1"
)

type ProductsService struct {
	CreateFn  func(context.Context, string, string, int64, string, int32) (*productsv1.Product, error)
	GetByIDFn func(context.Context, string) (*productsv1.Product, error)
	ListFn    func(context.Context, int32, int32) (*productsv1.ListProductsResponse, error)
}

func (f *ProductsService) CreateProduct(ctx context.Context, name, description string, priceCents int64, merchantID string, initialStock int32) (*productsv1.Product, error) {
	if f.CreateFn != nil {
		return f.CreateFn(ctx, name, description, priceCents, merchantID, initialStock)
	}
	return &productsv1.Product{}, nil
}

func (f *ProductsService) GetProductByID(ctx context.Context, id string) (*productsv1.Product, error) {
	if f.GetByIDFn != nil {
		return f.GetByIDFn(ctx, id)
	}
	return nil, nil
}

func (f *ProductsService) ListProducts(ctx context.Context, limit, offset int32) (*productsv1.ListProductsResponse, error) {
	if f.ListFn != nil {
		return f.ListFn(ctx, limit, offset)
	}
	return &productsv1.ListProductsResponse{}, nil
}
