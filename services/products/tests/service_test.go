package tests

import (
	"database/sql"
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

	ownerID := uuid.New()
	created, err := svc.CreateProduct(ctx, ownerID, "iPhone 13", "Refurbished - Grade A", 49900, 7)
	if err != nil {
		t.Fatalf("create product: %v", err)
	}
	if created.OwnerUserID != ownerID {
		t.Fatalf("expected owner %s, got %s", ownerID, created.OwnerUserID)
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

	_, err := svc.CreateProduct(ctx, uuid.Nil, "Phone", "x", 100, 1)
	if !errors.Is(err, service.ErrInvalidOwner) {
		t.Fatalf("expected ErrInvalidOwner, got %v", err)
	}

	ownerID := uuid.New()

	_, err = svc.CreateProduct(ctx, ownerID, "", "x", 100, 1)
	if !errors.Is(err, service.ErrInvalidProductName) {
		t.Fatalf("expected ErrInvalidProductName, got %v", err)
	}

	_, err = svc.CreateProduct(ctx, ownerID, "Laptop", "x", 0, 1)
	if !errors.Is(err, service.ErrInvalidPrice) {
		t.Fatalf("expected ErrInvalidPrice, got %v", err)
	}

	_, err = svc.CreateProduct(ctx, ownerID, "Laptop", "x", 100, -1)
	if !errors.Is(err, service.ErrInvalidStock) {
		t.Fatalf("expected ErrInvalidStock, got %v", err)
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

func TestMissingProductAndNoRows(t *testing.T) {
	db := testutil.SetupPostgresWithMigrations(
		t,
		testutil.PostgresConfig{
			Database: "products_db",
			Username: "products_app",
			Password: "products_app_dev_password",
		},
		"../db/migrations",
	)

	queries := database.New(db)
	svc := service.New(queries)

	_, err := svc.GetProductByID(t.Context(), uuid.New())
	if !errors.Is(err, service.ErrProductNotFound) {
		t.Fatalf("expected ErrProductNotFound, got %v", err)
	}

	_, err = queries.GetProductByID(t.Context(), uuid.New())
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestProductOwnershipMutations(t *testing.T) {
	svc := newProductsService(t)
	ctx := t.Context()

	ownerID := uuid.New()
	otherOwnerID := uuid.New()

	created, err := svc.CreateProduct(ctx, ownerID, "Pixel 8", "Refurbished", 39900, 5)
	if err != nil {
		t.Fatalf("create product: %v", err)
	}

	newName := "Pixel 8 Pro"
	newStock := int32(3)
	updated, err := svc.UpdateProduct(ctx, created.ID, ownerID, service.UpdateProductInput{Name: &newName, Stock: &newStock})
	if err != nil {
		t.Fatalf("update product: %v", err)
	}
	if updated.Name != newName || updated.Stock != newStock {
		t.Fatalf("expected updated fields to be applied")
	}

	_, err = svc.UpdateProduct(ctx, created.ID, otherOwnerID, service.UpdateProductInput{Name: &newName})
	if !errors.Is(err, service.ErrForbiddenProduct) {
		t.Fatalf("expected ErrForbiddenProduct on non-owner update, got %v", err)
	}

	err = svc.DeleteProduct(ctx, created.ID, otherOwnerID)
	if !errors.Is(err, service.ErrForbiddenProduct) {
		t.Fatalf("expected ErrForbiddenProduct on non-owner delete, got %v", err)
	}

	err = svc.DeleteProduct(ctx, created.ID, ownerID)
	if err != nil {
		t.Fatalf("delete product: %v", err)
	}
}
