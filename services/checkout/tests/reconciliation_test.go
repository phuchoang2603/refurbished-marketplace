package tests

import (
	"testing"
	"time"

	"github.com/phuchoang2603/refurbished-marketplace/shared/messaging"
	checkoutv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/checkout/v1"
)

func TestPaymentDeadlineReconcilesRatherThanReleasing(t *testing.T) {
	db := newCheckoutDB(t)
	checkoutService, checkoutID, orderID := submitForTransition(t, db)
	sendResult(t, checkoutService, checkoutID, orderID, messaging.EventTypeCheckoutOrderCreated, "create-order", 1, &checkoutv1.OrderCreated{})
	sendResult(t, checkoutService, checkoutID, orderID, messaging.EventTypeCheckoutStockReserved, "reserve-stock", 2, &checkoutv1.StockReserved{})
	sendResult(t, checkoutService, checkoutID, orderID, messaging.EventTypeCheckoutSessionReady, "create-session", 3, &checkoutv1.PaymentSessionReady{
		PaymentSessionId: "session-1", ReturnUrl: "https://example.com/pay", ExpiresAtUnixMs: time.Now().Add(time.Hour).UnixMilli(),
	})
	if _, err := db.ExecContext(t.Context(), "UPDATE checkouts SET deadline_at = NOW() - INTERVAL '1 minute' WHERE id = $1", checkoutID); err != nil {
		t.Fatal(err)
	}
	if claimed, err := checkoutService.ProcessDue(t.Context()); err != nil || !claimed {
		t.Fatalf("payment deadline: claimed=%v err=%v", claimed, err)
	}
	assertCheckoutState(t, db, checkoutID, checkoutv1.CheckoutState_CHECKOUT_STATE_RECONCILING, 4)
	sendResult(t, checkoutService, checkoutID, orderID, messaging.EventTypeCheckoutPaymentUncertain, "cancel-session", 5, &checkoutv1.PaymentUncertain{PaymentSessionId: "session-1"})
	assertCheckoutState(t, db, checkoutID, checkoutv1.CheckoutState_CHECKOUT_STATE_NEEDS_REVIEW, 4)
	var exceptions int
	if err := db.QueryRowContext(t.Context(), "SELECT count(*) FROM checkout_exceptions WHERE checkout_id = $1", checkoutID).Scan(&exceptions); err != nil || exceptions != 1 {
		t.Fatalf("uncertain outcome must be recorded: count=%d err=%v", exceptions, err)
	}
	sendResult(t, checkoutService, checkoutID, orderID, messaging.EventTypeCheckoutPaymentSucceeded, "create-session", 3, &checkoutv1.PaymentSucceeded{PaymentSessionId: "session-1"})
	assertCheckoutState(t, db, checkoutID, checkoutv1.CheckoutState_CHECKOUT_STATE_COMMITTING_STOCK, 5)
}

func TestLatePaidAfterFailedRecordsFinancialException(t *testing.T) {
	db := newCheckoutDB(t)
	checkoutService, checkoutID, orderID := submitForTransition(t, db)
	sendResult(t, checkoutService, checkoutID, orderID, messaging.EventTypeCheckoutOrderCreated, "create-order", 1, &checkoutv1.OrderCreated{})
	sendResult(t, checkoutService, checkoutID, orderID, messaging.EventTypeCheckoutStockReserved, "reserve-stock", 2, &checkoutv1.StockReserved{})
	sendResult(t, checkoutService, checkoutID, orderID, messaging.EventTypeCheckoutSessionReady, "create-session", 3, &checkoutv1.PaymentSessionReady{
		PaymentSessionId: "session-1", ReturnUrl: "https://example.com/pay", ExpiresAtUnixMs: time.Now().Add(time.Hour).UnixMilli(),
	})
	sendResult(t, checkoutService, checkoutID, orderID, messaging.EventTypeCheckoutPaymentFailed, "create-session", 3, &checkoutv1.PaymentFailed{PaymentSessionId: "session-1"})
	sendResult(t, checkoutService, checkoutID, orderID, messaging.EventTypeCheckoutStockReleased, "cancel-stock", 5, &checkoutv1.StockReleased{})
	sendResult(t, checkoutService, checkoutID, orderID, messaging.EventTypeCheckoutOrderFinalized, "finalize-failed", 6, &checkoutv1.OrderFinalized{Status: checkoutv1.FinalOrderStatus_FINAL_ORDER_STATUS_FAILED})
	late := sendResult(t, checkoutService, checkoutID, orderID, messaging.EventTypeCheckoutPaymentSucceeded, "create-session", 3, &checkoutv1.PaymentSucceeded{PaymentSessionId: "session-1"})
	if err := checkoutService.HandleResult(t.Context(), late); err != nil {
		t.Fatal(err)
	}
	assertCheckoutState(t, db, checkoutID, checkoutv1.CheckoutState_CHECKOUT_STATE_FAILED, 5)
	var exceptions int
	if err := db.QueryRowContext(t.Context(), "SELECT count(*) FROM checkout_exceptions WHERE checkout_id = $1 AND kind = 'contradictory_payment'", checkoutID).Scan(&exceptions); err != nil || exceptions != 1 {
		t.Fatalf("late charge must require manual reconciliation: count=%d err=%v", exceptions, err)
	}
}
