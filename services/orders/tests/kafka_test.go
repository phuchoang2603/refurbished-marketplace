package tests

import (
	"fmt"
	"testing"
	"time"

	"refurbished-marketplace/services/orders/internal/service"
	"refurbished-marketplace/shared/messaging"
	inventoryv1 "refurbished-marketplace/shared/proto/inventory/v1"
	paymentv1 "refurbished-marketplace/shared/proto/payment/v1"
	"refurbished-marketplace/shared/testutil"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

func TestKafkaPaymentResultHandler_EndToEnd(t *testing.T) {
	svc := newOrdersService(t)
	ctx := t.Context()

	buyerID := uuid.New()
	productID := uuid.New()
	merchantID := uuid.New()
	created, err := svc.CreateOrder(ctx, buyerID, merchantID, []service.OrderItemInput{{ProductID: productID, Quantity: 1, UnitPriceCents: 1000}}, 1000)
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
	topic := messaging.EventTypePaymentSucceeded
	payload, err := proto.Marshal(&paymentv1.PaymentOutcome{
		OrderId: created.ID.String(),
	})
	if err != nil {
		t.Fatal(err)
	}

	testutil.ProduceKafkaRecord(t, ctx, brokers, topic, payload)
	cancel, errRun := testutil.StartKafkaConsumer(t, ctx, brokers, fmt.Sprintf("orders-kafka-e2e-%s", uuid.New().String()), []string{messaging.EventTypePaymentSucceeded}, svc.KafkaOrderResultHandler())
	defer cancel()
	testutil.WaitForKafkaCondition(t, errRun, cancel, 30*time.Second, 200*time.Millisecond, "timeout waiting for order status PAID", func() (bool, error) {
		got, err := svc.GetOrderByID(ctx, created.ID)
		if err != nil {
			return false, fmt.Errorf("GetOrderByID: %w", err)
		}
		return got.Status == service.OrderStatusPaid, nil
	})
}

func TestKafkaInventoryReservationFailedHandler_EndToEnd(t *testing.T) {
	svc := newOrdersService(t)
	ctx := t.Context()

	buyerID := uuid.New()
	productID := uuid.New()
	merchantID := uuid.New()
	created, err := svc.CreateOrder(ctx, buyerID, merchantID, []service.OrderItemInput{{ProductID: productID, Quantity: 1, UnitPriceCents: 1000}}, 1000)
	if err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}

	k := testutil.SetupKafka(t)
	brokers, err := k.Brokers(ctx)
	if err != nil {
		t.Fatalf("Brokers: %v", err)
	}
	topic := messaging.EventTypeInventoryReservationFailed
	payload, err := proto.Marshal(&inventoryv1.InventoryReservationFailed{OrderId: created.ID.String()})
	if err != nil {
		t.Fatal(err)
	}

	testutil.ProduceKafkaRecord(t, ctx, brokers, topic, payload)
	cancel, errRun := testutil.StartKafkaConsumer(t, ctx, brokers, fmt.Sprintf("orders-kafka-inventory-failed-%s", uuid.New().String()), []string{topic}, svc.KafkaOrderResultHandler())
	defer cancel()
	testutil.WaitForKafkaCondition(t, errRun, cancel, 30*time.Second, 200*time.Millisecond, "timeout waiting for order status FAILED", func() (bool, error) {
		got, err := svc.GetOrderByID(ctx, created.ID)
		if err != nil {
			return false, fmt.Errorf("GetOrderByID: %w", err)
		}
		return got.Status == service.OrderStatusFailed, nil
	})
}
