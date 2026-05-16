package tests

import (
	"errors"
	"testing"

	"refurbished-marketplace/services/inventory/internal/service"
	testpostgres "refurbished-marketplace/shared/testutil/postgres"

	"github.com/google/uuid"
)

func newInventoryService(t *testing.T) *service.Service {
	t.Helper()
	db := testpostgres.SetupPostgresWithMigrations(
		t,
		testpostgres.Config{
			Database: "inventory_db",
			Username: "inventory_app",
			Password: "inventory_app_dev_password",
		},
		"../db/migrations",
	)

	return service.New(db)
}

func TestInventoryLifecycle(t *testing.T) {
	svc := newInventoryService(t)
	ctx := t.Context()

	t.Run("create inventory", func(t *testing.T) {
		productID := uuid.New()
		created, err := svc.CreateInventory(ctx, productID, 10)
		if err != nil {
			t.Fatalf("create inventory: %v", err)
		}
		if created.ProductID != productID || created.AvailableQty != 10 || created.ReservedQty != 0 {
			t.Fatalf("unexpected inventory state")
		}
	})
}

func TestInventoryValidation(t *testing.T) {
	svc := newInventoryService(t)
	ctx := t.Context()

	t.Run("invalid product id on get", func(t *testing.T) {
		_, err := svc.GetInventoryByProductID(ctx, uuid.Nil)
		if !errors.Is(err, service.ErrInvalidProductID) {
			t.Fatalf("expected ErrInvalidProductID, got %v", err)
		}
	})

	t.Run("invalid product id on create", func(t *testing.T) {
		_, err := svc.CreateInventory(ctx, uuid.Nil, 1)
		if !errors.Is(err, service.ErrInvalidProductID) {
			t.Fatalf("expected ErrInvalidProductID, got %v", err)
		}
	})

	t.Run("invalid quantity on create", func(t *testing.T) {
		_, err := svc.CreateInventory(ctx, uuid.New(), -1)
		if !errors.Is(err, service.ErrInvalidQuantity) {
			t.Fatalf("expected ErrInvalidQuantity, got %v", err)
		}
	})
}
