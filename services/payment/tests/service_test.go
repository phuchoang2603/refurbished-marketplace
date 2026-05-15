package tests

import (
	"encoding/json"
	"errors"
	"testing"

	"refurbished-marketplace/services/payment/internal/database"
	"refurbished-marketplace/services/payment/internal/service"
	"refurbished-marketplace/shared/messaging"
	ordersv1 "refurbished-marketplace/shared/proto/orders/v1"
	"refurbished-marketplace/shared/testutil"

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
	return service.New(db), queries
}

func orderCreatedPayload(orderID, merchantID uuid.UUID, totalCents int64) []byte {
	msg := &ordersv1.OrderCreated{
		OrderId:     orderID.String(),
		BuyerUserId: uuid.New().String(),
		MerchantId:  merchantID.String(),
		TotalCents:  totalCents,
	}
	b, err := proto.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return b
}

func TestPaymentService_InitiatePaymentAndOrdersCreated(t *testing.T) {
	t.Run("creates per-order transaction and get returns view", func(t *testing.T) {
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

		merchantID := uuid.New()
		msgID := "orders.created/test/0/42"
		if err := svc.HandleOrdersCreated(ctx, msgID, orderCreatedPayload(orderID, merchantID, 9900)); err != nil {
			t.Fatalf("HandleOrdersCreated: %v", err)
		}

		row, err := queries.GetPaymentTransactionByOrderID(ctx, orderID)
		if err != nil {
			t.Fatalf("GetPaymentTransactionByOrderID: %v", err)
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
		if view.OrderID != orderID.String() {
			t.Fatalf("view order_id mismatch")
		}
	})
}

func TestPaymentService_HandleOrdersCreated(t *testing.T) {
	t.Run("missing intent", func(t *testing.T) {
		svc, _ := newPaymentFixture(t)
		ctx := t.Context()

		orderID := uuid.New()
		merchantID := uuid.New()
		err := svc.HandleOrdersCreated(ctx, "kafka/t/0/1", orderCreatedPayload(orderID, merchantID, 100))
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

		merchantID := uuid.New()
		payload := orderCreatedPayload(orderID, merchantID, 5000)
		msgID := "kafka/t/0/100"

		if err := svc.HandleOrdersCreated(ctx, msgID, payload); err != nil {
			t.Fatalf("first: %v", err)
		}
		if err := svc.HandleOrdersCreated(ctx, msgID, payload); err != nil {
			t.Fatalf("second (replay): %v", err)
		}
	})

	t.Run("rejects empty message id", func(t *testing.T) {
		svc, _ := newPaymentFixture(t)
		ctx := t.Context()

		err := svc.HandleOrdersCreated(ctx, "", []byte(`{}`))
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

		merchantID := uuid.New()
		if err := svc.HandleOrdersCreated(ctx, "kafka/t/0/200", orderCreatedPayload(orderID, merchantID, 3000)); err != nil {
			t.Fatalf("HandleOrdersCreated: %v", err)
		}

		txRow, err := queries.GetPaymentTransactionByOrderID(ctx, orderID)
		if err != nil {
			t.Fatalf("GetPaymentTransactionByOrderID: %v", err)
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
			AggregateID: orderID,
			EventType:   messaging.EventTypePaymentSucceeded,
		})
		if err != nil {
			t.Fatalf("GetPaymentOutboxByAggregateIDAndEventType: %v", err)
		}
		if outboxRow.AggregateID != orderID {
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
