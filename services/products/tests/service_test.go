package tests

import (
	"errors"
	"refurbished-marketplace/services/products/internal/database"
	"refurbished-marketplace/services/products/internal/service"
	"refurbished-marketplace/shared/testutil"
	"testing"

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

	t.Run("create product", func(t *testing.T) {
		merchantID := uuid.New()
		created, err := svc.CreateProduct(ctx, "iPhone 13", "Refurbished - Grade A", 49900, merchantID)
		if err != nil {
			t.Fatalf("create product: %v", err)
		}
		if created.Name == "" {
			t.Fatalf("expected created product")
		}
	})

	t.Run("get product by id", func(t *testing.T) {
		merchantID := uuid.New()
		created, err := svc.CreateProduct(ctx, "iPhone 13", "Refurbished - Grade A", 49900, merchantID)
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
		if got.MerchantID != merchantID {
			t.Fatalf("expected merchant id %s, got %s", merchantID, got.MerchantID)
		}
	})

	t.Run("list products", func(t *testing.T) {
		created, err := svc.CreateProduct(ctx, "iPhone 13", "Refurbished - Grade A", 49900, uuid.New())
		if err != nil {
			t.Fatalf("create product: %v", err)
		}

		list, err := svc.ListProducts(ctx, 20, 0)
		if err != nil {
			t.Fatalf("list products: %v", err)
		}

		found := false
		for _, item := range list {
			if item.ID == created.ID {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected created product in list")
		}
	})
}

func TestProductValidation(t *testing.T) {
	t.Run("missing product", func(t *testing.T) {
		svc := newProductsService(t)
		ctx := t.Context()

		_, err := svc.GetProductByID(ctx, uuid.New())
		if !errors.Is(err, service.ErrProductNotFound) {
			t.Fatalf("expected ErrProductNotFound, got %v", err)
		}
	})

	t.Run("invalid product name", func(t *testing.T) {
		svc := newProductsService(t)
		ctx := t.Context()

		_, err := svc.CreateProduct(ctx, "", "x", 100, uuid.New())
		if !errors.Is(err, service.ErrInvalidProductName) {
			t.Fatalf("expected ErrInvalidProductName, got %v", err)
		}
	})

	t.Run("invalid merchant id", func(t *testing.T) {
		svc := newProductsService(t)
		ctx := t.Context()

		_, err := svc.CreateProduct(ctx, "Laptop", "x", 100, uuid.Nil)
		if !errors.Is(err, service.ErrInvalidMerchantID) {
			t.Fatalf("expected ErrInvalidMerchantID, got %v", err)
		}
	})

	t.Run("invalid price", func(t *testing.T) {
		svc := newProductsService(t)
		ctx := t.Context()

		_, err := svc.CreateProduct(ctx, "Laptop", "x", 0, uuid.New())
		if !errors.Is(err, service.ErrInvalidPrice) {
			t.Fatalf("expected ErrInvalidPrice, got %v", err)
		}
	})

	t.Run("invalid list limit", func(t *testing.T) {
		svc := newProductsService(t)
		ctx := t.Context()

		_, err := svc.ListProducts(ctx, 0, 0)
		if !errors.Is(err, service.ErrInvalidListLimit) {
			t.Fatalf("expected ErrInvalidListLimit, got %v", err)
		}
	})

	t.Run("invalid list offset", func(t *testing.T) {
		svc := newProductsService(t)
		ctx := t.Context()

		_, err := svc.ListProducts(ctx, 10, -1)
		if !errors.Is(err, service.ErrInvalidListOffset) {
			t.Fatalf("expected ErrInvalidListOffset, got %v", err)
		}
	})
}
