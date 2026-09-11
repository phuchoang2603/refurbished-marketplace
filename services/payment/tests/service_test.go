package tests

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/phuchoang2603/refurbished-marketplace/services/payment/internal/database"
	"github.com/phuchoang2603/refurbished-marketplace/services/payment/internal/service"
	"github.com/phuchoang2603/refurbished-marketplace/shared/err/dberr"
	"github.com/phuchoang2603/refurbished-marketplace/shared/messaging"
	testpostgres "github.com/phuchoang2603/refurbished-marketplace/shared/testutil/postgres"

	"github.com/google/uuid"
)

func newPaymentFixture(t *testing.T) (*service.Service, *database.Queries) {
	t.Helper()
	db := testpostgres.SetupPostgresWithMigrations(
		t,
		testpostgres.Config{
			Database: "payment_db",
			Username: "payment_app",
			Password: "payment_app_dev_password",
		},
		"../db/migrations",
	)
	queries := database.New(db)
	return service.New(db), queries
}

func createTestSession(t *testing.T, svc *service.Service, orderID uuid.UUID) service.HostedPaymentSessionView {
	t.Helper()
	session, err := svc.CreateHostedPaymentSession(t.Context(), service.CreateHostedPaymentSessionParams{
		OrderID:         orderID,
		BuyerUserID:     uuid.New(),
		MerchantID:      uuid.New(),
		TotalCents:      3000,
		Currency:        "USD",
		ShippingAddress: json.RawMessage(`{"line1":"1 Main St","city":"New York","postal_code":"10001","country":"US"}`),
		LineItems:       json.RawMessage(`[]`),
		ReturnURL:       "/orders/" + orderID.String(),
	})
	if err != nil {
		t.Fatalf("CreateHostedPaymentSession: %v", err)
	}
	return session
}

func TestPaymentService_ApplyGatewayWebhook(t *testing.T) {
	t.Run("succeeded updates transaction writes outbox and ignores duplicate apply", func(t *testing.T) {
		svc, queries := newPaymentFixture(t)
		ctx := t.Context()

		orderID := uuid.New()
		session := createTestSession(t, svc, orderID)

		txRow, err := queries.GetPaymentTransactionByOrderID(ctx, orderID)
		if err != nil {
			t.Fatalf("GetPaymentTransactionByOrderID: %v", err)
		}

		if err := svc.ApplyGatewayWebhook(ctx, orderID, session.PaymentSessionID, service.HostedPaymentSessionStatusSucceeded, ""); err != nil {
			t.Fatalf("ApplyGatewayWebhook: %v", err)
		}

		view, err := svc.GetPaymentTransaction(ctx, txRow.ID)
		if err != nil {
			t.Fatalf("GetPaymentTransaction: %v", err)
		}
		if view.Status != service.PaymentTxStatusSucceeded {
			t.Fatalf("status: got %q", view.Status)
		}

		if err := svc.ApplyGatewayWebhook(ctx, orderID, session.PaymentSessionID, service.HostedPaymentSessionStatusSucceeded, ""); err != nil {
			t.Fatalf("ApplyGatewayWebhook idempotent second call: %v", err)
		}
	})

	t.Run("session not found", func(t *testing.T) {
		svc, _ := newPaymentFixture(t)
		ctx := t.Context()

		err := svc.ApplyGatewayWebhook(ctx, uuid.New(), "sess", service.HostedPaymentSessionStatusSucceeded, "")
		if !errors.Is(err, service.ErrIntentNotFound) {
			t.Fatalf("expected ErrIntentNotFound, got %v", err)
		}
	})
}

func TestPaymentService_ExpireDueSessions(t *testing.T) {
	t.Run("expired pending with transaction emits payment.failed", func(t *testing.T) {
		svc, queries := newPaymentFixture(t)
		ctx := t.Context()

		orderID := uuid.New()
		_ = createTestSession(t, svc, orderID)

		past := time.Now().UTC().Add(-time.Minute)
		if err := queries.SetPaymentIntentExpiresAt(ctx, database.SetPaymentIntentExpiresAtParams{
			OrderID:   orderID,
			ExpiresAt: dberr.OptionalNullTime(past),
		}); err != nil {
			t.Fatalf("SetPaymentIntentExpiresAt: %v", err)
		}

		if err := svc.ExpireDueSessions(ctx); err != nil {
			t.Fatalf("ExpireDueSessions: %v", err)
		}

		intent, err := queries.GetPaymentIntentByOrderID(ctx, orderID)
		if err != nil {
			t.Fatalf("GetPaymentIntentByOrderID: %v", err)
		}
		if intent.Status != service.HostedPaymentSessionStatusExpired {
			t.Fatalf("intent status: got %q want EXPIRED", intent.Status)
		}

		txRow, err := queries.GetPaymentTransactionByOrderID(ctx, orderID)
		if err != nil {
			t.Fatalf("GetPaymentTransactionByOrderID: %v", err)
		}
		if txRow.Status != service.PaymentTxStatusFailed {
			t.Fatalf("transaction status: got %q want FAILED", txRow.Status)
		}

		outbox, err := queries.ListPaymentOutboxByAggregateID(ctx, orderID)
		if err != nil {
			t.Fatalf("ListPaymentOutboxByAggregateID: %v", err)
		}
		if len(outbox) != 1 {
			t.Fatalf("outbox rows: got %d want 1", len(outbox))
		}
		if outbox[0].EventType != messaging.EventTypePaymentFailed {
			t.Fatalf("outbox event: got %q want %q", outbox[0].EventType, messaging.EventTypePaymentFailed)
		}

		if err := svc.ExpireDueSessions(ctx); err != nil {
			t.Fatalf("ExpireDueSessions idempotent: %v", err)
		}
		outbox2, err := queries.ListPaymentOutboxByAggregateID(ctx, orderID)
		if err != nil {
			t.Fatalf("ListPaymentOutboxByAggregateID after second sweep: %v", err)
		}
		if len(outbox2) != 1 {
			t.Fatalf("outbox rows after second sweep: got %d want 1", len(outbox2))
		}
	})

	t.Run("expired before inventory.reserved catch-up is idempotent", func(t *testing.T) {
		svc, queries := newPaymentFixture(t)
		ctx := t.Context()

		orderID := uuid.New()
		merchantID := uuid.New()
		_ = createTestSession(t, svc, orderID)

		past := time.Now().UTC().Add(-time.Minute)
		if err := queries.SetPaymentIntentExpiresAt(ctx, database.SetPaymentIntentExpiresAtParams{
			OrderID:   orderID,
			ExpiresAt: dberr.OptionalNullTime(past),
		}); err != nil {
			t.Fatalf("SetPaymentIntentExpiresAt: %v", err)
		}
		if err := svc.ExpireDueSessions(ctx); err != nil {
			t.Fatalf("ExpireDueSessions: %v", err)
		}

		handler := svc.KafkaInventoryReservedHandler()
		if err := handler(ctx, messaging.KafkaMessage{
			Topic:     messaging.EventTypeInventoryReserved,
			Partition: 0,
			Offset:    42,
			Value:     inventoryReservedPayload(orderID, merchantID, 4200),
		}); err != nil {
			t.Fatalf("KafkaInventoryReservedHandler: %v", err)
		}

		txRow, err := queries.GetPaymentTransactionByOrderID(ctx, orderID)
		if err != nil {
			t.Fatalf("GetPaymentTransactionByOrderID: %v", err)
		}
		if txRow.Status != service.PaymentTxStatusFailed {
			t.Fatalf("transaction status: got %q want FAILED", txRow.Status)
		}

		outbox, err := queries.ListPaymentOutboxByAggregateID(ctx, orderID)
		if err != nil {
			t.Fatalf("ListPaymentOutboxByAggregateID: %v", err)
		}
		if len(outbox) != 1 {
			t.Fatalf("outbox rows: got %d want 1", len(outbox))
		}

		if err := handler(ctx, messaging.KafkaMessage{
			Topic:     messaging.EventTypeInventoryReserved,
			Partition: 0,
			Offset:    42,
			Value:     inventoryReservedPayload(orderID, merchantID, 4200),
		}); err != nil {
			t.Fatalf("KafkaInventoryReservedHandler retry: %v", err)
		}
		outbox2, err := queries.ListPaymentOutboxByAggregateID(ctx, orderID)
		if err != nil {
			t.Fatalf("ListPaymentOutboxByAggregateID after retry: %v", err)
		}
		if len(outbox2) != 1 {
			t.Fatalf("outbox rows after retry: got %d want 1", len(outbox2))
		}
	})

	t.Run("gateway webhook cannot overwrite expired session", func(t *testing.T) {
		svc, queries := newPaymentFixture(t)
		ctx := t.Context()

		orderID := uuid.New()
		session := createTestSession(t, svc, orderID)
		past := time.Now().UTC().Add(-time.Minute)
		if err := queries.SetPaymentIntentExpiresAt(ctx, database.SetPaymentIntentExpiresAtParams{
			OrderID:   orderID,
			ExpiresAt: dberr.OptionalNullTime(past),
		}); err != nil {
			t.Fatalf("SetPaymentIntentExpiresAt: %v", err)
		}
		if err := svc.ExpireDueSessions(ctx); err != nil {
			t.Fatalf("ExpireDueSessions: %v", err)
		}

		if err := svc.ApplyGatewayWebhook(ctx, orderID, session.PaymentSessionID, service.HostedPaymentSessionStatusSucceeded, ""); err != nil {
			t.Fatalf("ApplyGatewayWebhook: %v", err)
		}
		intent, err := queries.GetPaymentIntentByOrderID(ctx, orderID)
		if err != nil {
			t.Fatalf("GetPaymentIntentByOrderID: %v", err)
		}
		if intent.Status != service.HostedPaymentSessionStatusExpired {
			t.Fatalf("intent status: got %q want EXPIRED", intent.Status)
		}
	})
}

func TestPaymentService_CreateHostedPaymentSession_OneShot(t *testing.T) {
	svc, _ := newPaymentFixture(t)
	orderID := uuid.New()
	first := createTestSession(t, svc, orderID)
	second, err := svc.CreateHostedPaymentSession(t.Context(), service.CreateHostedPaymentSessionParams{
		OrderID:         orderID,
		BuyerUserID:     uuid.New(),
		MerchantID:      uuid.New(),
		TotalCents:      3000,
		Currency:        "USD",
		ShippingAddress: json.RawMessage(`{}`),
		ReturnURL:       first.ReturnURL,
	})
	if err != nil {
		t.Fatalf("repeat pending create: %v", err)
	}
	if second.PaymentSessionID != first.PaymentSessionID {
		t.Fatalf("session id rotated: %q vs %q", second.PaymentSessionID, first.PaymentSessionID)
	}

	if err := svc.ApplyGatewayWebhook(t.Context(), orderID, first.PaymentSessionID, service.HostedPaymentSessionStatusFailed, "declined"); err != nil {
		t.Fatalf("ApplyGatewayWebhook: %v", err)
	}
	_, err = svc.CreateHostedPaymentSession(t.Context(), service.CreateHostedPaymentSessionParams{
		OrderID:     orderID,
		BuyerUserID: uuid.New(),
		MerchantID:  uuid.New(),
		TotalCents:  3000,
		Currency:    "USD",
		ReturnURL:   first.ReturnURL,
	})
	if !errors.Is(err, service.ErrSessionTerminal) {
		t.Fatalf("expected ErrSessionTerminal, got %v", err)
	}
}

func TestPaymentService_CreateHostedPaymentSession_RequiresShipping(t *testing.T) {
	svc, _ := newPaymentFixture(t)
	_, err := svc.CreateHostedPaymentSession(t.Context(), service.CreateHostedPaymentSessionParams{
		OrderID:     uuid.New(),
		BuyerUserID: uuid.New(),
		MerchantID:  uuid.New(),
		TotalCents:  3000,
		Currency:    "USD",
		ReturnURL:   "/orders/x",
	})
	if !errors.Is(err, service.ErrInvalidSessionFacts) {
		t.Fatalf("expected ErrInvalidSessionFacts, got %v", err)
	}
}
