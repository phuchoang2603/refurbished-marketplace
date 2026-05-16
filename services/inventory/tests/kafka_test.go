package tests

import (
	"fmt"
	"testing"
	"time"

	"refurbished-marketplace/shared/messaging"
	ordersv1 "refurbished-marketplace/shared/proto/orders/v1"
	paymentv1 "refurbished-marketplace/shared/proto/payment/v1"
	"refurbished-marketplace/shared/testutil"

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
	svc := newInventoryService(t)
	ctx := t.Context()

	firstProductID := uuid.New()
	secondProductID := uuid.New()
	merchantID := uuid.New()
	orderID := uuid.New()

	if _, err := svc.CreateInventory(ctx, firstProductID, 5); err != nil {
		t.Fatalf("CreateInventory first: %v", err)
	}
	if _, err := svc.CreateInventory(ctx, secondProductID, 4); err != nil {
		t.Fatalf("CreateInventory second: %v", err)
	}

	payload := orderCreatedPayload(
		orderID,
		merchantID,
		2500,
		&ordersv1.OrderCreatedItem{ProductId: firstProductID.String(), Quantity: 2},
		&ordersv1.OrderCreatedItem{ProductId: secondProductID.String(), Quantity: 1},
	)

	k := testutil.SetupKafka(t)
	brokers, err := k.Brokers(ctx)
	if err != nil {
		t.Fatalf("Brokers: %v", err)
	}
	topic := messaging.EventTypeOrderCreated

	testutil.ProduceKafkaRecord(t, ctx, brokers, topic, payload)
	cancel, errRun := testutil.StartKafkaConsumer(
		t,
		ctx,
		brokers,
		fmt.Sprintf("inventory-kafka-orders-created-%s", uuid.New().String()),
		[]string{topic},
		svc.KafkaReservationHandler(),
	)
	defer cancel()
	testutil.WaitForKafkaCondition(
		t,
		errRun,
		cancel,
		30*time.Second,
		200*time.Millisecond,
		"timeout waiting for inventory reservation",
		func() (bool, error) {
			firstInventory, err := svc.GetInventoryByProductID(ctx, firstProductID)
			if err != nil {
				return false, fmt.Errorf("GetInventoryByProductID first: %w", err)
			}
			secondInventory, err := svc.GetInventoryByProductID(ctx, secondProductID)
			if err != nil {
				return false, fmt.Errorf("GetInventoryByProductID second: %w", err)
			}
			return firstInventory.AvailableQty == 3 && firstInventory.ReservedQty == 2 && secondInventory.AvailableQty == 3 && secondInventory.ReservedQty == 1, nil
		},
	)
}

func TestKafkaPaymentOutcomeHandler_EndToEnd(t *testing.T) {
	svc := newInventoryService(t)
	ctx := t.Context()

	productID := uuid.New()
	merchantID := uuid.New()
	orderID := uuid.New()

	if _, err := svc.CreateInventory(ctx, productID, 4); err != nil {
		t.Fatalf("CreateInventory: %v", err)
	}
	if err := svc.HandleOrdersCreated(ctx, "orders.created/test/0/seed", orderCreatedPayload(
		orderID,
		merchantID,
		1000,
		&ordersv1.OrderCreatedItem{ProductId: productID.String(), Quantity: 2},
	)); err != nil {
		t.Fatalf("HandleOrdersCreated seed: %v", err)
	}

	k := testutil.SetupKafka(t)
	brokers, err := k.Brokers(ctx)
	if err != nil {
		t.Fatalf("Brokers: %v", err)
	}
	topic := messaging.EventTypePaymentSucceeded
	payload := paymentOutcomePayload(orderID)

	testutil.ProduceKafkaRecord(t, ctx, brokers, topic, payload)
	cancel, errRun := testutil.StartKafkaConsumer(
		t,
		ctx,
		brokers,
		fmt.Sprintf("inventory-kafka-payment-succeeded-%s", uuid.New().String()),
		[]string{topic},
		svc.KafkaReservationHandler(),
	)
	defer cancel()
	testutil.WaitForKafkaCondition(
		t,
		errRun,
		cancel,
		30*time.Second,
		200*time.Millisecond,
		"timeout waiting for committed inventory state",
		func() (bool, error) {
			inventoryRow, err := svc.GetInventoryByProductID(ctx, productID)
			if err != nil {
				return false, fmt.Errorf("GetInventoryByProductID: %w", err)
			}
			return inventoryRow.AvailableQty == 2 && inventoryRow.ReservedQty == 0, nil
		},
	)
}

func TestKafkaOrdersCreatedFailure_EndToEnd(t *testing.T) {
	svc := newInventoryService(t)
	ctx := t.Context()

	firstProductID := uuid.New()
	secondProductID := uuid.New()
	orderID := uuid.New()

	if _, err := svc.CreateInventory(ctx, firstProductID, 5); err != nil {
		t.Fatalf("CreateInventory first: %v", err)
	}
	if _, err := svc.CreateInventory(ctx, secondProductID, 1); err != nil {
		t.Fatalf("CreateInventory second: %v", err)
	}

	payload := orderCreatedPayload(
		orderID,
		uuid.New(),
		3000,
		&ordersv1.OrderCreatedItem{ProductId: firstProductID.String(), Quantity: 2},
		&ordersv1.OrderCreatedItem{ProductId: secondProductID.String(), Quantity: 2},
	)

	k := testutil.SetupKafka(t)
	brokers, err := k.Brokers(ctx)
	if err != nil {
		t.Fatalf("Brokers: %v", err)
	}
	topic := messaging.EventTypeOrderCreated

	testutil.ProduceKafkaRecord(t, ctx, brokers, topic, payload)
	cancel, errRun := testutil.StartKafkaConsumer(
		t,
		ctx,
		brokers,
		fmt.Sprintf("inventory-kafka-orders-failed-%s", uuid.New().String()),
		[]string{topic},
		svc.KafkaReservationHandler(),
	)
	defer cancel()
	testutil.WaitForKafkaCondition(
		t,
		errRun,
		cancel,
		30*time.Second,
		200*time.Millisecond,
		"timeout waiting for failed inventory reservation",
		func() (bool, error) {
			firstInventory, err := svc.GetInventoryByProductID(ctx, firstProductID)
			if err != nil {
				return false, fmt.Errorf("GetInventoryByProductID first: %w", err)
			}
			secondInventory, err := svc.GetInventoryByProductID(ctx, secondProductID)
			if err != nil {
				return false, fmt.Errorf("GetInventoryByProductID second: %w", err)
			}
			return firstInventory.AvailableQty == 5 && firstInventory.ReservedQty == 0 && secondInventory.AvailableQty == 1 && secondInventory.ReservedQty == 0, nil
		},
	)
}
