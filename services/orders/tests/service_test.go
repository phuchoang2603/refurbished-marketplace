package tests

import (
	"encoding/json"
	"errors"
	"testing"

	"refurbished-marketplace/services/orders/internal/service"
	"refurbished-marketplace/shared/messaging"
	"refurbished-marketplace/shared/testutil"

	"github.com/google/uuid"
)

func newOrdersService(t *testing.T) (*service.Service, func(aggregateID uuid.UUID) (string, []byte, error)) {
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

	svc := service.New(db)
	readOutbox := func(aggregateID uuid.UUID) (string, []byte, error) {
		var eventType string
		var payloadBytes []byte
		err := db.QueryRowContext(t.Context(), `SELECT event_type, payload FROM orders_outbox WHERE aggregate_id = $1`, aggregateID).Scan(&eventType, &payloadBytes)
		return eventType, payloadBytes, err
	}

	return svc, readOutbox
}

func TestCreateGetListOrder(t *testing.T) {
	svc, _ := newOrdersService(t)
	ctx := t.Context()

	buyerID := uuid.New()
	productID := uuid.New()
	created, err := svc.CreateOrder(
		ctx,
		buyerID,
		[]service.OrderItemInput{{ProductID: productID, Quantity: 2, UnitPriceCents: 9950}},
		19900,
	)
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
	svc, _ := newOrdersService(t)
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

func TestCreateOrderWritesOutbox(t *testing.T) {
	svc, readOutbox := newOrdersService(t)
	ctx := t.Context()

	buyerID := uuid.New()
	productID := uuid.New()
	created, err := svc.CreateOrder(ctx, buyerID, []service.OrderItemInput{{ProductID: productID, Quantity: 2, UnitPriceCents: 9950}}, 19900)
	if err != nil {
		t.Fatalf("create order: %v", err)
	}

	eventType, payloadBytes, err := readOutbox(created.ID)
	if err != nil {
		t.Fatalf("load outbox: %v", err)
	}
	if eventType != string(messaging.EventTypeOrderCreated) {
		t.Fatalf("expected order.created, got %s", eventType)
	}

	type itemPayload struct {
		ProductID      string `json:"product_id"`
		Quantity       int32  `json:"quantity"`
		UnitPriceCents int64  `json:"unit_price_cents"`
	}
	type orderPayload struct {
		OrderID     string        `json:"order_id"`
		BuyerUserID string        `json:"buyer_user_id"`
		TotalCents  int64         `json:"total_cents"`
		Items       []itemPayload `json:"items"`
	}

	var got orderPayload
	if err := json.Unmarshal(payloadBytes, &got); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if got.OrderID != created.ID.String() || got.BuyerUserID != buyerID.String() || got.TotalCents != 19900 {
		t.Fatalf("unexpected payload: %#v", got)
	}
	if len(got.Items) != 1 || got.Items[0].ProductID != productID.String() || got.Items[0].Quantity != 2 {
		t.Fatalf("unexpected payload items: %#v", got.Items)
	}
}
