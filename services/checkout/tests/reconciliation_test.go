package tests

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
	"testing"

	checkoutv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/checkout/v1"
)

func TestPaymentDeadlineReadsOnceAndAcceptsLateWebhook(t *testing.T) {
	for _, outcome := range []string{"OPEN", "ERROR", "SUCCEEDED", "EXPIRED"} {
		t.Run(outcome, func(t *testing.T) {
			db := newCheckoutDB(t)
			var reads atomic.Int32
			var orderID string
			svc := gatewayService(t, db, func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodPost {
					writePayment(t, w)
					return
				}
				reads.Add(1)
				if outcome == "ERROR" {
					w.WriteHeader(http.StatusServiceUnavailable)
					return
				}
				if err := json.NewEncoder(w).Encode(map[string]any{"payment_id": "payment-1", "order_reference": orderID, "status": outcome}); err != nil {
					t.Error(err)
				}
			})
			checkoutID, createdOrderID := submitWithService(t, db, svc)
			orderID = createdOrderID
			reserveForPayment(t, svc, checkoutID, orderID)
			if _, err := svc.ProcessGatewayOperation(t.Context()); err != nil {
				t.Fatal(err)
			}
			if _, err := db.ExecContext(t.Context(), "UPDATE checkouts SET deadline_at=NOW()-INTERVAL '1 second' WHERE id=$1", checkoutID); err != nil {
				t.Fatal(err)
			}
			if claimed, err := svc.ProcessDue(t.Context()); err != nil || !claimed {
				t.Fatalf("deadline %v %v", claimed, err)
			}
			if claimed, err := svc.ProcessDue(t.Context()); err != nil || claimed {
				t.Fatalf("must not repeat status read %v %v", claimed, err)
			}
			if reads.Load() != 1 {
				t.Fatalf("status reads=%d", reads.Load())
			}
			if outcome == "SUCCEEDED" {
				assertCheckoutState(t, db, checkoutID, checkoutv1.CheckoutState_CHECKOUT_STATE_COMMITTING_STOCK, 2)
				return
			}
			if outcome == "EXPIRED" {
				assertCheckoutState(t, db, checkoutID, checkoutv1.CheckoutState_CHECKOUT_STATE_RELEASING_STOCK, 2)
				return
			}
			assertCheckoutState(t, db, checkoutID, checkoutv1.CheckoutState_CHECKOUT_STATE_NEEDS_REVIEW, 1)
			var exceptions int
			if err := db.QueryRowContext(t.Context(), "SELECT count(*) FROM checkout_exceptions WHERE resolved_at IS NULL").Scan(&exceptions); err != nil || exceptions != 1 {
				t.Fatalf("exception %d %v", exceptions, err)
			}
			sendWebhook(t, svc, "late-success", "payment.succeeded", orderID)
			assertCheckoutState(t, db, checkoutID, checkoutv1.CheckoutState_CHECKOUT_STATE_COMMITTING_STOCK, 2)
			if err := db.QueryRowContext(t.Context(), "SELECT count(*) FROM checkout_exceptions WHERE resolved_at IS NULL").Scan(&exceptions); err != nil || exceptions != 0 {
				t.Fatalf("resolved exception %d %v", exceptions, err)
			}
		})
	}
}

func TestLateSuccessAfterExpiryRecordsFinancialException(t *testing.T) {
	db := newCheckoutDB(t)
	svc, checkoutID, orderID := readyPayment(t, db)
	sendWebhook(t, svc, "expired", "payment.expired", orderID)
	sendWebhook(t, svc, "late-success", "payment.succeeded", orderID)
	response, err := svc.HandleGatewayWebhook(t.Context(), webhookRequest(t, "late-success", "payment.succeeded", orderID))
	if err != nil || response.GetResult() != checkoutv1.GatewayWebhookResult_GATEWAY_WEBHOOK_RESULT_DUPLICATE {
		t.Fatalf("replay: %v %v", response, err)
	}
	sendWebhook(t, svc, "repeated-expiry", "payment.expired", orderID)
	assertCheckoutState(t, db, checkoutID, checkoutv1.CheckoutState_CHECKOUT_STATE_NEEDS_REVIEW, 2)
	var exceptions int
	if err := db.QueryRowContext(t.Context(), "SELECT count(*) FROM checkout_exceptions WHERE kind='contradictory_payment'").Scan(&exceptions); err != nil || exceptions != 1 {
		t.Fatalf("financial exception %d %v", exceptions, err)
	}
}
