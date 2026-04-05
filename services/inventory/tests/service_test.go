package tests

import (
	"errors"
	"testing"

	"refurbished-marketplace/services/inventory/internal/database"
	"refurbished-marketplace/services/inventory/internal/service"
	"refurbished-marketplace/shared/testutil"

	"github.com/google/uuid"
)

func newInventoryService(t *testing.T) *service.Service {
	t.Helper()
	db := testutil.SetupPostgresWithMigrations(
		t,
		testutil.PostgresConfig{
			Database: "inventory_db",
			Username: "inventory_app",
			Password: "inventory_app_dev_password",
		},
		"../db/migrations",
	)

	return service.New(database.New(db))
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

	t.Run("reserve stock", func(t *testing.T) {
		productID := uuid.New()
		_, err := svc.CreateInventory(ctx, productID, 10)
		if err != nil {
			t.Fatalf("create inventory: %v", err)
		}

		reserved, err := svc.ReserveStock(ctx, productID, 3)
		if err != nil {
			t.Fatalf("reserve stock: %v", err)
		}
		if reserved.AvailableQty != 7 || reserved.ReservedQty != 3 {
			t.Fatalf("unexpected reserved state")
		}
	})

	t.Run("commit reservation", func(t *testing.T) {
		productID := uuid.New()
		_, err := svc.CreateInventory(ctx, productID, 10)
		if err != nil {
			t.Fatalf("create inventory: %v", err)
		}
		_, err = svc.ReserveStock(ctx, productID, 3)
		if err != nil {
			t.Fatalf("reserve stock: %v", err)
		}

		committed, err := svc.CommitReservation(ctx, productID, 3)
		if err != nil {
			t.Fatalf("commit reservation: %v", err)
		}
		if committed.AvailableQty != 7 || committed.ReservedQty != 0 {
			t.Fatalf("unexpected committed state")
		}
	})

	t.Run("invalid release quantity", func(t *testing.T) {
		productID := uuid.New()
		_, err := svc.CreateInventory(ctx, productID, 10)
		if err != nil {
			t.Fatalf("create inventory: %v", err)
		}

		released, err := svc.ReleaseReservation(ctx, productID, 0)
		if !errors.Is(err, service.ErrInvalidQuantity) {
			t.Fatalf("expected ErrInvalidQuantity, got %v", err)
		}
		_ = released
	})
}

func TestInventoryValidation(t *testing.T) {
	t.Run("invalid product id on get", func(t *testing.T) {
		svc := newInventoryService(t)
		ctx := t.Context()

		_, err := svc.GetInventoryByProductID(ctx, uuid.Nil)
		if !errors.Is(err, service.ErrInvalidProductID) {
			t.Fatalf("expected ErrInvalidProductID, got %v", err)
		}
	})

	t.Run("invalid product id on create", func(t *testing.T) {
		svc := newInventoryService(t)
		ctx := t.Context()

		_, err := svc.CreateInventory(ctx, uuid.Nil, 1)
		if !errors.Is(err, service.ErrInvalidProductID) {
			t.Fatalf("expected ErrInvalidProductID, got %v", err)
		}
	})

	t.Run("invalid quantity on reserve", func(t *testing.T) {
		svc := newInventoryService(t)
		ctx := t.Context()

		_, err := svc.ReserveStock(ctx, uuid.New(), 0)
		if !errors.Is(err, service.ErrInvalidQuantity) {
			t.Fatalf("expected ErrInvalidQuantity, got %v", err)
		}
	})
}
