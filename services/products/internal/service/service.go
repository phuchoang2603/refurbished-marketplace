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
	UpdateProductByIDAndOwner(ctx context.Context, arg database.UpdateProductByIDAndOwnerParams) (database.Product, error)
	DeleteProductByIDAndOwner(ctx context.Context, arg database.DeleteProductByIDAndOwnerParams) (int64, error)
}

var (
	ErrInvalidProductName = errors.New("invalid product name")
	ErrInvalidPrice       = errors.New("invalid product price")
	ErrInvalidStock       = errors.New("invalid product stock")
	ErrProductNotFound    = errors.New("product not found")
	ErrForbiddenProduct   = errors.New("forbidden product access")
	ErrInvalidOwner       = errors.New("invalid owner user id")
	ErrInvalidListLimit   = errors.New("invalid list limit")
	ErrInvalidListOffset  = errors.New("invalid list offset")
)

type UpdateProductInput struct {
	Name        *string
	Description *string
	PriceCents  *int64
	Stock       *int32
}

type Service struct {
	queries queryStore
}

func New(queries queryStore) *Service {
	return &Service{queries: queries}
}
