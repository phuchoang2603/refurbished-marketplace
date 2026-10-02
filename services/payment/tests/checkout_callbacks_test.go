package tests

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/phuchoang2603/refurbished-marketplace/services/payment/internal/database"
	"github.com/phuchoang2603/refurbished-marketplace/services/payment/internal/service"
	"github.com/phuchoang2603/refurbished-marketplace/shared/err/dberr"
	"github.com/phuchoang2603/refurbished-marketplace/shared/messaging"
	checkoutv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/checkout/v1"
)

func TestCheckoutGatewayCallbacksAndExpiry(t *testing.T) {
	paymentService, queries := newPaymentFixture(t)
	checkoutID, orderID := uuid.NewString(), uuid.NewString()
	request := createCheckoutPaymentRequest()
	if err := paymentService.HandleCheckoutCommand(t.Context(), paymentCommandRecord(t, checkoutID, orderID, messaging.EventTypeCheckoutCreateSession, "create-session", request)); err != nil {
		t.Fatal(err)
	}
	session, err := paymentService.GetHostedPaymentSessionByOrder(t.Context(), uuid.MustParse(orderID))
	if err != nil {
		t.Fatal(err)
	}
	if err := queries.SetPaymentIntentExpiresAt(t.Context(), database.SetPaymentIntentExpiresAtParams{
		OrderID: uuid.MustParse(orderID), ExpiresAt: dberr.OptionalNullTime(time.Now().Add(-time.Minute)),
	}); err != nil {
		t.Fatal(err)
	}
	if err := paymentService.ExpireDueSessions(t.Context()); err != nil {
		t.Fatal(err)
	}
	stillPending, err := paymentService.GetHostedPaymentSessionByOrder(t.Context(), uuid.MustParse(orderID))
	if err != nil || stillPending.Status != service.HostedPaymentSessionStatusPending {
		t.Fatalf("checkout expiry must be reconciled by Checkout: %v %v", stillPending, err)
	}
	for range 2 {
		if err := paymentService.ApplyGatewayWebhook(t.Context(), uuid.MustParse(orderID), session.PaymentSessionID, service.HostedPaymentSessionStatusSucceeded, ""); err != nil {
			t.Fatal(err)
		}
	}
	transaction, err := queries.GetPaymentTransactionByOrderID(t.Context(), uuid.MustParse(orderID))
	if err != nil || transaction.Status != service.PaymentTxStatusSucceeded {
		t.Fatalf("successful payment not persisted: %v %v", transaction, err)
	}
	if err := paymentService.ApplyGatewayWebhook(t.Context(), uuid.MustParse(orderID), session.PaymentSessionID, service.HostedPaymentSessionStatusFailed, "contradictory provider callback"); err != nil {
		t.Fatal(err)
	}
	transaction, err = queries.GetPaymentTransactionByOrderID(t.Context(), uuid.MustParse(orderID))
	if err != nil || transaction.Status != service.PaymentTxStatusSucceeded {
		t.Fatalf("contradictory failure overwrote a charge: %v %v", transaction, err)
	}
	rows, err := queries.ListPaymentOutboxByAggregateID(t.Context(), uuid.MustParse(checkoutID))
	if err != nil || len(rows) != 3 {
		t.Fatalf("expected ready, success, contradictory failure once each: %d %v", len(rows), err)
	}
}

func TestCheckoutCancellationDoesNotGuessPaymentFailure(t *testing.T) {
	paymentService, queries := newPaymentFixture(t)
	checkoutID, orderID := uuid.NewString(), uuid.NewString()
	if err := paymentService.HandleCheckoutCommand(t.Context(), paymentCommandRecord(t, checkoutID, orderID, messaging.EventTypeCheckoutCreateSession, "create-session", createCheckoutPaymentRequest())); err != nil {
		t.Fatal(err)
	}
	if err := paymentService.HandleCheckoutCommand(t.Context(), paymentCommandRecord(t, checkoutID, orderID, messaging.EventTypeCheckoutCancelSession, "cancel-session", &checkoutv1.CancelPaymentSession{})); err != nil {
		t.Fatal(err)
	}
	transaction, err := queries.GetPaymentTransactionByOrderID(t.Context(), uuid.MustParse(orderID))
	if err != nil || transaction.Status != service.PaymentTxStatusInitialized {
		t.Fatalf("uncertain cancellation must not fail transaction: %v %v", transaction, err)
	}
	rows, err := queries.ListPaymentOutboxByAggregateID(t.Context(), uuid.MustParse(checkoutID))
	if err != nil || len(rows) != 2 || rows[1].EventType != messaging.EventTypeCheckoutPaymentUncertain {
		t.Fatalf("cancellation must report uncertain outcome: %v %v", rows, err)
	}
}
