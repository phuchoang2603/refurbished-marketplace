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
	productID := uuid.New()

	created, err := svc.CreateInventory(ctx, productID, 10)
	if err != nil {
		t.Fatalf("create inventory: %v", err)
	}
	if created.ProductID != productID || created.AvailableQty != 10 || created.ReservedQty != 0 {
		t.Fatalf("unexpected inventory state")
	}

	reserved, err := svc.ReserveStock(ctx, productID, 3)
	if err != nil {
		t.Fatalf("reserve stock: %v", err)
	}
	if reserved.AvailableQty != 7 || reserved.ReservedQty != 3 {
		t.Fatalf("unexpected reserved state")
	}

	committed, err := svc.CommitReservation(ctx, productID, 3)
	if err != nil {
		t.Fatalf("commit reservation: %v", err)
	}
	if committed.AvailableQty != 7 || committed.ReservedQty != 0 {
		t.Fatalf("unexpected committed state")
	}

	released, err := svc.ReleaseReservation(ctx, productID, 0)
	if !errors.Is(err, service.ErrInvalidQuantity) {
		t.Fatalf("expected ErrInvalidQuantity, got %v", err)
	}
	_ = released
}

func TestInventoryValidation(t *testing.T) {
	svc := newInventoryService(t)
	ctx := t.Context()

	_, err := svc.GetInventoryByProductID(ctx, uuid.Nil)
	if !errors.Is(err, service.ErrInvalidProductID) {
		t.Fatalf("expected ErrInvalidProductID, got %v", err)
	}

	_, err = svc.CreateInventory(ctx, uuid.Nil, 1)
	if !errors.Is(err, service.ErrInvalidProductID) {
		t.Fatalf("expected ErrInvalidProductID, got %v", err)
	}

	_, err = svc.ReserveStock(ctx, uuid.New(), 0)
	if !errors.Is(err, service.ErrInvalidQuantity) {
		t.Fatalf("expected ErrInvalidQuantity, got %v", err)
	}
}
