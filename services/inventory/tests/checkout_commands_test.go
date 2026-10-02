package tests

import (
	"testing"

	"github.com/google/uuid"
	"github.com/phuchoang2603/refurbished-marketplace/shared/messaging"
	checkoutv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/checkout/v1"
	"google.golang.org/protobuf/proto"
)

func stockRecord(t *testing.T, checkoutID, orderID, topic, operation string, command any) messaging.KafkaMessage {
	t.Helper()
	message := &checkoutv1.CheckoutMessage{Context: &checkoutv1.MessageContext{
		MessageId: uuid.NewString(), CheckoutId: checkoutID, OrderId: orderID,
		OperationId: uuid.NewSHA1(uuid.MustParse(checkoutID), []byte(operation)).String(), WorkflowVersion: 2,
	}}
	switch payload := command.(type) {
	case *checkoutv1.ReserveStock:
		message.Payload = &checkoutv1.CheckoutMessage_ReserveStock{ReserveStock: payload}
	case *checkoutv1.CancelStock:
		message.Payload = &checkoutv1.CheckoutMessage_CancelStock{CancelStock: payload}
	case *checkoutv1.CommitStock:
		message.Payload = &checkoutv1.CheckoutMessage_CommitStock{CommitStock: payload}
	default:
		t.Fatalf("unrecognized checkout stock command %T", command)
	}
	encoded, err := proto.Marshal(message)
	if err != nil {
		t.Fatal(err)
	}
	return messaging.KafkaMessage{Topic: topic, Value: encoded}
}

func TestCheckoutReserveAllOrNothingAndReplay(t *testing.T) {
	inventoryService := newInventoryService(t)
	first, second := uuid.New(), uuid.New()
	for _, id := range []uuid.UUID{first, second} {
		if _, err := inventoryService.EnsureStock(t.Context(), id, 1); err != nil {
			t.Fatal(err)
		}
	}
	checkoutID, orderID := uuid.NewString(), uuid.NewString()
	command := &checkoutv1.ReserveStock{
		MerchantId: uuid.NewString(), TotalCents: 2000,
		Items: []*checkoutv1.CheckoutLine{
			{ProductId: first.String(), Name: "first", UnitPriceCents: 1000, Quantity: 1},
			{ProductId: second.String(), Name: "second", UnitPriceCents: 500, Quantity: 2},
		},
	}
	message := stockRecord(t, checkoutID, orderID, messaging.EventTypeCheckoutReserveStock, "reserve-stock", command)
	if err := inventoryService.HandleCheckoutCommand(t.Context(), message); err != nil {
		t.Fatal(err)
	}
	for _, id := range []uuid.UUID{first, second} {
		stock, err := inventoryService.GetInventoryByProductID(t.Context(), id)
		if err != nil || stock.AvailableQty != 1 || stock.ReservedQty != 0 {
			t.Fatalf("shortage must not partially hold product %s: %v %v", id, stock, err)
		}
	}
	if err := inventoryService.HandleCheckoutCommand(t.Context(), message); err != nil {
		t.Fatal(err)
	}
	freshID, freshOrder := uuid.NewString(), uuid.NewString()
	available := &checkoutv1.ReserveStock{
		MerchantId: uuid.NewString(), TotalCents: 1000,
		Items: []*checkoutv1.CheckoutLine{{ProductId: first.String(), Name: "first", UnitPriceCents: 1000, Quantity: 1}},
	}
	firstAttempt := stockRecord(t, freshID, freshOrder, messaging.EventTypeCheckoutReserveStock, "reserve-stock", available)
	if err := inventoryService.HandleCheckoutCommand(t.Context(), firstAttempt); err != nil {
		t.Fatal(err)
	}
	if err := inventoryService.HandleCheckoutCommand(t.Context(), stockRecord(t, freshID, freshOrder, messaging.EventTypeCheckoutReserveStock, "reserve-stock", available)); err != nil {
		t.Fatal(err)
	}
	stock, err := inventoryService.GetInventoryByProductID(t.Context(), first)
	if err != nil || stock.AvailableQty != 0 || stock.ReservedQty != 1 {
		t.Fatalf("retry double-reserved: %v %v", stock, err)
	}
	changed := proto.Clone(available).(*checkoutv1.ReserveStock)
	changed.Items[0].Name = "different"
	if err := inventoryService.HandleCheckoutCommand(t.Context(), stockRecord(t, freshID, freshOrder, messaging.EventTypeCheckoutReserveStock, "reserve-stock", changed)); err == nil {
		t.Fatal("mismatched reservation facts must conflict")
	}
}

func TestCheckoutCancelBeforeReserveFencesLateHold(t *testing.T) {
	inventoryService := newInventoryService(t)
	productID := uuid.New()
	if _, err := inventoryService.EnsureStock(t.Context(), productID, 1); err != nil {
		t.Fatal(err)
	}
	checkoutID, orderID := uuid.NewString(), uuid.NewString()
	cancel := stockRecord(t, checkoutID, orderID, messaging.EventTypeCheckoutCancelStock, "cancel-stock", &checkoutv1.CancelStock{})
	if err := inventoryService.HandleCheckoutCommand(t.Context(), cancel); err != nil {
		t.Fatal(err)
	}
	reserve := stockRecord(t, checkoutID, orderID, messaging.EventTypeCheckoutReserveStock, "reserve-stock", &checkoutv1.ReserveStock{
		MerchantId: uuid.NewString(), TotalCents: 1000,
		Items: []*checkoutv1.CheckoutLine{{ProductId: productID.String(), Name: "item", Quantity: 1, UnitPriceCents: 1000}},
	})
	if err := inventoryService.HandleCheckoutCommand(t.Context(), reserve); err != nil {
		t.Fatal(err)
	}
	stock, err := inventoryService.GetInventoryByProductID(t.Context(), productID)
	if err != nil || stock.AvailableQty != 1 || stock.ReservedQty != 0 {
		t.Fatalf("late reserve bypassed cancellation: %v %v", stock, err)
	}
}

func TestCheckoutCommitAndReleaseAreIdempotent(t *testing.T) {
	inventoryService := newInventoryService(t)
	productID := uuid.New()
	if _, err := inventoryService.EnsureStock(t.Context(), productID, 2); err != nil {
		t.Fatal(err)
	}
	for _, commit := range []bool{true, false} {
		checkoutID, orderID := uuid.NewString(), uuid.NewString()
		reserve := stockRecord(t, checkoutID, orderID, messaging.EventTypeCheckoutReserveStock, "reserve-stock", &checkoutv1.ReserveStock{
			MerchantId: uuid.NewString(), TotalCents: 1000,
			Items: []*checkoutv1.CheckoutLine{{ProductId: productID.String(), Name: "item", Quantity: 1, UnitPriceCents: 1000}},
		})
		if err := inventoryService.HandleCheckoutCommand(t.Context(), reserve); err != nil {
			t.Fatal(err)
		}
		topic, operation := messaging.EventTypeCheckoutCancelStock, "cancel-stock"
		var payload any = &checkoutv1.CancelStock{}
		if commit {
			topic, operation, payload = messaging.EventTypeCheckoutCommitStock, "commit-stock", &checkoutv1.CommitStock{}
		}
		for range 2 {
			if err := inventoryService.HandleCheckoutCommand(t.Context(), stockRecord(t, checkoutID, orderID, topic, operation, payload)); err != nil {
				t.Fatal(err)
			}
		}
		stock, err := inventoryService.GetInventoryByProductID(t.Context(), productID)
		if err != nil || stock.ReservedQty != 0 {
			t.Fatalf("stock not settled: %v %v", stock, err)
		}
		if stock.AvailableQty != 1 {
			t.Fatalf("settlement adjusted available stock twice: %v", stock)
		}
	}
}
