package tests

import (
	"context"
	"database/sql"
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

	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers":                  bootstrap,
		"broker.address.family":              "v4",
		"socket.connection.setup.timeout.ms": 60000,
	})
	if err != nil {
		t.Fatalf("NewProducer: %v", err)
	}
	defer p.Close()

	if err := p.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          payload,
	}, nil); err != nil {
		t.Fatalf("Produce: %v", err)
	}
	p.Flush(15 * 1000)

	deadline := time.Now().Add(45 * time.Second)
	for time.Now().Before(deadline) {
		row, err := queries.GetPaymentTransactionByOrderItemID(ctx, orderItemID)
		if err == nil && row.OrderID == orderID && row.Status == service.PaymentTxStatusInitialized {
			cancel()
			if err := <-errRun; err != nil && !errors.Is(err, context.Canceled) {
				t.Fatalf("consumer Run: %v", err)
			}
			return
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("GetPaymentTransactionByOrderItemID: %v", err)
		}
		time.Sleep(200 * time.Millisecond)
	}

	cancel()
	if err := <-errRun; err != nil && !errors.Is(err, context.Canceled) {
		t.Fatalf("consumer Run: %v", err)
	}
	t.Fatal("timeout waiting for payment transaction after Kafka message")
}
