package tests

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"refurbished-marketplace/services/payment/internal/service"
	"refurbished-marketplace/shared/messaging"
	"refurbished-marketplace/shared/testutil"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/twmb/franz-go/pkg/kgo"
)

func TestKafkaOrdersItemCreatedHandler_EndToEnd(t *testing.T) {
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

	orderItemID := uuid.New()
	merchantID := uuid.New()
	payload := orderItemCreatedPayload(orderID, orderItemID, merchantID, 7500)

	k := testutil.SetupKafka(t)
	brokers, err := k.Brokers(ctx)
	if err != nil {
		t.Fatalf("Brokers: %v", err)
	}
	topic := messaging.EventTypeOrderItemCreated

	prod, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.AllowAutoTopicCreation(),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	defer prod.Close()

	res := prod.ProduceSync(ctx, &kgo.Record{Topic: topic, Value: payload})
	if err := res.FirstErr(); err != nil {
		t.Fatalf("ProduceSync: %v", err)
	}

	consumer, err := messaging.NewKafkaConsumer(messaging.KafkaConsumerConfig{
		BootstrapServers: brokers,
		GroupID:          fmt.Sprintf("payment-kafka-e2e-%s", uuid.New().String()),
		Topics:           []string{topic},
	}, svc.KafkaOrdersItemCreatedHandler())
	if err != nil {
		t.Fatalf("NewKafkaConsumer: %v", err)
	}
	defer func() { _ = consumer.Close() }()

	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	errRun := make(chan error, 1)
	go func() {
		errRun <- consumer.Run(runCtx)
	}()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	timeout := time.After(30 * time.Second)

	for {
		select {
		case err := <-errRun:
			if err != nil && !errors.Is(err, context.Canceled) {
				t.Fatalf("Consumer exited unexpectedly: %v", err)
			}
			return
		case <-timeout:
			t.Fatal("timeout waiting for payment transaction")
		case <-ticker.C:
			row, err := queries.GetPaymentTransactionByOrderItemID(ctx, orderItemID)
			if err == nil && row.OrderID == orderID && row.Status == service.PaymentTxStatusInitialized {
				cancel()
				return
			}
		}
	}
}
