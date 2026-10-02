package tests

import (
	"testing"

	"github.com/google/uuid"
	"github.com/phuchoang2603/refurbished-marketplace/shared/messaging"
	checkoutv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/checkout/v1"
	"google.golang.org/protobuf/proto"
)

func paymentCommandRecord(t *testing.T, checkoutID, orderID, topic, operation string, payload any) messaging.KafkaMessage {
	t.Helper()
	message := &checkoutv1.CheckoutMessage{Context: &checkoutv1.MessageContext{
		MessageId: uuid.NewString(), CheckoutId: checkoutID, OrderId: orderID,
		OperationId: uuid.NewSHA1(uuid.MustParse(checkoutID), []byte(operation)).String(), WorkflowVersion: 3,
	}}
	switch command := payload.(type) {
	case *checkoutv1.CreatePaymentSession:
		message.Payload = &checkoutv1.CheckoutMessage_CreatePaymentSession{CreatePaymentSession: command}
	case *checkoutv1.CancelPaymentSession:
		message.Payload = &checkoutv1.CheckoutMessage_CancelPaymentSession{CancelPaymentSession: command}
	default:
		t.Fatalf("unexpected checkout payment command %T", payload)
	}
	encoded, err := proto.Marshal(message)
	if err != nil {
		t.Fatal(err)
	}
	return messaging.KafkaMessage{Topic: topic, Value: encoded}
}

func createCheckoutPaymentRequest() *checkoutv1.CreatePaymentSession {
	checkoutID := uuid.New()
	return &checkoutv1.CreatePaymentSession{
		BuyerUserId: uuid.NewString(), BuyerEmail: "buyer@example.com", MerchantId: uuid.NewString(),
		Currency: "USD", TotalCents: 2000, ReturnUrl: "https://example.com/orders/return",
		ReservationOperationId: uuid.NewSHA1(checkoutID, []byte("reserve-stock")).String(),
		ShippingAddress:        &checkoutv1.CheckoutAddress{Line1: "1 Main St", City: "New York", PostalCode: "10001", Country: "US"},
		Items:                  []*checkoutv1.CheckoutLine{{ProductId: uuid.NewString(), Name: "Laptop", Quantity: 1, UnitPriceCents: 2000}},
	}
}

func TestCheckoutSessionCommandIdempotency(t *testing.T) {
	paymentService, queries := newPaymentFixture(t)
	checkoutID, orderID := uuid.NewString(), uuid.NewString()
	request := createCheckoutPaymentRequest()
	command := paymentCommandRecord(t, checkoutID, orderID, messaging.EventTypeCheckoutCreateSession, "create-session", request)
	if err := paymentService.HandleCheckoutCommand(t.Context(), command); err != nil {
		t.Fatal(err)
	}
	if err := paymentService.HandleCheckoutCommand(t.Context(), command); err != nil {
		t.Fatalf("replay original command: %v", err)
	}
	if err := paymentService.HandleCheckoutCommand(t.Context(), paymentCommandRecord(t, checkoutID, orderID, messaging.EventTypeCheckoutCreateSession, "create-session", request)); err != nil {
		t.Fatalf("replay operation with new message ID: %v", err)
	}
	original, err := paymentService.GetHostedPaymentSessionByOrder(t.Context(), uuid.MustParse(orderID))
	if err != nil || original.PaymentSessionID == "" {
		t.Fatalf("missing checkout session: %v %v", original, err)
	}
	transaction, err := queries.GetPaymentTransactionByOrderID(t.Context(), uuid.MustParse(orderID))
	if err != nil || transaction.AmountCents != 2000 {
		t.Fatalf("missing original payment transaction: %v %v", transaction, err)
	}
	changed := proto.Clone(request).(*checkoutv1.CreatePaymentSession)
	changed.BuyerEmail = "other@example.com"
	if err := paymentService.HandleCheckoutCommand(t.Context(), paymentCommandRecord(t, checkoutID, orderID, messaging.EventTypeCheckoutCreateSession, "create-session", changed)); err == nil {
		t.Fatal("changed checkout session facts must conflict")
	}
	same, err := paymentService.GetHostedPaymentSessionByOrder(t.Context(), uuid.MustParse(orderID))
	if err != nil || same.PaymentSessionID != original.PaymentSessionID {
		t.Fatalf("conflicting retry changed session: %v %v", same, err)
	}
}

func TestCheckoutCancelBeforeSessionCreation(t *testing.T) {
	paymentService, queries := newPaymentFixture(t)
	checkoutID, orderID := uuid.NewString(), uuid.NewString()
	cancel := paymentCommandRecord(t, checkoutID, orderID, messaging.EventTypeCheckoutCancelSession, "cancel-session", &checkoutv1.CancelPaymentSession{})
	if err := paymentService.HandleCheckoutCommand(t.Context(), cancel); err != nil {
		t.Fatal(err)
	}
	request := createCheckoutPaymentRequest()
	create := paymentCommandRecord(t, checkoutID, orderID, messaging.EventTypeCheckoutCreateSession, "create-session", request)
	if err := paymentService.HandleCheckoutCommand(t.Context(), create); err != nil {
		t.Fatal(err)
	}
	if _, err := queries.GetPaymentTransactionByOrderID(t.Context(), uuid.MustParse(orderID)); err == nil {
		t.Fatal("cancelled order unexpectedly created a transaction")
	}
}
