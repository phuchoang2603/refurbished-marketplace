package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/phuchoang2603/refurbished-marketplace/services/orders/internal/service"
	testpostgres "github.com/phuchoang2603/refurbished-marketplace/shared/testutil/postgres"

	"github.com/google/uuid"
)

func newOrdersService(t *testing.T) *service.Service {
	t.Helper()
	db := testpostgres.SetupPostgresWithMigrations(
		t,
		testpostgres.Config{
			Database: "orders_db",
			Username: "orders_app",
			Password: "orders_app_dev_password",
		},
		"../db/migrations",
	)

	return service.New(db, nil)
}

func TestCreateGetListOrder(t *testing.T) {
	svc := newOrdersService(t)
	ctx := t.Context()

	t.Run("create order", func(t *testing.T) {
		buyerID := uuid.New()
		productID := uuid.New()
		merchantID := uuid.New()
		created, err := svc.CreateOrder(
			ctx,
			buyerID,
			merchantID,
			[]service.OrderItemInput{{ProductID: productID, Quantity: 2, UnitPriceCents: 9950}},
			19900,
			uuid.New(),
		)
		if err != nil {
			t.Fatalf("create order: %v", err)
		}
		if created.BuyerUserID != buyerID || created.MerchantID != merchantID || len(created.Items) != 1 || created.Items[0].ProductID != productID {
			t.Fatalf("unexpected order items")
		}
	})

	t.Run("get order by id", func(t *testing.T) {
		createdBuyerID := uuid.New()
		createdProductID := uuid.New()
		createdMerchantID := uuid.New()
		created, err := svc.CreateOrder(
			ctx,
			createdBuyerID,
			createdMerchantID,
			[]service.OrderItemInput{{ProductID: createdProductID, Quantity: 2, UnitPriceCents: 9950}},
			19900,
			uuid.New(),
		)
		if err != nil {
			t.Fatalf("create order: %v", err)
		}

		got, err := svc.GetOrderByID(ctx, created.ID)
		if err != nil {
			t.Fatalf("get order: %v", err)
		}
		if got.ID != created.ID {
			t.Fatalf("expected same id")
		}
	})

	t.Run("list orders by buyer", func(t *testing.T) {
		buyerID := uuid.New()
		productID := uuid.New()
		merchantID := uuid.New()
		created, err := svc.CreateOrder(
			ctx,
			buyerID,
			merchantID,
			[]service.OrderItemInput{{ProductID: productID, Quantity: 2, UnitPriceCents: 9950}},
			19900,
			uuid.New(),
		)
		if err != nil {
			t.Fatalf("create order: %v", err)
		}

		list, err := svc.ListOrdersByBuyer(ctx, buyerID, 20, 0)
		if err != nil {
			t.Fatalf("list orders: %v", err)
		}
		if len(list) != 1 {
			t.Fatalf("expected 1 order, got %d", len(list))
		}
		if list[0].ID != created.ID {
			t.Fatalf("expected created order in list")
		}
	})

	t.Run("update order status", func(t *testing.T) {
		buyerID := uuid.New()
		productID := uuid.New()
		merchantID := uuid.New()
		created, err := svc.CreateOrder(
			ctx,
			buyerID,
			merchantID,
			[]service.OrderItemInput{{ProductID: productID, Quantity: 2, UnitPriceCents: 9950}},
			19900,
			uuid.New(),
		)
		if err != nil {
			t.Fatalf("create order: %v", err)
		}

		updated, err := svc.UpdateOrderStatus(ctx, created.ID, service.OrderStatusPaid)
		if err != nil {
			t.Fatalf("update order: %v", err)
		}
		if updated.Status != service.OrderStatusPaid {
			t.Fatalf("expected %s, got %s", service.OrderStatusPaid, updated.Status)
		}
	})
}

func TestOrderValidation(t *testing.T) {
	svc := newOrdersService(t)
	ctx := t.Context()

	t.Run("invalid buyer id", func(t *testing.T) {
		_, err := svc.CreateOrder(ctx, uuid.Nil, uuid.New(), []service.OrderItemInput{{ProductID: uuid.New(), Quantity: 1, UnitPriceCents: 100}}, 100, uuid.New())
		if !errors.Is(err, service.ErrInvalidBuyerID) {
			t.Fatalf("expected ErrInvalidBuyerID, got %v", err)
		}
	})

	t.Run("invalid product id", func(t *testing.T) {
		_, err := svc.CreateOrder(ctx, uuid.New(), uuid.New(), []service.OrderItemInput{{ProductID: uuid.Nil, Quantity: 1, UnitPriceCents: 100}}, 100, uuid.New())
		if !errors.Is(err, service.ErrInvalidProductID) {
			t.Fatalf("expected ErrInvalidProductID, got %v", err)
		}
	})

	t.Run("invalid quantity", func(t *testing.T) {
		_, err := svc.CreateOrder(ctx, uuid.New(), uuid.New(), []service.OrderItemInput{{ProductID: uuid.New(), Quantity: 0, UnitPriceCents: 100}}, 100, uuid.New())
		if !errors.Is(err, service.ErrInvalidQuantity) {
			t.Fatalf("expected ErrInvalidQuantity, got %v", err)
		}
	})

	t.Run("invalid merchant id", func(t *testing.T) {
		_, err := svc.CreateOrder(ctx, uuid.New(), uuid.Nil, []service.OrderItemInput{{ProductID: uuid.New(), Quantity: 1, UnitPriceCents: 100}}, 100, uuid.New())
		if !errors.Is(err, service.ErrInvalidMerchantID) {
			t.Fatalf("expected ErrInvalidMerchantID, got %v", err)
		}
	})

	t.Run("missing order", func(t *testing.T) {
		_, err := svc.GetOrderByID(ctx, uuid.Nil)
		if !errors.Is(err, service.ErrOrderNotFound) {
			t.Fatalf("expected ErrOrderNotFound, got %v", err)
		}
	})

	t.Run("invalid buyer id for list", func(t *testing.T) {
		_, err := svc.ListOrdersByBuyer(ctx, uuid.Nil, 10, 0)
		if !errors.Is(err, service.ErrInvalidBuyerID) {
			t.Fatalf("expected ErrInvalidBuyerID, got %v", err)
		}
	})

	t.Run("missing order on update", func(t *testing.T) {
		_, err := svc.UpdateOrderStatus(ctx, uuid.Nil, "")
		if !errors.Is(err, service.ErrOrderNotFound) {
			t.Fatalf("expected ErrOrderNotFound, got %v", err)
		}
	})

	t.Run("invalid status", func(t *testing.T) {
		_, err := svc.UpdateOrderStatus(ctx, uuid.New(), "CONFIRMED")
		if !errors.Is(err, service.ErrInvalidStatus) {
			t.Fatalf("expected ErrInvalidStatus, got %v", err)
		}
	})
}

type failingStock struct{}

func (failingStock) ReserveStock(context.Context, uuid.UUID, uuid.UUID, int64, []service.OrderItemInput) error {
	return service.ErrInsufficientStock
}

func TestCreateOrderMarksFailedWhenReserveFails(t *testing.T) {
	db := testpostgres.SetupPostgresWithMigrations(
		t,
		testpostgres.Config{
			Database: "orders_db",
			Username: "orders_app",
			Password: "orders_app_dev_password",
		},
		"../db/migrations",
	)
	svc := service.New(db, failingStock{})
	ctx := t.Context()
	buyerID := uuid.New()
	merchantID := uuid.New()
	key := uuid.New()
	items := []service.OrderItemInput{{ProductID: uuid.New(), Quantity: 1, UnitPriceCents: 1000}}

	_, err := svc.CreateOrder(ctx, buyerID, merchantID, items, 1000, key)
	if !errors.Is(err, service.ErrInsufficientStock) {
		t.Fatalf("expected ErrInsufficientStock, got %v", err)
	}

	list, err := svc.ListOrdersByBuyer(ctx, buyerID, 20, 0)
	if err != nil {
		t.Fatalf("list orders: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 order, got %d", len(list))
	}
	if list[0].Status != service.OrderStatusFailed {
		t.Fatalf("expected failed order, got %s", list[0].Status)
	}

	_, err = svc.CreateOrder(ctx, buyerID, merchantID, items, 1000, key)
	if !errors.Is(err, service.ErrOrderNotPayable) {
		t.Fatalf("expected ErrOrderNotPayable on retry, got %v", err)
	}
}

func TestCreateOrderReplaysSameIntentKey(t *testing.T) {
	svc := newOrdersService(t)
	ctx := t.Context()
	buyerID := uuid.New()
	merchantID := uuid.New()
	key := uuid.New()
	items := []service.OrderItemInput{{ProductID: uuid.New(), Quantity: 1, UnitPriceCents: 1000}}

	first, err := svc.CreateOrder(ctx, buyerID, merchantID, items, 1000, key)
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	second, err := svc.CreateOrder(ctx, buyerID, merchantID, items, 1000, key)
	if err != nil {
		t.Fatalf("replay create order: %v", err)
	}
	if first.ID != second.ID {
		t.Fatalf("expected same order id, got %s and %s", first.ID, second.ID)
	}
	list, err := svc.ListOrdersByBuyer(ctx, buyerID, 20, 0)
	if err != nil {
		t.Fatalf("list orders: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 order, got %d", len(list))
	}
}
