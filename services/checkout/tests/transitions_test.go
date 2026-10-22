package tests

import (
	"context"
	"database/sql"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/phuchoang2603/refurbished-marketplace/services/checkout/internal/service"
	"github.com/phuchoang2603/refurbished-marketplace/shared/messaging"
	checkoutv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/checkout/v1"
	"google.golang.org/protobuf/proto"
)

func submitForTransition(t *testing.T, db *sql.DB) (*service.Service, string, string) {
	t.Helper()
	svc := newTestService(db)
	checkoutID, orderID := submitWithService(t, db, svc)
	return svc, checkoutID, orderID
}

func submitWithService(t *testing.T, db *sql.DB, svc *service.Service) (string, string) {
	t.Helper()
	result, err := svc.SubmitCheckout(t.Context(), checkoutRequest())
	if err != nil {
		t.Fatal(err)
	}
	var orderID string
	if err := db.QueryRowContext(t.Context(), "SELECT order_id FROM checkouts WHERE id=$1", result.GetCheckoutId()).Scan(&orderID); err != nil {
		t.Fatal(err)
	}
	return result.GetCheckoutId(), orderID
}

func reserveForPayment(t *testing.T, svc *service.Service, checkoutID, orderID string) {
	t.Helper()
	sendResult(t, svc, checkoutID, orderID, messaging.EventTypeCheckoutOrderCreated, "create-order", 1, &checkoutv1.OrderCreated{})
	sendResult(t, svc, checkoutID, orderID, messaging.EventTypeCheckoutStockReserved, "reserve-stock", 2, &checkoutv1.StockReserved{})
}

func readyPayment(t *testing.T, db *sql.DB) (*service.Service, string, string) {
	t.Helper()
	svc := gatewayService(t, db, func(w http.ResponseWriter, r *http.Request) { writePayment(t, w) })
	checkoutID, orderID := submitWithService(t, db, svc)
	reserveForPayment(t, svc, checkoutID, orderID)
	if processed, err := svc.ProcessGatewayOperation(t.Context()); err != nil || !processed {
		t.Fatalf("create payment: %v %v", processed, err)
	}
	return svc, checkoutID, orderID
}

func sendResult(t *testing.T, svc *service.Service, checkoutID, orderID, topic, operation string, version uint64, payload any) messaging.KafkaMessage {
	t.Helper()
	message := buildResult(t, checkoutID, orderID, topic, operation, version, payload)
	if err := svc.HandleResult(context.Background(), message); err != nil {
		t.Fatalf("process %s: %v", topic, err)
	}
	return message
}

func buildResult(t *testing.T, checkoutID, orderID, topic, operation string, version uint64, payload any) messaging.KafkaMessage {
	t.Helper()
	message := &checkoutv1.CheckoutMessage{Context: &checkoutv1.MessageContext{MessageId: uuid.NewString(), CheckoutId: checkoutID, OrderId: orderID, OperationId: uuid.NewSHA1(uuid.MustParse(checkoutID), []byte(operation)).String(), WorkflowVersion: version}}
	switch value := payload.(type) {
	case *checkoutv1.OrderCreated:
		message.Payload = &checkoutv1.CheckoutMessage_OrderCreated{OrderCreated: value}
	case *checkoutv1.StockReserved:
		message.Payload = &checkoutv1.CheckoutMessage_StockReserved{StockReserved: value}
	case *checkoutv1.StockRejected:
		message.Payload = &checkoutv1.CheckoutMessage_StockRejected{StockRejected: value}
	case *checkoutv1.StockCommitted:
		message.Payload = &checkoutv1.CheckoutMessage_StockCommitted{StockCommitted: value}
	case *checkoutv1.StockReleased:
		message.Payload = &checkoutv1.CheckoutMessage_StockReleased{StockReleased: value}
	case *checkoutv1.OrderFinalized:
		message.Payload = &checkoutv1.CheckoutMessage_OrderFinalized{OrderFinalized: value}
	default:
		t.Fatalf("unexpected result %T", payload)
	}
	encoded, err := proto.Marshal(message)
	if err != nil {
		t.Fatal(err)
	}
	return messaging.KafkaMessage{Topic: topic, Value: encoded}
}

func assertCheckoutState(t *testing.T, db *sql.DB, checkoutID string, state checkoutv1.CheckoutState, commandCount int) {
	t.Helper()
	var actualState string
	var count int
	if err := db.QueryRowContext(t.Context(), "SELECT state FROM checkouts WHERE id=$1", checkoutID).Scan(&actualState); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(t.Context(), "SELECT count(*) FROM checkout_outbox WHERE aggregate_id=$1", checkoutID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if actualState != state.String() || count != commandCount {
		t.Fatalf("state %s commands %d, want %s/%d", actualState, count, state, commandCount)
	}
}

func TestResultReplayAndSuccessfulStockGate(t *testing.T) {
	db := newCheckoutDB(t)
	svc := gatewayService(t, db, func(w http.ResponseWriter, r *http.Request) { writePayment(t, w) })
	checkoutID, orderID := submitWithService(t, db, svc)
	future := buildResult(t, checkoutID, orderID, messaging.EventTypeCheckoutStockReserved, "reserve-stock", 2, &checkoutv1.StockReserved{})
	if err := svc.HandleResult(t.Context(), future); err == nil {
		t.Fatal("future result must be retried")
	}
	created := sendResult(t, svc, checkoutID, orderID, messaging.EventTypeCheckoutOrderCreated, "create-order", 1, &checkoutv1.OrderCreated{})
	if err := svc.HandleResult(t.Context(), created); err != nil {
		t.Fatal(err)
	}
	if err := svc.HandleResult(t.Context(), future); err != nil {
		t.Fatal(err)
	}
	sendResult(t, svc, checkoutID, orderID, messaging.EventTypeCheckoutStockReserved, "reserve-stock", 2, &checkoutv1.StockReserved{})
	assertCheckoutState(t, db, checkoutID, checkoutv1.CheckoutState_CHECKOUT_STATE_CREATING_PAYMENT, 2)
	if processed, err := svc.ProcessGatewayOperation(t.Context()); err != nil || !processed {
		t.Fatalf("operation: %v %v", processed, err)
	}
	assertCheckoutState(t, db, checkoutID, checkoutv1.CheckoutState_CHECKOUT_STATE_READY_TO_PAY, 2)
	sendWebhook(t, svc, "decline", "attempt.declined", orderID)
	assertCheckoutState(t, db, checkoutID, checkoutv1.CheckoutState_CHECKOUT_STATE_READY_TO_PAY, 2)
	request := webhookRequest(t, "success", "payment.succeeded", orderID)
	sendWebhook(t, svc, "success", "payment.succeeded", orderID)
	response, err := svc.HandleGatewayWebhook(t.Context(), request)
	if err != nil || response.GetResult() != checkoutv1.GatewayWebhookResult_GATEWAY_WEBHOOK_RESULT_DUPLICATE {
		t.Fatalf("duplicate: %v %v", response, err)
	}
	assertCheckoutState(t, db, checkoutID, checkoutv1.CheckoutState_CHECKOUT_STATE_COMMITTING_STOCK, 3)
	sendResult(t, svc, checkoutID, orderID, messaging.EventTypeCheckoutStockCommitted, "commit-stock", 5, &checkoutv1.StockCommitted{})
	assertCheckoutState(t, db, checkoutID, checkoutv1.CheckoutState_CHECKOUT_STATE_FINALIZING_PAID, 4)
	sendResult(t, svc, checkoutID, orderID, messaging.EventTypeCheckoutOrderFinalized, "finalize-paid", 6, &checkoutv1.OrderFinalized{Status: checkoutv1.FinalOrderStatus_FINAL_ORDER_STATUS_PAID})
	assertCheckoutState(t, db, checkoutID, checkoutv1.CheckoutState_CHECKOUT_STATE_COMPLETED, 4)
}

func TestExpiredPaymentWaitsForStockRelease(t *testing.T) {
	db := newCheckoutDB(t)
	svc, checkoutID, orderID := readyPayment(t, db)
	sendWebhook(t, svc, "expired", "payment.expired", orderID)
	assertCheckoutState(t, db, checkoutID, checkoutv1.CheckoutState_CHECKOUT_STATE_RELEASING_STOCK, 3)
	sendResult(t, svc, checkoutID, orderID, messaging.EventTypeCheckoutStockReleased, "cancel-stock", 5, &checkoutv1.StockReleased{})
	assertCheckoutState(t, db, checkoutID, checkoutv1.CheckoutState_CHECKOUT_STATE_FINALIZING_FAILED, 4)
	sendResult(t, svc, checkoutID, orderID, messaging.EventTypeCheckoutOrderFinalized, "finalize-failed", 6, &checkoutv1.OrderFinalized{Status: checkoutv1.FinalOrderStatus_FINAL_ORDER_STATUS_FAILED})
	assertCheckoutState(t, db, checkoutID, checkoutv1.CheckoutState_CHECKOUT_STATE_FAILED, 4)
}
