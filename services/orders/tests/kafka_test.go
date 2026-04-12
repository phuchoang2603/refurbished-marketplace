package tests

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"refurbished-marketplace/services/orders/internal/service"
	"refurbished-marketplace/shared/messaging"
	"refurbished-marketplace/shared/testutil"
	"strings"
	"testing"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/google/uuid"
)

func TestKafkaPaymentResultHandler_EndToEnd(t *testing.T) {
	svc := newOrdersService(t)
	ctx := t.Context()

	buyerID := uuid.New()
	productID := uuid.New()
	merchantID := uuid.New()
	created, err := svc.CreateOrder(ctx, buyerID, []service.OrderItemInput{
		{ProductID: productID, MerchantID: merchantID, Quantity: 1, UnitPriceCents: 1000},
	}, 1000)
	if err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}
	if created.Status != service.OrderStatusPending {
		t.Fatalf("expected pending order, got %s", created.Status)
	}

	k := testutil.SetupKafka(t)
	brokers, err := k.Brokers(ctx)
	if err != nil {
		t.Fatalf("Brokers: %v", err)
	}
	bootstrap := strings.Join(brokers, ",")

	topic := messaging.EventTypePaymentItemSucceeded
	payload, err := json.Marshal(map[string]string{
		"order_id":      created.ID.String(),
		"order_item_id": created.Items[0].ID.String(),
	})
	if err != nil {
		t.Fatal(err)
	}

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
		GroupID:          fmt.Sprintf("orders-kafka-e2e-%s", uuid.New().String()),
		Topics:           []string{messaging.EventTypePaymentItemSucceeded},
		PreferIPv4:       true,
	}, svc.KafkaPaymentResultHandler())
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
			t.Fatal("timeout waiting for order status PAID")
		case <-ticker.C:
			got, err := svc.GetOrderByID(ctx, created.ID)
			if err != nil {
				t.Fatalf("GetOrderByID: %v", err)
			}
			if got.Status == service.OrderStatusPaid {
				return // Success!
			}
		}
	}
}
