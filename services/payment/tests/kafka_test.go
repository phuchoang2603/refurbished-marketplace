package tests

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"refurbished-marketplace/services/payment/internal/service"
	"refurbished-marketplace/shared/messaging"
	"refurbished-marketplace/shared/testutil"
	"strings"
	"testing"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/google/uuid"
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
	bootstrap := strings.Join(brokers, ",")

	topic := messaging.EventTypeOrderItemCreated

	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers":                  bootstrap,
		"broker.address.family":              "v4",
		"socket.connection.setup.timeout.ms": 60000,
	})
	if err != nil {
		t.Fatalf("NewProducer: %v", err)
	}
	defer producer.Close()

	if err := producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          payload,
	}, nil); err != nil {
		t.Fatalf("Produce: %v", err)
	}
	producer.Flush(15 * 1000)

	consumer, err := messaging.NewKafkaConsumer(messaging.KafkaConsumerConfig{
		BootstrapServers: bootstrap,
		GroupID:          fmt.Sprintf("payment-kafka-e2e-%s", uuid.New().String()),
		Topics:           []string{topic},
		PreferIPv4:       true,
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

	ticker := time.NewTicker(200 * time.Millisecond)
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
				return
			}
		}
	}
}
