package tests

import (
	"encoding/json"
	"errors"
	"refurbished-marketplace/services/payment/internal/database"
	"refurbished-marketplace/services/payment/internal/service"
	"refurbished-marketplace/shared/messaging"
	ordersv1 "refurbished-marketplace/shared/proto/orders/v1"
	"refurbished-marketplace/shared/testutil"
	"testing"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

func newPaymentFixture(t *testing.T) (*service.Service, *database.Queries) {
	t.Helper()
	db := testutil.SetupPostgresWithMigrations(
		t,
		testutil.PostgresConfig{
			Database: "payment_db",
			Username: "payment_app",
			Password: "payment_app_dev_password",
		},
		"../db/migrations",
	)
	queries := database.New(db)
	return service.New(queries, db), queries
}

func orderItemCreatedPayload(orderID, orderItemID, merchantID uuid.UUID, lineTotal int64) []byte {
	msg := &ordersv1.OrderItemCreated{
		OrderId:        orderID.String(),
		OrderItemId:    orderItemID.String(),
		BuyerUserId:    uuid.New().String(),
		ProductId:      uuid.New().String(),
		MerchantId:     merchantID.String(),
		Quantity:       1,
		UnitPriceCents: lineTotal,
		LineTotalCents: lineTotal,
	}
	b, err := proto.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return b
}

func TestPaymentService_InitiatePaymentAndOrdersItemCreated(t *testing.T) {
	t.Run("creates per-item transaction and get returns view", func(t *testing.T) {
		svc, queries := newPaymentFixture(t)
		ctx := t.Context()

		orderID := uuid.New()
		buyerID := uuid.New()
		if err := svc.InitiatePayment(ctx, service.InitiatePaymentParams{
			OrderID:         orderID,
			BuyerUserID:     buyerID,
			PaymentToken:    "tok_visa",
			Currency:        "USD",
			BillingAddress:  json.RawMessage(`{}`),
			ShippingAddress: json.RawMessage(`{}`),
		}); err != nil {
			t.Fatalf("InitiatePayment: %v", err)
		}

		orderItemID := uuid.New()
		merchantID := uuid.New()
		msgID := "orders.item.created/test/0/42"
		if err := svc.HandleOrdersItemCreated(ctx, msgID, orderItemCreatedPayload(orderID, orderItemID, merchantID, 9900)); err != nil {
			t.Fatalf("HandleOrdersItemCreated: %v", err)
		}

		row, err := queries.GetPaymentTransactionByOrderItemID(ctx, orderItemID)
		if err != nil {
			t.Fatalf("GetPaymentTransactionByOrderItemID: %v", err)
		}
		if row.OrderID != orderID {
			t.Fatalf("order_id mismatch")
		}
		if row.Status != service.PaymentTxStatusInitialized {
			t.Fatalf("status: got %q want %s", row.Status, service.PaymentTxStatusInitialized)
		}

		view, err := svc.GetPaymentTransaction(ctx, row.ID)
		if err != nil {
			t.Fatalf("GetPaymentTransaction: %v", err)
		}
		if view.OrderItemID != orderItemID.String() {
			t.Fatalf("view order_item_id")
		}
	})
}

func TestPaymentService_HandleOrdersItemCreated(t *testing.T) {
	t.Run("missing intent", func(t *testing.T) {
		svc, _ := newPaymentFixture(t)
		ctx := t.Context()

		orderID := uuid.New()
		orderItemID := uuid.New()
		merchantID := uuid.New()
		err := svc.HandleOrdersItemCreated(ctx, "kafka/t/0/1", orderItemCreatedPayload(orderID, orderItemID, merchantID, 100))
		if !errors.Is(err, service.ErrIntentNotFound) {
			t.Fatalf("expected ErrIntentNotFound, got %v", err)
		}
	})

	t.Run("idempotent kafka replay", func(t *testing.T) {
		svc, _ := newPaymentFixture(t)
		ctx := t.Context()

		orderID := uuid.New()
		buyerID := uuid.New()
		if err := svc.InitiatePayment(ctx, service.InitiatePaymentParams{
			OrderID:         orderID,
			BuyerUserID:     buyerID,
			PaymentToken:    "tok_visa",
			Currency:        "USD",
			BillingAddress:  json.RawMessage(`{}`),
			ShippingAddress: json.RawMessage(`{}`),
		}); err != nil {
			t.Fatalf("InitiatePayment: %v", err)
		}

		orderItemID := uuid.New()
		merchantID := uuid.New()
		payload := orderItemCreatedPayload(orderID, orderItemID, merchantID, 5000)
		msgID := "kafka/t/0/100"

		if err := svc.HandleOrdersItemCreated(ctx, msgID, payload); err != nil {
			t.Fatalf("first: %v", err)
		}
		if err := svc.HandleOrdersItemCreated(ctx, msgID, payload); err != nil {
			t.Fatalf("second (replay): %v", err)
		}
	})

	t.Run("rejects empty message id", func(t *testing.T) {
		svc, _ := newPaymentFixture(t)
		ctx := t.Context()

		err := svc.HandleOrdersItemCreated(ctx, "", []byte(`{}`))
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestPaymentService_ApplyGatewayWebhook(t *testing.T) {
	t.Run("succeeded updates transaction writes outbox and ignores duplicate apply", func(t *testing.T) {
		svc, queries := newPaymentFixture(t)
		ctx := t.Context()

		orderID := uuid.New()
		buyerID := uuid.New()
		if err := svc.InitiatePayment(ctx, service.InitiatePaymentParams{
			OrderID:         orderID,
			BuyerUserID:     buyerID,
			PaymentToken:    "tok_visa",
			Currency:        "USD",
			BillingAddress:  json.RawMessage(`{}`),
			ShippingAddress: json.RawMessage(`{}`),
		}); err != nil {
			t.Fatalf("InitiatePayment: %v", err)
		}

		orderItemID := uuid.New()
		merchantID := uuid.New()
		if err := svc.HandleOrdersItemCreated(ctx, "kafka/t/0/200", orderItemCreatedPayload(orderID, orderItemID, merchantID, 3000)); err != nil {
			t.Fatalf("HandleOrdersItemCreated: %v", err)
		}

		txRow, err := queries.GetPaymentTransactionByOrderItemID(ctx, orderItemID)
		if err != nil {
			t.Fatalf("GetPaymentTransactionByOrderItemID: %v", err)
		}

		if err := svc.ApplyGatewayWebhook(ctx, txRow.ID, "gw_abc", true, ""); err != nil {
			t.Fatalf("ApplyGatewayWebhook: %v", err)
		}

		view, err := svc.GetPaymentTransaction(ctx, txRow.ID)
		if err != nil {
			t.Fatalf("GetPaymentTransaction: %v", err)
		}
		if view.Status != service.PaymentTxStatusSucceeded {
			t.Fatalf("status: got %q", view.Status)
		}

		outboxRow, err := queries.GetPaymentOutboxByAggregateIDAndEventType(ctx, database.GetPaymentOutboxByAggregateIDAndEventTypeParams{
			AggregateID: orderItemID,
			EventType:   messaging.EventTypePaymentItemSucceeded,
		})
		if err != nil {
			t.Fatalf("GetPaymentOutboxByAggregateIDAndEventType: %v", err)
		}
		if outboxRow.AggregateID != orderItemID {
			t.Fatalf("outbox aggregate_id")
		}

		if err := svc.ApplyGatewayWebhook(ctx, txRow.ID, "gw_ignored", true, ""); err != nil {
			t.Fatalf("ApplyGatewayWebhook idempotent second call: %v", err)
		}
	})

	t.Run("transaction not found", func(t *testing.T) {
		svc, _ := newPaymentFixture(t)
		ctx := t.Context()

		err := svc.ApplyGatewayWebhook(ctx, uuid.New(), "gw", true, "")
		if !errors.Is(err, service.ErrTransactionNotFound) {
			t.Fatalf("expected ErrTransactionNotFound, got %v", err)
		}
	})
}
