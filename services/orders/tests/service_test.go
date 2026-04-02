package tests

import (
	"errors"
	"testing"

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

	return service.New(db)
}

func TestCreateGetListOrder(t *testing.T) {
	svc := newOrdersService(t)
	ctx := t.Context()

	buyerID := uuid.New()
	productID := uuid.New()
	created, err := svc.CreateOrder(ctx, buyerID, []service.OrderItemInput{{ProductID: productID, Quantity: 2, UnitPriceCents: 9950}}, 19900)
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	if created.BuyerUserID != buyerID || len(created.Items) != 1 || created.Items[0].ProductID != productID {
		t.Fatalf("unexpected order items")
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

	updated, err := svc.UpdateOrderStatus(ctx, created.ID, service.OrderStatusPaid)
	if err != nil {
		t.Fatalf("update order: %v", err)
	}
	if updated.Status != service.OrderStatusPaid {
		t.Fatalf("expected %s, got %s", service.OrderStatusPaid, updated.Status)
	}
}

func TestOrderValidation(t *testing.T) {
	svc := newOrdersService(t)
	ctx := t.Context()

	_, err := svc.CreateOrder(ctx, uuid.Nil, []service.OrderItemInput{{ProductID: uuid.New(), Quantity: 1, UnitPriceCents: 100}}, 100)
	if !errors.Is(err, service.ErrInvalidBuyerID) {
		t.Fatalf("expected ErrInvalidBuyerID, got %v", err)
	}

	_, err = svc.CreateOrder(ctx, uuid.New(), []service.OrderItemInput{{ProductID: uuid.Nil, Quantity: 1, UnitPriceCents: 100}}, 100)
	if !errors.Is(err, service.ErrInvalidProductID) {
		t.Fatalf("expected ErrInvalidProductID, got %v", err)
	}

	_, err = svc.CreateOrder(ctx, uuid.New(), []service.OrderItemInput{{ProductID: uuid.New(), Quantity: 0, UnitPriceCents: 100}}, 100)
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

	_, err = svc.UpdateOrderStatus(ctx, uuid.New(), "CONFIRMED")
	if !errors.Is(err, service.ErrInvalidStatus) {
		t.Fatalf("expected ErrInvalidStatus, got %v", err)
	}
}
