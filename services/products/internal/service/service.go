package service

import (
	"context"
	"errors"

	"refurbished-marketplace/services/products/internal/database"

	"github.com/google/uuid"
)

type queryStore interface {
	CreateProduct(ctx context.Context, arg database.CreateProductParams) (database.Product, error)
	GetProductByID(ctx context.Context, id uuid.UUID) (database.Product, error)
	ListProducts(ctx context.Context, arg database.ListProductsParams) ([]database.Product, error)
}

var (
	ErrInvalidProductName = errors.New("invalid product name")
	ErrInvalidPrice       = errors.New("invalid product price")
	ErrInvalidStock       = errors.New("invalid product stock")
	ErrProductNotFound    = errors.New("product not found")
	ErrInvalidListLimit   = errors.New("invalid list limit")
	ErrInvalidListOffset  = errors.New("invalid list offset")
)

type Service struct {
	queries queryStore
}

func New(queries queryStore) *Service {
	return &Service{queries: queries}
}
