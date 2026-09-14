package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/phuchoang2603/refurbished-marketplace/services/inventory/internal/service"
	"github.com/phuchoang2603/refurbished-marketplace/shared/messaging"
	ordersv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/orders/v1"
	paymentv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/payment/v1"
	testkafka "github.com/phuchoang2603/refurbished-marketplace/shared/testutil/kafka"

	"github.com/google/uuid"
	productsv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/products/v1"
	"google.golang.org/protobuf/proto"
)

func orderCreatedPayload(orderID, merchantID uuid.UUID, totalCents int64, items ...*ordersv1.OrderCreatedItem) []byte {
	msg := &ordersv1.OrderCreated{
		OrderId:     orderID.String(),
		BuyerUserId: uuid.New().String(),
		MerchantId:  merchantID.String(),
		TotalCents:  totalCents,
		Items:       items,
	}
	b, err := proto.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return b
}

func paymentOutcomePayload(orderID uuid.UUID) []byte {
	msg := &paymentv1.PaymentOutcome{OrderId: orderID.String()}
	b, err := proto.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return b
}

func TestKafkaOrdersCreatedHandler_EndToEnd(t *testing.T) {
	svc := newInventoryService(t)
	ctx := t.Context()

	firstID := uuid.New()
	if _, err := svc.EnsureStock(ctx, firstID, 5); err != nil {
		t.Fatalf("EnsureStock first: %v", err)
	}
	secondID := uuid.New()
	if _, err := svc.EnsureStock(ctx, secondID, 4); err != nil {
		t.Fatalf("EnsureStock second: %v", err)
	}

	merchantID := uuid.New()
	orderID := uuid.New()
	payload := orderCreatedPayload(
		orderID,
		merchantID,
		2500,
		&ordersv1.OrderCreatedItem{ProductId: firstID.String(), Quantity: 2},
		&ordersv1.OrderCreatedItem{ProductId: secondID.String(), Quantity: 1},
	)

	k := testkafka.SetupKafka(t)
	brokers, err := k.Brokers(ctx)
	if err != nil {
		t.Fatalf("Brokers: %v", err)
	}
	topic := messaging.EventTypeOrderCreated

	testkafka.ProduceKafkaRecord(t, ctx, brokers, topic, payload)
	cancel, errRun := testkafka.StartKafkaConsumer(t, ctx, brokers, fmt.Sprintf("inventory-kafka-orders-created-%s", uuid.New().String()), []string{topic}, svc.KafkaReservationHandler())
	defer cancel()
	testkafka.WaitForKafkaCondition(t, errRun, cancel, 30*time.Second, 200*time.Millisecond, "timeout waiting for inventory reservation", func() (bool, error) {
		firstInventory, err := svc.GetInventoryByProductID(ctx, firstID)
		if err != nil {
			return false, fmt.Errorf("GetInventoryByProductID first: %w", err)
		}
		secondInventory, err := svc.GetInventoryByProductID(ctx, secondID)
		if err != nil {
			return false, fmt.Errorf("GetInventoryByProductID second: %w", err)
		}
		return firstInventory.AvailableQty == 3 && firstInventory.ReservedQty == 2 && secondInventory.AvailableQty == 3 && secondInventory.ReservedQty == 1, nil
	})
}

func TestKafkaPaymentOutcomeHandler_EndToEnd(t *testing.T) {
	svc := newInventoryService(t)
	ctx := t.Context()

	productID := uuid.New()
	if _, err := svc.EnsureStock(ctx, productID, 4); err != nil {
		t.Fatalf("EnsureStock: %v", err)
	}
	merchantID := uuid.New()
	orderID := uuid.New()
	if err := svc.HandleOrdersCreated(ctx, "orders.created/test/0/seed", orderCreatedPayload(orderID, merchantID, 1000, &ordersv1.OrderCreatedItem{ProductId: productID.String(), Quantity: 2})); err != nil {
		t.Fatalf("HandleOrdersCreated seed: %v", err)
	}

	k := testkafka.SetupKafka(t)
	brokers, err := k.Brokers(ctx)
	if err != nil {
		t.Fatalf("Brokers: %v", err)
	}
	topic := messaging.EventTypePaymentSucceeded
	payload := paymentOutcomePayload(orderID)

	testkafka.ProduceKafkaRecord(t, ctx, brokers, topic, payload)
	cancel, errRun := testkafka.StartKafkaConsumer(t, ctx, brokers, fmt.Sprintf("inventory-kafka-payment-succeeded-%s", uuid.New().String()), []string{topic}, svc.KafkaReservationHandler())
	defer cancel()
	testkafka.WaitForKafkaCondition(t, errRun, cancel, 30*time.Second, 200*time.Millisecond, "timeout waiting for committed inventory state", func() (bool, error) {
		inventoryRow, err := svc.GetInventoryByProductID(ctx, productID)
		if err != nil {
			return false, fmt.Errorf("GetInventoryByProductID: %w", err)
		}
		return inventoryRow.AvailableQty == 2 && inventoryRow.ReservedQty == 0, nil
	})
}

func TestKafkaOrdersCreatedFailure_EndToEnd(t *testing.T) {
	db := newInventoryDB(t)
	svc := service.New(db)
	ctx := t.Context()

	firstID := uuid.New()
	if _, err := svc.EnsureStock(ctx, firstID, 5); err != nil {
		t.Fatalf("EnsureStock first: %v", err)
	}
	secondID := uuid.New()
	if _, err := svc.EnsureStock(ctx, secondID, 1); err != nil {
		t.Fatalf("EnsureStock second: %v", err)
	}
	orderID := uuid.New()
	payload := orderCreatedPayload(
		orderID,
		uuid.New(),
		3000,
		&ordersv1.OrderCreatedItem{ProductId: firstID.String(), Quantity: 2},
		&ordersv1.OrderCreatedItem{ProductId: secondID.String(), Quantity: 2},
	)

	k := testkafka.SetupKafka(t)
	brokers, err := k.Brokers(ctx)
	if err != nil {
		t.Fatalf("Brokers: %v", err)
	}
	topic := messaging.EventTypeOrderCreated

	testkafka.ProduceKafkaRecord(t, ctx, brokers, topic, payload)
	cancel, errRun := testkafka.StartKafkaConsumer(t, ctx, brokers, fmt.Sprintf("inventory-kafka-orders-failed-%s", uuid.New().String()), []string{topic}, svc.KafkaReservationHandler())
	defer cancel()
	testkafka.WaitForKafkaCondition(t, errRun, cancel, 30*time.Second, 200*time.Millisecond, "timeout waiting for failed inventory reservation", func() (bool, error) {
		var failures int
		if err := db.QueryRowContext(ctx, "SELECT count(*) FROM inventory_outbox WHERE aggregate_id = $1 AND event_type = $2", orderID, messaging.EventTypeInventoryReservationFailed).Scan(&failures); err != nil {
			return false, err
		}
		if failures == 0 {
			return false, nil
		}
		if failures != 1 {
			return false, fmt.Errorf("failure events = %d, want 1", failures)
		}
		firstInventory, err := svc.GetInventoryByProductID(ctx, firstID)
		if err != nil {
			return false, fmt.Errorf("GetInventoryByProductID first: %w", err)
		}
		secondInventory, err := svc.GetInventoryByProductID(ctx, secondID)
		if err != nil {
			return false, fmt.Errorf("GetInventoryByProductID second: %w", err)
		}
		return firstInventory.AvailableQty == 5 && firstInventory.ReservedQty == 0 && secondInventory.AvailableQty == 1 && secondInventory.ReservedQty == 0, nil
	})
}

func TestKafkaProductCreatedIndependentConsumers(t *testing.T) {
	db := newInventoryDB(t)
	svc := service.New(db)
	event := creationEvent(proto.Int32(5))
	payload, err := proto.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	k := testkafka.SetupKafka(t)
	brokers, err := k.Brokers(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	topic := messaging.EventTypeProductCreated
	testkafka.ProduceKafkaRecord(t, t.Context(), brokers, topic, payload)
	cancel, runErr := testkafka.StartKafkaConsumer(t, t.Context(), brokers, "inventory-created-"+uuid.NewString(), []string{topic}, svc.KafkaProductCreatedHandler())
	defer cancel()
	testkafka.WaitForKafkaCondition(t, runErr, cancel, 30*time.Second, 200*time.Millisecond, "creation was not consumed", func() (bool, error) {
		var count int
		err := db.QueryRowContext(t.Context(), "SELECT count(*) FROM inventory_inbox WHERE message_id=$1", topic+"/"+event.EventId).Scan(&count)
		return count == 1, err
	})
	assertStock(t, svc, uuid.MustParse(event.ProductId), 5, 0)
	// A later independent group must still receive the original creation event.
	observed := make(chan []byte, 1)
	observerCancel, observerErr := testkafka.StartKafkaConsumer(t, t.Context(), brokers, "catalog-observer-"+uuid.NewString(), []string{topic}, func(_ context.Context, msg messaging.KafkaMessage) error {
		select {
		case observed <- msg.Value:
		default:
		}
		return nil
	})
	defer observerCancel()
	testkafka.WaitForKafkaCondition(t, observerErr, observerCancel, 30*time.Second, 200*time.Millisecond, "independent group did not receive creation", func() (bool, error) {
		select {
		case value := <-observed:
			var got productsv1.ProductCreated
			if err := messaging.UnmarshalKafkaProtobuf(value, &got); err != nil {
				return false, err
			}
			if !proto.Equal(event, &got) {
				return false, fmt.Errorf("independent consumer received different event")
			}
			return true, nil
		default:
			return false, nil
		}
	})
}
