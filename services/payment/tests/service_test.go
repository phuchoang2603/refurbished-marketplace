package tests

import (
	"encoding/json"
	"errors"
	"testing"

	"refurbished-marketplace/services/payment/internal/database"
	"refurbished-marketplace/services/payment/internal/service"
	testpostgres "refurbished-marketplace/shared/testutil/postgres"

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
