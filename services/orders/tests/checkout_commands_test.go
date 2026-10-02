package tests

import (
	"testing"

	"github.com/google/uuid"
	"github.com/phuchoang2603/refurbished-marketplace/shared/messaging"
	checkoutv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/checkout/v1"
	"google.golang.org/protobuf/proto"

	"github.com/phuchoang2603/refurbished-marketplace/services/orders/internal/service"
)

func checkoutOrderRecord(t *testing.T, checkoutID, orderID, operationID, messageID string, payload *checkoutv1.CreateOrder) messaging.KafkaMessage {
	t.Helper()
	command := &checkoutv1.CheckoutMessage{Context: &checkoutv1.MessageContext{
		CheckoutId: checkoutID, OrderId: orderID, OperationId: operationID,
		MessageId: messageID, WorkflowVersion: 1,
	}, Payload: &checkoutv1.CheckoutMessage_CreateOrder{CreateOrder: payload}}
	encoded, err := proto.Marshal(command)
	if err != nil {
		t.Fatal(err)
	}
	return messaging.KafkaMessage{Topic: messaging.EventTypeCheckoutCreateOrder, Value: encoded}
}

func checkoutFinalizeRecord(t *testing.T, checkoutID, orderID, operationID string, status checkoutv1.FinalOrderStatus) messaging.KafkaMessage {
	t.Helper()
	command := &checkoutv1.CheckoutMessage{Context: &checkoutv1.MessageContext{
		CheckoutId: checkoutID, OrderId: orderID, OperationId: operationID,
		MessageId: uuid.NewString(), WorkflowVersion: 6,
	}, Payload: &checkoutv1.CheckoutMessage_FinalizeOrder{FinalizeOrder: &checkoutv1.FinalizeOrder{Status: status}}}
	encoded, err := proto.Marshal(command)
	if err != nil {
		t.Fatal(err)
	}
	return messaging.KafkaMessage{Topic: messaging.EventTypeCheckoutFinalizeOrder, Value: encoded}
}

func TestCheckoutOrderCreateReplaysWithoutMutation(t *testing.T) {
	ordersService := newOrdersService(t)
	checkoutID, orderID := uuid.NewString(), uuid.NewString()
	operationID := uuid.NewSHA1(uuid.MustParse(checkoutID), []byte("create-order")).String()
	create := &checkoutv1.CreateOrder{
		BuyerUserId: uuid.NewString(), MerchantId: uuid.NewString(), Currency: "USD", TotalCents: 2500,
		Items: []*checkoutv1.CheckoutLine{{ProductId: uuid.NewString(), Name: "Laptop", Quantity: 1, UnitPriceCents: 2500}},
	}
	first := checkoutOrderRecord(t, checkoutID, orderID, operationID, uuid.NewString(), create)
	if err := ordersService.HandleCheckoutCommand(t.Context(), first); err != nil {
		t.Fatal(err)
	}
	if err := ordersService.HandleCheckoutCommand(t.Context(), first); err != nil {
		t.Fatalf("same message should be idempotent: %v", err)
	}
	retry := checkoutOrderRecord(t, checkoutID, orderID, operationID, uuid.NewString(), create)
	if err := ordersService.HandleCheckoutCommand(t.Context(), retry); err != nil {
		t.Fatalf("same operation retried with new delivery: %v", err)
	}
	order, err := ordersService.GetOrderByID(t.Context(), uuid.MustParse(orderID))
	if err != nil || order.Status != service.OrderStatusPending || len(order.Items) != 1 {
		t.Fatalf("pending order must remain immutable: %v, %v", order, err)
	}
	conflicting := proto.Clone(create).(*checkoutv1.CreateOrder)
	conflicting.Items[0].Name = "Not a laptop"
	conflict := checkoutOrderRecord(t, checkoutID, orderID, operationID, uuid.NewString(), conflicting)
	if err := ordersService.HandleCheckoutCommand(t.Context(), conflict); err != service.ErrIdempotencyConflict {
		t.Fatalf("changed snapshot must conflict: %v", err)
	}
	order, err = ordersService.GetOrderByID(t.Context(), uuid.MustParse(orderID))
	if err != nil || order.Items[0].UnitPriceCents != 2500 {
		t.Fatalf("conflict changed order: %v, %v", order, err)
	}
}

func TestCheckoutFinalizationRequiresCommandAndRejectsTerminalConflict(t *testing.T) {
	ordersService := newOrdersService(t)
	checkoutID, orderID := uuid.NewString(), uuid.NewString()
	create := &checkoutv1.CreateOrder{
		BuyerUserId: uuid.NewString(), MerchantId: uuid.NewString(), Currency: "USD", TotalCents: 1000,
		Items: []*checkoutv1.CheckoutLine{{ProductId: uuid.NewString(), Name: "Laptop", Quantity: 1, UnitPriceCents: 1000}},
	}
	if err := ordersService.HandleCheckoutCommand(t.Context(), checkoutOrderRecord(t, checkoutID, orderID, uuid.NewString(), uuid.NewString(), create)); err != nil {
		t.Fatal(err)
	}
	order, err := ordersService.GetOrderByID(t.Context(), uuid.MustParse(orderID))
	if err != nil || order.Status != service.OrderStatusPending {
		t.Fatalf("created order should remain pending: %v %v", order, err)
	}
	if _, err := ordersService.UpdateOrderStatus(t.Context(), uuid.MustParse(orderID), service.OrderStatusPaid); err != service.ErrOrderNotPayable {
		t.Fatalf("legacy status write must not finalize checkout order: %v", err)
	}
	paid := checkoutFinalizeRecord(t, checkoutID, orderID, uuid.NewString(), checkoutv1.FinalOrderStatus_FINAL_ORDER_STATUS_PAID)
	if err := ordersService.HandleCheckoutCommand(t.Context(), paid); err != nil {
		t.Fatal(err)
	}
	if err := ordersService.HandleCheckoutCommand(t.Context(), paid); err != nil {
		t.Fatalf("duplicate finalization: %v", err)
	}
	order, err = ordersService.GetOrderByID(t.Context(), uuid.MustParse(orderID))
	if err != nil || order.Status != service.OrderStatusPaid {
		t.Fatalf("order should be paid: %v %v", order, err)
	}
	failed := checkoutFinalizeRecord(t, checkoutID, orderID, uuid.NewString(), checkoutv1.FinalOrderStatus_FINAL_ORDER_STATUS_FAILED)
	if err := ordersService.HandleCheckoutCommand(t.Context(), failed); err == nil {
		t.Fatal("paid order must reject failed finalization")
	}
	order, err = ordersService.GetOrderByID(t.Context(), uuid.MustParse(orderID))
	if err != nil || order.Status != service.OrderStatusPaid {
		t.Fatalf("conflict changed paid order: %v %v", order, err)
	}
}
