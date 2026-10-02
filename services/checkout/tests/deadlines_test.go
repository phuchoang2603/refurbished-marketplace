package tests

import (
	"sync"
	"testing"

	"github.com/phuchoang2603/refurbished-marketplace/services/checkout/internal/service"
	"github.com/phuchoang2603/refurbished-marketplace/shared/messaging"
	checkoutv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/checkout/v1"
	"google.golang.org/protobuf/proto"
)

func TestDeadlineRetryAfterWorkerRestart(t *testing.T) {
	db := newCheckoutDB(t)
	_, checkoutID, _ := submitForTransition(t, db)
	if _, err := db.ExecContext(t.Context(), "UPDATE checkouts SET deadline_at = NOW() - INTERVAL '1 minute' WHERE id = $1", checkoutID); err != nil {
		t.Fatal(err)
	}
	var workers sync.WaitGroup
	processed := make(chan bool, 2)
	for range 2 {
		workers.Go(func() {
			restarted := service.New(db)
			claimed, err := restarted.ProcessDue(t.Context())
			if err != nil {
				t.Errorf("process due: %v", err)
			}
			processed <- claimed
		})
	}
	workers.Wait()
	close(processed)
	var claims int
	for claim := range processed {
		if claim {
			claims++
		}
	}
	if claims != 1 {
		t.Fatalf("expected one worker to claim deadline, got %d", claims)
	}
	assertCheckoutState(t, db, checkoutID, checkoutv1.CheckoutState_CHECKOUT_STATE_CREATING_ORDER, 2)
	rows, err := db.QueryContext(t.Context(), "SELECT event_type, payload FROM checkout_outbox WHERE aggregate_id = $1", checkoutID)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var operationID string
	seen := map[string]bool{}
	for rows.Next() {
		var topic string
		var payload []byte
		if err := rows.Scan(&topic, &payload); err != nil {
			t.Fatal(err)
		}
		if topic != messaging.EventTypeCheckoutCreateOrder {
			t.Fatalf("retry topic: %s", topic)
		}
		message := new(checkoutv1.CheckoutMessage)
		if err := proto.Unmarshal(payload, message); err != nil {
			t.Fatal(err)
		}
		if operationID != "" && operationID != message.GetContext().GetOperationId() {
			t.Fatal("retry must preserve operation identity")
		}
		operationID = message.GetContext().GetOperationId()
		if seen[message.GetContext().GetMessageId()] {
			t.Fatal("retry must use a distinct message id")
		}
		seen[message.GetContext().GetMessageId()] = true
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if len(seen) != 2 {
		t.Fatalf("expected two deliveries, got %d", len(seen))
	}
}

func TestReservationDeadlineFencesLateHold(t *testing.T) {
	db := newCheckoutDB(t)
	checkoutService, checkoutID, orderID := submitForTransition(t, db)
	sendResult(t, checkoutService, checkoutID, orderID, messaging.EventTypeCheckoutOrderCreated, "create-order", 1, &checkoutv1.OrderCreated{})
	if _, err := db.ExecContext(t.Context(), "UPDATE checkouts SET deadline_at = NOW() - INTERVAL '1 minute' WHERE id = $1", checkoutID); err != nil {
		t.Fatal(err)
	}
	claimed, err := service.New(db).ProcessDue(t.Context())
	if err != nil || !claimed {
		t.Fatalf("reservation timeout: claimed=%t, err=%v", claimed, err)
	}
	assertCheckoutState(t, db, checkoutID, checkoutv1.CheckoutState_CHECKOUT_STATE_RELEASING_STOCK, 3)
	sendResult(t, checkoutService, checkoutID, orderID, messaging.EventTypeCheckoutStockReserved, "reserve-stock", 2, &checkoutv1.StockReserved{})
	assertCheckoutState(t, db, checkoutID, checkoutv1.CheckoutState_CHECKOUT_STATE_RELEASING_STOCK, 3)
	sendResult(t, checkoutService, checkoutID, orderID, messaging.EventTypeCheckoutStockReleased, "cancel-stock", 3, &checkoutv1.StockReleased{})
	assertCheckoutState(t, db, checkoutID, checkoutv1.CheckoutState_CHECKOUT_STATE_FINALIZING_FAILED, 4)
}
