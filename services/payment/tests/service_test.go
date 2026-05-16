package tests

import (
	"encoding/json"
	"errors"
	"testing"

	"refurbished-marketplace/services/payment/internal/database"
	"refurbished-marketplace/services/payment/internal/service"
	"refurbished-marketplace/shared/testutil"

	"github.com/google/uuid"
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

		_, err := queries.CreatePaymentTransaction(ctx, database.CreatePaymentTransactionParams{
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
