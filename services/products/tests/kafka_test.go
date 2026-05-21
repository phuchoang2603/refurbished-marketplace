package tests

import (
	"fmt"
	"testing"
	"time"

	"refurbished-marketplace/shared/messaging"
	ordersv1 "refurbished-marketplace/shared/proto/orders/v1"
	paymentv1 "refurbished-marketplace/shared/proto/payment/v1"
	testkafka "refurbished-marketplace/shared/testutil/kafka"

	"github.com/google/uuid"
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
	svc := newProductsService(t)
	ctx := t.Context()

	firstProduct, err := svc.CreateProduct(ctx, "Phone", "A", 1000, uuid.New(), 5)
	if err != nil {
		t.Fatalf("CreateProduct first: %v", err)
	}
	secondProduct, err := svc.CreateProduct(ctx, "Tablet", "B", 2000, uuid.New(), 4)
	if err != nil {
		t.Fatalf("CreateProduct second: %v", err)
	}

	merchantID := uuid.New()
	orderID := uuid.New()
	payload := orderCreatedPayload(
		orderID,
		merchantID,
		2500,
		&ordersv1.OrderCreatedItem{ProductId: firstProduct.ID.String(), Quantity: 2},
		&ordersv1.OrderCreatedItem{ProductId: secondProduct.ID.String(), Quantity: 1},
	)

	k := testkafka.SetupKafka(t)
	brokers, err := k.Brokers(ctx)
	if err != nil {
		t.Fatalf("Brokers: %v", err)
	}
	topic := messaging.EventTypeOrderCreated

	testkafka.ProduceKafkaRecord(t, ctx, brokers, topic, payload)
	cancel, errRun := testkafka.StartKafkaConsumer(t, ctx, brokers, fmt.Sprintf("products-kafka-orders-created-%s", uuid.New().String()), []string{topic}, svc.KafkaReservationHandler())
	defer cancel()
	testkafka.WaitForKafkaCondition(t, errRun, cancel, 30*time.Second, 200*time.Millisecond, "timeout waiting for inventory reservation", func() (bool, error) {
		firstInventory, err := svc.GetInventoryByProductID(ctx, firstProduct.ID)
		if err != nil {
			return false, fmt.Errorf("GetInventoryByProductID first: %w", err)
		}
		secondInventory, err := svc.GetInventoryByProductID(ctx, secondProduct.ID)
		if err != nil {
			return false, fmt.Errorf("GetInventoryByProductID second: %w", err)
		}
		return firstInventory.AvailableQty == 3 && firstInventory.ReservedQty == 2 && secondInventory.AvailableQty == 3 && secondInventory.ReservedQty == 1, nil
	})
}

func TestKafkaPaymentOutcomeHandler_EndToEnd(t *testing.T) {
	svc := newProductsService(t)
	ctx := t.Context()

	product, err := svc.CreateProduct(ctx, "Phone", "A", 1000, uuid.New(), 4)
	if err != nil {
		t.Fatalf("CreateProduct: %v", err)
	}
	merchantID := uuid.New()
	orderID := uuid.New()
	if err := svc.HandleOrdersCreated(ctx, "orders.created/test/0/seed", orderCreatedPayload(orderID, merchantID, 1000, &ordersv1.OrderCreatedItem{ProductId: product.ID.String(), Quantity: 2})); err != nil {
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
	cancel, errRun := testkafka.StartKafkaConsumer(t, ctx, brokers, fmt.Sprintf("products-kafka-payment-succeeded-%s", uuid.New().String()), []string{topic}, svc.KafkaReservationHandler())
	defer cancel()
	testkafka.WaitForKafkaCondition(t, errRun, cancel, 30*time.Second, 200*time.Millisecond, "timeout waiting for committed inventory state", func() (bool, error) {
		inventoryRow, err := svc.GetInventoryByProductID(ctx, product.ID)
		if err != nil {
			return false, fmt.Errorf("GetInventoryByProductID: %w", err)
		}
		return inventoryRow.AvailableQty == 2 && inventoryRow.ReservedQty == 0, nil
	})
}

func TestKafkaOrdersCreatedFailure_EndToEnd(t *testing.T) {
	svc := newProductsService(t)
	ctx := t.Context()

	firstProduct, err := svc.CreateProduct(ctx, "Phone", "A", 1000, uuid.New(), 5)
	if err != nil {
		t.Fatalf("CreateProduct first: %v", err)
	}
	secondProduct, err := svc.CreateProduct(ctx, "Tablet", "B", 2000, uuid.New(), 1)
	if err != nil {
		t.Fatalf("CreateProduct second: %v", err)
	}
	orderID := uuid.New()
	payload := orderCreatedPayload(
		orderID,
		uuid.New(),
		3000,
		&ordersv1.OrderCreatedItem{ProductId: firstProduct.ID.String(), Quantity: 2},
		&ordersv1.OrderCreatedItem{ProductId: secondProduct.ID.String(), Quantity: 2},
	)

	k := testkafka.SetupKafka(t)
	brokers, err := k.Brokers(ctx)
	if err != nil {
		t.Fatalf("Brokers: %v", err)
	}
	topic := messaging.EventTypeOrderCreated

	testkafka.ProduceKafkaRecord(t, ctx, brokers, topic, payload)
	cancel, errRun := testkafka.StartKafkaConsumer(t, ctx, brokers, fmt.Sprintf("products-kafka-orders-failed-%s", uuid.New().String()), []string{topic}, svc.KafkaReservationHandler())
	defer cancel()
	testkafka.WaitForKafkaCondition(t, errRun, cancel, 30*time.Second, 200*time.Millisecond, "timeout waiting for failed inventory reservation", func() (bool, error) {
		firstInventory, err := svc.GetInventoryByProductID(ctx, firstProduct.ID)
		if err != nil {
			return false, fmt.Errorf("GetInventoryByProductID first: %w", err)
		}
		secondInventory, err := svc.GetInventoryByProductID(ctx, secondProduct.ID)
		if err != nil {
			return false, fmt.Errorf("GetInventoryByProductID second: %w", err)
		}
		return firstInventory.AvailableQty == 5 && firstInventory.ReservedQty == 0 && secondInventory.AvailableQty == 1 && secondInventory.ReservedQty == 0, nil
	})
}
