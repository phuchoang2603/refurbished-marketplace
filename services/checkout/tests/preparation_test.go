package tests

import (
	"database/sql"
	"net/http"
	"sync"
	"testing"

	"github.com/phuchoang2603/refurbished-marketplace/shared/messaging"
	checkoutv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/checkout/v1"
)

func countGatewayOperations(t *testing.T, db *sql.DB, checkoutID string) int {
	t.Helper()
	var count int
	if err := db.QueryRowContext(t.Context(), "SELECT count(*) FROM checkout_gateway_operations WHERE checkout_id=$1", checkoutID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count
}

func TestSubmitPublishesOneSharedPreparationCommand(t *testing.T) {
	db := newCheckoutDB(t)
	_, checkoutID, _ := submitForTransition(t, db)
	assertCheckoutState(t, db, checkoutID, checkoutv1.CheckoutState_CHECKOUT_STATE_PREPARING, 1)
	var topic, outcomeOrders, outcomeInventory string
	if err := db.QueryRowContext(t.Context(), "SELECT o.event_type, c.orders_outcome, c.inventory_outcome FROM checkout_outbox o JOIN checkouts c ON c.id=o.aggregate_id WHERE c.id=$1", checkoutID).Scan(&topic, &outcomeOrders, &outcomeInventory); err != nil {
		t.Fatal(err)
	}
	if topic != messaging.EventTypeCheckoutPrepare || outcomeOrders != "PENDING" || outcomeInventory != "PENDING" {
		t.Fatalf("topic %s orders %s inventory %s", topic, outcomeOrders, outcomeInventory)
	}
}

func TestPreparationJoinAcceptsEitherArrivalOrder(t *testing.T) {
	for name, first := range map[string]string{"orders first": "orders", "inventory first": "inventory"} {
		t.Run(name, func(t *testing.T) {
			db := newCheckoutDB(t)
			svc, checkoutID, orderID := submitForTransition(t, db)
			send := func(participant string) {
				if participant == "orders" {
					sendResult(t, svc, checkoutID, orderID, messaging.EventTypeCheckoutOrderCreated, "prepare", 1, &checkoutv1.OrderCreated{})
					return
				}
				sendResult(t, svc, checkoutID, orderID, messaging.EventTypeCheckoutStockReserved, "prepare", 1, &checkoutv1.StockReserved{})
			}
			send(first)
			assertCheckoutState(t, db, checkoutID, checkoutv1.CheckoutState_CHECKOUT_STATE_PREPARING, 1)
			if countGatewayOperations(t, db, checkoutID) != 0 {
				t.Fatal("payment must wait for both branches")
			}
			if first == "orders" {
				send("inventory")
			} else {
				send("orders")
			}
			assertCheckoutState(t, db, checkoutID, checkoutv1.CheckoutState_CHECKOUT_STATE_CREATING_PAYMENT, 1)
			if countGatewayOperations(t, db, checkoutID) != 1 {
				t.Fatal("expected exactly one gateway operation")
			}
		})
	}
}

func TestConcurrentPreparationResultsOpenPaymentOnce(t *testing.T) {
	db := newCheckoutDB(t)
	svc, checkoutID, orderID := submitForTransition(t, db)
	orders := buildResult(t, checkoutID, orderID, messaging.EventTypeCheckoutOrderCreated, "prepare", 1, &checkoutv1.OrderCreated{})
	stock := buildResult(t, checkoutID, orderID, messaging.EventTypeCheckoutStockReserved, "prepare", 1, &checkoutv1.StockReserved{})
	var wait sync.WaitGroup
	for _, message := range []messaging.KafkaMessage{orders, stock, orders, stock} {
		wait.Go(func() {
			if err := svc.HandleResult(t.Context(), message); err != nil {
				t.Errorf("handle result: %v", err)
			}
		})
	}
	wait.Wait()
	assertCheckoutState(t, db, checkoutID, checkoutv1.CheckoutState_CHECKOUT_STATE_CREATING_PAYMENT, 1)
	if countGatewayOperations(t, db, checkoutID) != 1 {
		t.Fatal("expected exactly one gateway operation")
	}
}

func TestPreparationRejectionCancelsStockBeforeFailingOrder(t *testing.T) {
	cases := map[string]struct {
		topic   string
		payload any
	}{
		"orders rejected": {messaging.EventTypeCheckoutOrderRejected, &checkoutv1.OrderRejected{Reason: "order invalid"}},
		"stock rejected":  {messaging.EventTypeCheckoutStockRejected, &checkoutv1.StockRejected{Reason: "insufficient stock"}},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			db := newCheckoutDB(t)
			svc, checkoutID, orderID := submitForTransition(t, db)
			sendResult(t, svc, checkoutID, orderID, tc.topic, "prepare", 1, tc.payload)
			assertCheckoutState(t, db, checkoutID, checkoutv1.CheckoutState_CHECKOUT_STATE_RELEASING_STOCK, 2)
			sendResult(t, svc, checkoutID, orderID, messaging.EventTypeCheckoutOrderCreated, "prepare", 1, &checkoutv1.OrderCreated{})
			sendResult(t, svc, checkoutID, orderID, messaging.EventTypeCheckoutStockReserved, "prepare", 1, &checkoutv1.StockReserved{})
			assertCheckoutState(t, db, checkoutID, checkoutv1.CheckoutState_CHECKOUT_STATE_RELEASING_STOCK, 2)
			if countGatewayOperations(t, db, checkoutID) != 0 {
				t.Fatal("late success must not open payment")
			}
			sendResult(t, svc, checkoutID, orderID, messaging.EventTypeCheckoutStockReleased, "cancel-stock", 2, &checkoutv1.StockReleased{})
			assertCheckoutState(t, db, checkoutID, checkoutv1.CheckoutState_CHECKOUT_STATE_FINALIZING_FAILED, 3)
			sendResult(t, svc, checkoutID, orderID, messaging.EventTypeCheckoutOrderFinalized, "finalize-failed", 3, &checkoutv1.OrderFinalized{Status: checkoutv1.FinalOrderStatus_FINAL_ORDER_STATUS_FAILED})
			assertCheckoutState(t, db, checkoutID, checkoutv1.CheckoutState_CHECKOUT_STATE_FAILED, 3)
		})
	}
}

func TestPreparationRejectsConflictingCorrelationAndOutcomes(t *testing.T) {
	db := newCheckoutDB(t)
	svc, checkoutID, orderID := submitForTransition(t, db)
	sendResult(t, svc, checkoutID, orderID, messaging.EventTypeCheckoutOrderCreated, "prepare", 2, &checkoutv1.OrderCreated{})
	sendResult(t, svc, checkoutID, orderID, messaging.EventTypeCheckoutStockReserved, "other-operation", 1, &checkoutv1.StockReserved{})
	assertCheckoutState(t, db, checkoutID, checkoutv1.CheckoutState_CHECKOUT_STATE_PREPARING, 1)
	sendResult(t, svc, checkoutID, orderID, messaging.EventTypeCheckoutOrderCreated, "prepare", 1, &checkoutv1.OrderCreated{})
	sendResult(t, svc, checkoutID, orderID, messaging.EventTypeCheckoutOrderRejected, "prepare", 1, &checkoutv1.OrderRejected{Reason: "contradiction"})
	assertCheckoutState(t, db, checkoutID, checkoutv1.CheckoutState_CHECKOUT_STATE_PREPARING, 1)
	var exceptions int
	if err := db.QueryRowContext(t.Context(), "SELECT count(*) FROM checkout_exceptions WHERE checkout_id=$1", checkoutID).Scan(&exceptions); err != nil || exceptions != 3 {
		t.Fatalf("exceptions %d %v", exceptions, err)
	}
	sendResult(t, svc, checkoutID, orderID, messaging.EventTypeCheckoutStockReserved, "prepare", 1, &checkoutv1.StockReserved{})
	assertCheckoutState(t, db, checkoutID, checkoutv1.CheckoutState_CHECKOUT_STATE_CREATING_PAYMENT, 1)
}

func TestPreparationRestartAfterOneBranchCommits(t *testing.T) {
	db := newCheckoutDB(t)
	svc := gatewayService(t, db, func(w http.ResponseWriter, r *http.Request) { writePayment(t, w) })
	checkoutID, orderID := submitWithService(t, db, svc)
	sendResult(t, svc, checkoutID, orderID, messaging.EventTypeCheckoutOrderCreated, "prepare", 1, &checkoutv1.OrderCreated{})
	restarted := gatewayService(t, db, func(w http.ResponseWriter, r *http.Request) { writePayment(t, w) })
	sendResult(t, restarted, checkoutID, orderID, messaging.EventTypeCheckoutStockReserved, "prepare", 1, &checkoutv1.StockReserved{})
	assertCheckoutState(t, db, checkoutID, checkoutv1.CheckoutState_CHECKOUT_STATE_CREATING_PAYMENT, 1)
}
