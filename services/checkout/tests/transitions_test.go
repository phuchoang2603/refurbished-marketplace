package tests

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/phuchoang2603/refurbished-marketplace/services/checkout/internal/service"
	"github.com/phuchoang2603/refurbished-marketplace/shared/messaging"
	checkoutv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/checkout/v1"
	"google.golang.org/protobuf/proto"
)

func submitForTransition(t *testing.T, db *sql.DB) (*service.Service, string, string) {
	t.Helper()
	checkoutService := service.New(db)
	result, err := checkoutService.SubmitCheckout(t.Context(), checkoutRequest())
	if err != nil {
		t.Fatal(err)
	}
	var orderID string
	if err := db.QueryRowContext(t.Context(), "SELECT order_id FROM checkouts WHERE id = $1", result.GetCheckoutId()).Scan(&orderID); err != nil {
		t.Fatal(err)
	}
	return checkoutService, result.GetCheckoutId(), orderID
}

func sendResult(t *testing.T, checkoutService *service.Service, checkoutID, orderID, topic, operation string, version uint64, payload any) messaging.KafkaMessage {
	t.Helper()
	kafkaMessage := buildResult(t, checkoutID, orderID, topic, operation, version, payload)
	if err := checkoutService.HandleResult(context.Background(), kafkaMessage); err != nil {
		t.Fatalf("process %s: %v", topic, err)
	}
	return kafkaMessage
}

func buildResult(t *testing.T, checkoutID, orderID, topic, operation string, version uint64, payload any) messaging.KafkaMessage {
	t.Helper()
	checkoutUUID := uuid.MustParse(checkoutID)
	message := &checkoutv1.CheckoutMessage{Context: &checkoutv1.MessageContext{
		MessageId: uuid.NewString(), CheckoutId: checkoutID, OrderId: orderID,
		OperationId: uuid.NewSHA1(checkoutUUID, []byte(operation)).String(), WorkflowVersion: version,
	}}
	switch value := payload.(type) {
	case *checkoutv1.OrderCreated:
		message.Payload = &checkoutv1.CheckoutMessage_OrderCreated{OrderCreated: value}
	case *checkoutv1.StockReserved:
		message.Payload = &checkoutv1.CheckoutMessage_StockReserved{StockReserved: value}
	case *checkoutv1.StockRejected:
		message.Payload = &checkoutv1.CheckoutMessage_StockRejected{StockRejected: value}
	case *checkoutv1.PaymentSessionReady:
		message.Payload = &checkoutv1.CheckoutMessage_PaymentSessionReady{PaymentSessionReady: value}
	case *checkoutv1.PaymentSucceeded:
		message.Payload = &checkoutv1.CheckoutMessage_PaymentSucceeded{PaymentSucceeded: value}
	case *checkoutv1.PaymentFailed:
		message.Payload = &checkoutv1.CheckoutMessage_PaymentFailed{PaymentFailed: value}
	case *checkoutv1.PaymentCancelled:
		message.Payload = &checkoutv1.CheckoutMessage_PaymentCancelled{PaymentCancelled: value}
	case *checkoutv1.PaymentUncertain:
		message.Payload = &checkoutv1.CheckoutMessage_PaymentUncertain{PaymentUncertain: value}
	case *checkoutv1.StockCommitted:
		message.Payload = &checkoutv1.CheckoutMessage_StockCommitted{StockCommitted: value}
	case *checkoutv1.StockReleased:
		message.Payload = &checkoutv1.CheckoutMessage_StockReleased{StockReleased: value}
	case *checkoutv1.OrderFinalized:
		message.Payload = &checkoutv1.CheckoutMessage_OrderFinalized{OrderFinalized: value}
	default:
		t.Fatalf("unexpected result payload %T", payload)
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
	var actualCount int
	if err := db.QueryRowContext(t.Context(), "SELECT state FROM checkouts WHERE id = $1", checkoutID).Scan(&actualState); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(t.Context(), "SELECT count(*) FROM checkout_outbox WHERE aggregate_id = $1", checkoutID).Scan(&actualCount); err != nil {
		t.Fatal(err)
	}
	if actualState != state.String() || actualCount != commandCount {
		t.Fatalf("state=%s commands=%d, want %s/%d", actualState, actualCount, state, commandCount)
	}
}

func TestResultReplayAndSuccessfulStockGate(t *testing.T) {
	db := newCheckoutDB(t)
	checkoutService, checkoutID, orderID := submitForTransition(t, db)
	outOfOrder := buildResult(t, checkoutID, orderID, messaging.EventTypeCheckoutStockReserved, "reserve-stock", 2, &checkoutv1.StockReserved{})
	if err := checkoutService.HandleResult(t.Context(), outOfOrder); err == nil {
		t.Fatal("future result must be retried rather than acknowledged")
	}
	assertCheckoutState(t, db, checkoutID, checkoutv1.CheckoutState_CHECKOUT_STATE_CREATING_ORDER, 1)
	created := sendResult(t, checkoutService, checkoutID, orderID, messaging.EventTypeCheckoutOrderCreated, "create-order", 1, &checkoutv1.OrderCreated{})
	assertCheckoutState(t, db, checkoutID, checkoutv1.CheckoutState_CHECKOUT_STATE_RESERVING_STOCK, 2)
	if err := checkoutService.HandleResult(t.Context(), created); err != nil {
		t.Fatal(err)
	}
	assertCheckoutState(t, db, checkoutID, checkoutv1.CheckoutState_CHECKOUT_STATE_RESERVING_STOCK, 2)
	if err := checkoutService.HandleResult(t.Context(), outOfOrder); err != nil {
		t.Fatal(err)
	}
	sendResult(t, checkoutService, checkoutID, orderID, messaging.EventTypeCheckoutStockReserved, "reserve-stock", 2, &checkoutv1.StockReserved{})
	assertCheckoutState(t, db, checkoutID, checkoutv1.CheckoutState_CHECKOUT_STATE_CREATING_SESSION, 3)
	sendResult(t, checkoutService, checkoutID, orderID, messaging.EventTypeCheckoutSessionReady, "create-session", 3, &checkoutv1.PaymentSessionReady{
		PaymentSessionId: "session-1", ReturnUrl: "https://example.com/pay", ExpiresAtUnixMs: time.Now().Add(30 * time.Minute).UnixMilli(),
	})
	assertCheckoutState(t, db, checkoutID, checkoutv1.CheckoutState_CHECKOUT_STATE_READY_TO_PAY, 3)
	sendResult(t, checkoutService, checkoutID, orderID, messaging.EventTypeCheckoutPaymentSucceeded, "create-session", 3, &checkoutv1.PaymentSucceeded{PaymentSessionId: "session-1"})
	assertCheckoutState(t, db, checkoutID, checkoutv1.CheckoutState_CHECKOUT_STATE_COMMITTING_STOCK, 4)
	sendResult(t, checkoutService, checkoutID, orderID, messaging.EventTypeCheckoutStockCommitted, "commit-stock", 5, &checkoutv1.StockCommitted{})
	assertCheckoutState(t, db, checkoutID, checkoutv1.CheckoutState_CHECKOUT_STATE_FINALIZING_PAID, 5)
	sendResult(t, checkoutService, checkoutID, orderID, messaging.EventTypeCheckoutOrderFinalized, "finalize-paid", 6, &checkoutv1.OrderFinalized{Status: checkoutv1.FinalOrderStatus_FINAL_ORDER_STATUS_PAID})
	assertCheckoutState(t, db, checkoutID, checkoutv1.CheckoutState_CHECKOUT_STATE_COMPLETED, 5)
}

func TestFailedPaymentWaitsForStockRelease(t *testing.T) {
	db := newCheckoutDB(t)
	checkoutService, checkoutID, orderID := submitForTransition(t, db)
	sendResult(t, checkoutService, checkoutID, orderID, messaging.EventTypeCheckoutOrderCreated, "create-order", 1, &checkoutv1.OrderCreated{})
	sendResult(t, checkoutService, checkoutID, orderID, messaging.EventTypeCheckoutStockReserved, "reserve-stock", 2, &checkoutv1.StockReserved{})
	sendResult(t, checkoutService, checkoutID, orderID, messaging.EventTypeCheckoutSessionReady, "create-session", 3, &checkoutv1.PaymentSessionReady{
		PaymentSessionId: "session-1", ReturnUrl: "https://example.com/pay", ExpiresAtUnixMs: time.Now().Add(30 * time.Minute).UnixMilli(),
	})
	sendResult(t, checkoutService, checkoutID, orderID, messaging.EventTypeCheckoutPaymentFailed, "create-session", 3, &checkoutv1.PaymentFailed{PaymentSessionId: "session-1"})
	assertCheckoutState(t, db, checkoutID, checkoutv1.CheckoutState_CHECKOUT_STATE_RELEASING_STOCK, 4)
	sendResult(t, checkoutService, checkoutID, orderID, messaging.EventTypeCheckoutStockReleased, "cancel-stock", 5, &checkoutv1.StockReleased{})
	assertCheckoutState(t, db, checkoutID, checkoutv1.CheckoutState_CHECKOUT_STATE_FINALIZING_FAILED, 5)
	sendResult(t, checkoutService, checkoutID, orderID, messaging.EventTypeCheckoutOrderFinalized, "finalize-failed", 6, &checkoutv1.OrderFinalized{Status: checkoutv1.FinalOrderStatus_FINAL_ORDER_STATUS_FAILED})
	assertCheckoutState(t, db, checkoutID, checkoutv1.CheckoutState_CHECKOUT_STATE_FAILED, 5)
}
