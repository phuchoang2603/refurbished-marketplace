package tests

import (
	"database/sql"
	"errors"
	"testing"

	"refurbished-marketplace/services/orders/internal/database"
	"refurbished-marketplace/services/orders/internal/service"
	"refurbished-marketplace/shared/testutil"

	"github.com/google/uuid"
)

func newOrdersService(t *testing.T) *service.Service {
	t.Helper()
	db := testutil.SetupPostgresWithMigrations(
		t,
		testutil.PostgresConfig{
			Database: "orders_db",
			Username: "orders_app",
			Password: "orders_app_dev_password",
		},
		"../db/migrations",
	)

	return service.New(database.New(db))
}

func TestCreateGetListOrder(t *testing.T) {
	svc := newOrdersService(t)
	ctx := t.Context()

	buyerID := uuid.New()
	productID := uuid.New()
	created, err := svc.CreateOrder(ctx, buyerID, productID, 2, 19900)
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	if created.BuyerUserID != buyerID || created.ProductID != productID {
		t.Fatalf("unexpected order ids")
	}

	got, err := svc.GetOrderByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("get order: %v", err)
	}
	if got.ID != created.ID {
		t.Fatalf("expected same id")
	}

	list, err := svc.ListOrdersByBuyer(ctx, buyerID, 20, 0)
	if err != nil {
		t.Fatalf("list orders: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 order, got %d", len(list))
	}

	updated, err := svc.UpdateOrderStatus(ctx, created.ID, "confirmed")
	if err != nil {
		t.Fatalf("update order: %v", err)
	}
	if updated.Status != "CONFIRMED" {
		t.Fatalf("expected CONFIRMED, got %s", updated.Status)
	}
}

func TestOrderValidation(t *testing.T) {
	svc := newOrdersService(t)
	ctx := t.Context()

	_, err := svc.CreateOrder(ctx, uuid.Nil, uuid.New(), 1, 100)
	if !errors.Is(err, service.ErrInvalidBuyerID) {
		t.Fatalf("expected ErrInvalidBuyerID, got %v", err)
	}

	_, err = svc.CreateOrder(ctx, uuid.New(), uuid.Nil, 1, 100)
	if !errors.Is(err, service.ErrInvalidProductID) {
		t.Fatalf("expected ErrInvalidProductID, got %v", err)
	}

	_, err = svc.CreateOrder(ctx, uuid.New(), uuid.New(), 0, 100)
	if !errors.Is(err, service.ErrInvalidQuantity) {
		t.Fatalf("expected ErrInvalidQuantity, got %v", err)
	}

	_, err = svc.GetOrderByID(ctx, uuid.Nil)
	if !errors.Is(err, service.ErrOrderNotFound) {
		t.Fatalf("expected ErrOrderNotFound, got %v", err)
	}

	_, err = svc.ListOrdersByBuyer(ctx, uuid.Nil, 10, 0)
	if !errors.Is(err, service.ErrInvalidBuyerID) {
		t.Fatalf("expected ErrInvalidBuyerID, got %v", err)
	}

	_, err = svc.UpdateOrderStatus(ctx, uuid.Nil, "")
	if !errors.Is(err, service.ErrOrderNotFound) {
		t.Fatalf("expected ErrOrderNotFound, got %v", err)
	}
}

func TestMissingOrderNoRows(t *testing.T) {
	db := testutil.SetupPostgresWithMigrations(
		t,
		testutil.PostgresConfig{
			Database: "orders_db",
			Username: "orders_app",
			Password: "orders_app_dev_password",
		},
		"../db/migrations",
	)

	queries := database.New(db)
	_, err := queries.GetOrderByID(t.Context(), uuid.New())
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
}
