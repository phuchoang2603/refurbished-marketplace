package tests

import (
	"database/sql"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/phuchoang2603/refurbished-marketplace/shared/messaging"
	checkoutv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/checkout/v1"
	"google.golang.org/protobuf/proto"

	"github.com/phuchoang2603/refurbished-marketplace/services/orders/internal/service"
)

func checkoutOrderRecord(t *testing.T, checkoutID, orderID, operationID, messageID string, payload *checkoutv1.PrepareCheckout) messaging.KafkaMessage {
	t.Helper()
	command := &checkoutv1.CheckoutMessage{Context: &checkoutv1.MessageContext{
		CheckoutId: checkoutID, OrderId: orderID, OperationId: operationID,
		MessageId: messageID, WorkflowVersion: 1,
	}, Payload: &checkoutv1.CheckoutMessage_PrepareCheckout{PrepareCheckout: payload}}
	encoded, err := proto.Marshal(command)
	if err != nil {
		t.Fatal(err)
	}
	return messaging.KafkaMessage{Topic: messaging.EventTypeCheckoutPrepare, Value: encoded}
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
	operationID := uuid.NewSHA1(uuid.MustParse(checkoutID), []byte("prepare")).String()
	create := &checkoutv1.PrepareCheckout{
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
	conflicting := proto.Clone(create).(*checkoutv1.PrepareCheckout)
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
	create := &checkoutv1.PrepareCheckout{
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

func validPreparation() *checkoutv1.PrepareCheckout {
	return &checkoutv1.PrepareCheckout{
		BuyerUserId: uuid.NewString(), MerchantId: uuid.NewString(), Currency: "USD", TotalCents: 1500,
		Items: []*checkoutv1.CheckoutLine{{ProductId: uuid.NewString(), Name: "Phone", Quantity: 1, UnitPriceCents: 1500}},
	}
}

func resultTopics(t *testing.T, db *sql.DB, checkoutID string) []string {
	t.Helper()
	rows, err := db.QueryContext(t.Context(), "SELECT event_type FROM orders_outbox WHERE aggregate_id=$1 ORDER BY created_at, id", checkoutID)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	var topics []string
	for rows.Next() {
		var topic string
		if err := rows.Scan(&topic); err != nil {
			t.Fatal(err)
		}
		topics = append(topics, topic)
	}
	return topics
}

func TestFailedFinalizationBeforeCreationFencesLatePreparation(t *testing.T) {
	ordersService, db := newOrdersServiceWithDB(t)
	checkoutID, orderID := uuid.NewString(), uuid.NewString()
	failed := checkoutFinalizeRecord(t, checkoutID, orderID, uuid.NewString(), checkoutv1.FinalOrderStatus_FINAL_ORDER_STATUS_FAILED)
	if err := ordersService.HandleCheckoutCommand(t.Context(), failed); err != nil {
		t.Fatalf("failed finalization without an order must fence: %v", err)
	}
	late := checkoutOrderRecord(t, checkoutID, orderID, uuid.NewString(), uuid.NewString(), validPreparation())
	if err := ordersService.HandleCheckoutCommand(t.Context(), late); err != nil {
		t.Fatal(err)
	}
	if _, err := ordersService.GetOrderByID(t.Context(), uuid.MustParse(orderID)); err != service.ErrOrderNotFound {
		t.Fatalf("fenced checkout must not create an order: %v", err)
	}
	topics := resultTopics(t, db, checkoutID)
	if len(topics) != 2 || topics[0] != messaging.EventTypeCheckoutOrderFinalized || topics[1] != messaging.EventTypeCheckoutOrderRejected {
		t.Fatalf("result topics: %v", topics)
	}
	paid := checkoutFinalizeRecord(t, uuid.NewString(), uuid.NewString(), uuid.NewString(), checkoutv1.FinalOrderStatus_FINAL_ORDER_STATUS_PAID)
	if err := ordersService.HandleCheckoutCommand(t.Context(), paid); err == nil {
		t.Fatal("paid finalization requires an existing order")
	}
}

func TestInvalidPreparationIsDurablyRejected(t *testing.T) {
	ordersService, db := newOrdersServiceWithDB(t)
	checkoutID, orderID := uuid.NewString(), uuid.NewString()
	invalid := validPreparation()
	invalid.Items[0].Quantity = 0
	operationID := uuid.NewString()
	for range 2 {
		if err := ordersService.HandleCheckoutCommand(t.Context(), checkoutOrderRecord(t, checkoutID, orderID, operationID, uuid.NewString(), invalid)); err != nil {
			t.Fatal(err)
		}
	}
	topics := resultTopics(t, db, checkoutID)
	if len(topics) != 2 || topics[0] != messaging.EventTypeCheckoutOrderRejected || topics[1] != messaging.EventTypeCheckoutOrderRejected {
		t.Fatalf("result topics: %v", topics)
	}
	if _, err := ordersService.GetOrderByID(t.Context(), uuid.MustParse(orderID)); err != service.ErrOrderNotFound {
		t.Fatalf("rejected preparation must not create an order: %v", err)
	}
}

func TestConcurrentCreateAndFailedFinalizationNeverLeavesPendingOrder(t *testing.T) {
	for range 5 {
		ordersService := newOrdersService(t)
		checkoutID, orderID := uuid.NewString(), uuid.NewString()
		create := checkoutOrderRecord(t, checkoutID, orderID, uuid.NewString(), uuid.NewString(), validPreparation())
		failed := checkoutFinalizeRecord(t, checkoutID, orderID, uuid.NewString(), checkoutv1.FinalOrderStatus_FINAL_ORDER_STATUS_FAILED)
		var wait sync.WaitGroup
		for _, record := range []messaging.KafkaMessage{create, failed} {
			wait.Go(func() {
				if err := ordersService.HandleCheckoutCommand(t.Context(), record); err != nil {
					t.Errorf("handle: %v", err)
				}
			})
		}
		wait.Wait()
		order, err := ordersService.GetOrderByID(t.Context(), uuid.MustParse(orderID))
		if err == nil && order.Status != service.OrderStatusFailed {
			t.Fatalf("order left %s after failed finalization", order.Status)
		}
		if err != nil && err != service.ErrOrderNotFound {
			t.Fatal(err)
		}
	}
}
