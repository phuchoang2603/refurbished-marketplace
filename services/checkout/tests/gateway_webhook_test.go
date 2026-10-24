package tests

import (
	"net/http"
	"testing"
	"time"

	"github.com/phuchoang2603/refurbished-marketplace/services/checkout/internal/service"
	checkoutv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/checkout/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGatewayWebhookRejectsBeforeDatabaseAccess(t *testing.T) {
	svc := service.New(nil, testConfig("http://gateway.invalid"), nil)
	for _, test := range []string{"missing", "tampered", "stale", "future", "invalid-json", "unknown-event", "duplicate-timestamp"} {
		t.Run(test, func(t *testing.T) {
			request := webhookRequest(t, "event-1", "payment.succeeded", "order-1")
			switch test {
			case "missing":
				request.SignatureHeader = ""
			case "tampered":
				request.RawBody = append(request.RawBody, ' ')
			case "stale":
				request.SignatureHeader = signWebhook(request.RawBody, "test-webhook-secret", time.Now().Add(-6*time.Minute))
			case "future":
				request.SignatureHeader = signWebhook(request.RawBody, "test-webhook-secret", time.Now().Add(6*time.Minute))
			case "invalid-json":
				request.RawBody = []byte("{")
				request.SignatureHeader = signWebhook(request.RawBody, "test-webhook-secret", time.Now())
			case "unknown-event":
				request = webhookRequest(t, "event-1", "payment.unknown", "order-1")
			case "duplicate-timestamp":
				request.SignatureHeader += ",t=123"
			}
			response, err := svc.HandleGatewayWebhook(t.Context(), request)
			if err != nil || response.GetResult() != checkoutv1.GatewayWebhookResult_GATEWAY_WEBHOOK_RESULT_REJECTED {
				t.Fatalf("rejection %v %v", response, err)
			}
		})
	}
}

func TestGatewayWebhookRollsBackInboxWhenOutboxFails(t *testing.T) {
	db := newCheckoutDB(t)
	svc, checkoutID, orderID := readyPayment(t, db)
	if _, err := db.ExecContext(t.Context(), `CREATE FUNCTION fail_command() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'simulated outbox failure'; END $$; CREATE TRIGGER fail_command BEFORE INSERT ON checkout_outbox FOR EACH ROW EXECUTE FUNCTION fail_command()`); err != nil {
		t.Fatal(err)
	}
	request := webhookRequest(t, "success", "payment.succeeded", orderID)
	if _, err := svc.HandleGatewayWebhook(t.Context(), request); status.Code(err) != codes.Unavailable {
		t.Fatalf("expected retryable store failure %v", err)
	}
	assertCheckoutState(t, db, checkoutID, checkoutv1.CheckoutState_CHECKOUT_STATE_READY_TO_PAY, 1)
	var inbox int
	if err := db.QueryRowContext(t.Context(), "SELECT count(*) FROM checkout_inbox WHERE message_id='gateway:success'").Scan(&inbox); err != nil || inbox != 0 {
		t.Fatalf("inbox must roll back %d %v", inbox, err)
	}
	if _, err := db.ExecContext(t.Context(), "DROP TRIGGER fail_command ON checkout_outbox"); err != nil {
		t.Fatal(err)
	}
	sendWebhook(t, svc, "success", "payment.succeeded", orderID)
	assertCheckoutState(t, db, checkoutID, checkoutv1.CheckoutState_CHECKOUT_STATE_COMMITTING_STOCK, 2)
}

func TestUnmatchedWebhookRetriesUntilPaymentRecorded(t *testing.T) {
	db := newCheckoutDB(t)
	svc := gatewayService(t, db, func(w http.ResponseWriter, r *http.Request) { writePayment(t, w) })
	checkoutID, orderID := submitWithService(t, db, svc)
	reserveForPayment(t, svc, checkoutID, orderID)
	request := webhookRequest(t, "early-expiry", "payment.expired", orderID)
	if response, err := svc.HandleGatewayWebhook(t.Context(), request); status.Code(err) != codes.Unavailable || response != nil {
		t.Fatalf("unrecorded payment must retry: response=%v error=%v", response, err)
	}
	var inbox int
	if err := db.QueryRowContext(t.Context(), "SELECT count(*) FROM checkout_inbox WHERE message_id='gateway:early-expiry'").Scan(&inbox); err != nil || inbox != 0 {
		t.Fatalf("unmatched webhook must not be deduplicated: count=%d error=%v", inbox, err)
	}
	assertCheckoutState(t, db, checkoutID, checkoutv1.CheckoutState_CHECKOUT_STATE_CREATING_PAYMENT, 1)
	if processed, err := svc.ProcessGatewayOperation(t.Context()); err != nil || !processed {
		t.Fatalf("record payment: processed=%v error=%v", processed, err)
	}
	response, err := svc.HandleGatewayWebhook(t.Context(), request)
	if err != nil || response.GetResult() != checkoutv1.GatewayWebhookResult_GATEWAY_WEBHOOK_RESULT_ACCEPTED {
		t.Fatalf("redelivery must apply: response=%v error=%v", response, err)
	}
	response, err = svc.HandleGatewayWebhook(t.Context(), request)
	if err != nil || response.GetResult() != checkoutv1.GatewayWebhookResult_GATEWAY_WEBHOOK_RESULT_DUPLICATE {
		t.Fatalf("applied event must deduplicate: response=%v error=%v", response, err)
	}
	assertCheckoutState(t, db, checkoutID, checkoutv1.CheckoutState_CHECKOUT_STATE_RELEASING_STOCK, 2)
}
