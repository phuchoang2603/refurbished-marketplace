package tests

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"refurbished-marketplace/services/payment/internal/service"
	"refurbished-marketplace/shared/messaging"
	inventoryv1 "refurbished-marketplace/shared/proto/inventory/v1"
	testkafka "refurbished-marketplace/shared/testutil/kafka"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

func inventoryReservedPayload(orderID, merchantID uuid.UUID, totalCents int64) []byte {
	msg := &inventoryv1.InventoryReserved{
		OrderId:    orderID.String(),
		MerchantId: merchantID.String(),
		TotalCents: totalCents,
	}
	b, err := proto.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return b
}

func TestKafkaInventoryReservedHandler_EndToEnd(t *testing.T) {
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

	merchantID := uuid.New()
	payload := inventoryReservedPayload(orderID, merchantID, 7500)

	k := testkafka.SetupKafka(t)
	brokers, err := k.Brokers(ctx)
	if err != nil {
		t.Fatalf("Brokers: %v", err)
	}
	topic := messaging.EventTypeInventoryReserved

	testkafka.ProduceKafkaRecord(t, ctx, brokers, topic, payload)
	cancel, errRun := testkafka.StartKafkaConsumer(t, ctx, brokers, fmt.Sprintf("payment-kafka-e2e-%s", uuid.New().String()), []string{topic}, svc.KafkaInventoryReservedHandler())
	defer cancel()
	testkafka.WaitForKafkaCondition(t, errRun, cancel, 30*time.Second, 500*time.Millisecond, "timeout waiting for payment transaction", func() (bool, error) {
		row, err := queries.GetPaymentTransactionByOrderID(ctx, orderID)
		if err != nil {
			return false, nil
		}
		return row.OrderID == orderID && row.Status == service.PaymentTxStatusInitialized, nil
	})
}
