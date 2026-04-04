package tests

import (
	"errors"
	"testing"

	"refurbished-marketplace/services/products/internal/database"
	"refurbished-marketplace/services/products/internal/service"
	"refurbished-marketplace/shared/testutil"

	"github.com/google/uuid"
)

func newProductsService(t *testing.T) *service.Service {
	t.Helper()
	db := testutil.SetupPostgresWithMigrations(
		t,
		testutil.PostgresConfig{
			Database: "products_db",
			Username: "products_app",
			Password: "products_app_dev_password",
		},
		"../db/migrations",
	)

	return service.New(database.New(db))
}

func TestCreateAndReadProducts(t *testing.T) {
	svc := newProductsService(t)
	ctx := t.Context()

	created, err := svc.CreateProduct(ctx, "iPhone 13", "Refurbished - Grade A", 49900, uuid.New(), 12.5, -4.25)
	if err != nil {
		t.Fatalf("create product: %v", err)
	}

	got, err := svc.GetProductByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("get product: %v", err)
	}

	if got.Name != created.Name {
		t.Fatalf("expected name %q, got %q", created.Name, got.Name)
	}

	list, err := svc.ListProducts(ctx, 20, 0)
	if err != nil {
		t.Fatalf("list products: %v", err)
	}

	if len(list) != 1 {
		t.Fatalf("expected 1 product, got %d", len(list))
	}
}

func TestProductValidation(t *testing.T) {
	svc := newProductsService(t)
	ctx := t.Context()

	_, err := svc.GetProductByID(ctx, uuid.New())
	if !errors.Is(err, service.ErrProductNotFound) {
		t.Fatalf("expected ErrProductNotFound, got %v", err)
	}

	_, err = svc.CreateProduct(ctx, "", "x", 100, uuid.New(), 0, 0)
	if !errors.Is(err, service.ErrInvalidProductName) {
		t.Fatalf("expected ErrInvalidProductName, got %v", err)
	}

	_, err = svc.CreateProduct(ctx, "Laptop", "x", 0, uuid.New(), 0, 0)
	if !errors.Is(err, service.ErrInvalidPrice) {
		t.Fatalf("expected ErrInvalidPrice, got %v", err)
	}

	_, err = svc.ListProducts(ctx, 0, 0)
	if !errors.Is(err, service.ErrInvalidListLimit) {
		t.Fatalf("expected ErrInvalidListLimit, got %v", err)
	}

	_, err = svc.ListProducts(ctx, 10, -1)
	if !errors.Is(err, service.ErrInvalidListOffset) {
		t.Fatalf("expected ErrInvalidListOffset, got %v", err)
	}
}
