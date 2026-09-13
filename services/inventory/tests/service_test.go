package tests

import (
	"database/sql"
	"testing"

	"github.com/phuchoang2603/refurbished-marketplace/services/inventory/internal/service"
	testpostgres "github.com/phuchoang2603/refurbished-marketplace/shared/testutil/postgres"

	"github.com/google/uuid"
)

func newInventoryService(t *testing.T) *service.Service {
	t.Helper()
	return service.New(newInventoryDB(t))
}

func newInventoryDB(t *testing.T) *sql.DB {
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
	return db
}

func TestEnsureAndReserveStock(t *testing.T) {
	svc := newInventoryService(t)
	ctx := t.Context()
	productID := uuid.New()

	created, err := svc.EnsureStock(ctx, productID, 5)
	if err != nil {
		t.Fatalf("EnsureStock: %v", err)
	}
	if created.AvailableQty != 5 {
		t.Fatalf("available = %d, want 5", created.AvailableQty)
	}
	again, err := svc.EnsureStock(ctx, productID, 99)
	if err != nil {
		t.Fatalf("EnsureStock retry: %v", err)
	}
	if again.AvailableQty != 5 {
		t.Fatalf("idempotent available = %d, want 5", again.AvailableQty)
	}

	orderID := uuid.New()
	if err := svc.ReserveStock(ctx, orderID, uuid.New(), 1000, []service.ReservationItemInput{{ProductID: productID, Quantity: 2}}); err != nil {
		t.Fatalf("ReserveStock: %v", err)
	}
	got, err := svc.GetInventoryByProductID(ctx, productID)
	if err != nil {
		t.Fatalf("GetInventoryByProductID: %v", err)
	}
	if got.AvailableQty != 3 || got.ReservedQty != 2 {
		t.Fatalf("unexpected stock %+v", got)
	}
	if err := svc.ReserveStock(ctx, orderID, uuid.New(), 1000, []service.ReservationItemInput{{ProductID: productID, Quantity: 2}}); err != nil {
		t.Fatalf("ReserveStock idempotent: %v", err)
	}
	got, err = svc.GetInventoryByProductID(ctx, productID)
	if err != nil {
		t.Fatalf("GetInventoryByProductID after retry: %v", err)
	}
	if got.AvailableQty != 3 || got.ReservedQty != 2 {
		t.Fatalf("idempotent reserve changed stock %+v", got)
	}
}
