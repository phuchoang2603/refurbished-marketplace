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

func TestPaymentService_ApplyGatewayWebhook(t *testing.T) {
	t.Run("succeeded updates transaction writes outbox and ignores duplicate apply", func(t *testing.T) {
		svc, queries := newPaymentFixture(t)
		ctx := t.Context()

		orderID := uuid.New()
		buyerID := uuid.New()
		session, err := svc.CreateHostedPaymentSession(ctx, service.CreateHostedPaymentSessionParams{
			OrderID:         orderID,
			BuyerUserID:     buyerID,
			Currency:        "USD",
			ShippingAddress: json.RawMessage(`{}`),
			ReturnURL:       "/orders/" + orderID.String(),
			CancelURL:       "/orders/" + orderID.String(),
		})
		if err != nil {
			t.Fatalf("CreateHostedPaymentSession: %v", err)
		}

		_, err = queries.CreatePaymentTransaction(ctx, database.CreatePaymentTransactionParams{
			ID:             uuid.New(),
			OrderID:        orderID,
			MerchantID:     uuid.New(),
			AmountCents:    3000,
			Currency:       "USD",
			Status:         service.PaymentTxStatusInitialized,
			IdempotencyKey: "order:" + orderID.String(),
		})
		if err != nil {
			t.Fatalf("CreatePaymentTransaction: %v", err)
		}

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
		_, err := svc.CreateHostedPaymentSession(ctx, service.CreateHostedPaymentSessionParams{
			OrderID:         orderID,
			BuyerUserID:     uuid.New(),
			Currency:        "USD",
			ShippingAddress: json.RawMessage(`{}`),
			ReturnURL:       "/orders/" + orderID.String(),
			CancelURL:       "/orders/" + orderID.String(),
		})
		if err != nil {
			t.Fatalf("CreateHostedPaymentSession: %v", err)
		}

		past := time.Now().UTC().Add(-time.Minute)
		if err := queries.SetPaymentIntentExpiresAt(ctx, database.SetPaymentIntentExpiresAtParams{
			OrderID:   orderID,
			ExpiresAt: dberr.OptionalNullTime(past),
		}); err != nil {
			t.Fatalf("SetPaymentIntentExpiresAt: %v", err)
		}

		_, err = queries.CreatePaymentTransaction(ctx, database.CreatePaymentTransactionParams{
			ID:             uuid.New(),
			OrderID:        orderID,
			MerchantID:     uuid.New(),
			AmountCents:    3000,
			Currency:       "USD",
			Status:         service.PaymentTxStatusInitialized,
			IdempotencyKey: "order:" + orderID.String(),
		})
		if err != nil {
			t.Fatalf("CreatePaymentTransaction: %v", err)
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

	t.Run("expired pending without transaction leaves intent expired only", func(t *testing.T) {
		svc, queries := newPaymentFixture(t)
		ctx := t.Context()

		orderID := uuid.New()
		_, err := svc.CreateHostedPaymentSession(ctx, service.CreateHostedPaymentSessionParams{
			OrderID:         orderID,
			BuyerUserID:     uuid.New(),
			Currency:        "USD",
			ShippingAddress: json.RawMessage(`{}`),
			ReturnURL:       "/orders/" + orderID.String(),
			CancelURL:       "/orders/" + orderID.String(),
		})
		if err != nil {
			t.Fatalf("CreateHostedPaymentSession: %v", err)
		}

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

		outbox, err := queries.ListPaymentOutboxByAggregateID(ctx, orderID)
		if err != nil {
			t.Fatalf("ListPaymentOutboxByAggregateID: %v", err)
		}
		if len(outbox) != 0 {
			t.Fatalf("outbox rows: got %d want 0", len(outbox))
		}
	})

	t.Run("expired before inventory.reserved emits payment.failed on catch-up", func(t *testing.T) {
		svc, queries := newPaymentFixture(t)
		ctx := t.Context()

		orderID := uuid.New()
		merchantID := uuid.New()
		_, err := svc.CreateHostedPaymentSession(ctx, service.CreateHostedPaymentSessionParams{
			OrderID:         orderID,
			BuyerUserID:     uuid.New(),
			Currency:        "USD",
			ShippingAddress: json.RawMessage(`{}`),
			ReturnURL:       "/orders/" + orderID.String(),
			CancelURL:       "/orders/" + orderID.String(),
		})
		if err != nil {
			t.Fatalf("CreateHostedPaymentSession: %v", err)
		}

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
		if outbox[0].EventType != messaging.EventTypePaymentFailed {
			t.Fatalf("outbox event: got %q want %q", outbox[0].EventType, messaging.EventTypePaymentFailed)
		}

		// Retry after inbox ack must remain idempotent.
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
		session, err := svc.CreateHostedPaymentSession(ctx, service.CreateHostedPaymentSessionParams{
			OrderID:         orderID,
			BuyerUserID:     uuid.New(),
			Currency:        "USD",
			ShippingAddress: json.RawMessage(`{}`),
			ReturnURL:       "/orders/" + orderID.String(),
			CancelURL:       "/orders/" + orderID.String(),
		})
		if err != nil {
			t.Fatalf("CreateHostedPaymentSession: %v", err)
		}
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
