package tests

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/phuchoang2603/refurbished-marketplace/services/inventory/internal/grpcserver"
	"github.com/phuchoang2603/refurbished-marketplace/services/inventory/internal/service"
	"github.com/phuchoang2603/refurbished-marketplace/shared/messaging"
	inventoryv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/inventory/v1"
	ordersv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/orders/v1"
	productsv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/products/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func scalar(t *testing.T, db *sql.DB, query string, args ...any) int {
	t.Helper()
	var n int
	if err := db.QueryRowContext(t.Context(), query, args...).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func assertStock(t *testing.T, svc *service.Service, id uuid.UUID, available, reserved int32) {
	t.Helper()
	got, err := svc.GetInventoryByProductID(t.Context(), id)
	if err != nil {
		t.Fatal(err)
	}
	if got.AvailableQty != available || got.ReservedQty != reserved {
		t.Fatalf("stock = %+v, want %d/%d", got, available, reserved)
	}
}

func TestCommandKafkaReplayAndPaymentRelease(t *testing.T) {
	db := newInventoryDB(t)
	svc := service.New(db)
	for _, outcome := range []string{"payment-failure", "session-expiry"} {
		t.Run(outcome, func(t *testing.T) {
			ctx := t.Context()
			id, order, merchant := uuid.New(), uuid.New(), uuid.New()
			if _, err := svc.EnsureStock(ctx, id, 5); err != nil {
				t.Fatal(err)
			}
			items := []service.ReservationItemInput{{ProductID: id, Quantity: 2}}
			if err := svc.ReserveStock(ctx, order, merchant, 1000, items); err != nil {
				t.Fatal(err)
			}
			payload := orderCreatedPayload(order, merchant, 1000, &ordersv1.OrderCreatedItem{ProductId: id.String(), Quantity: 2})
			for _, message := range []string{order.String() + "/1", order.String() + "/1", order.String() + "/2"} {
				if err := svc.HandleOrdersCreated(ctx, message, payload); err != nil {
					t.Fatal(err)
				}
			}
			assertStock(t, svc, id, 3, 2)
			if n := scalar(t, db, "SELECT count(*) FROM inventory_outbox WHERE aggregate_id=$1 AND event_type=$2", order, messaging.EventTypeInventoryReserved); n != 1 {
				t.Fatalf("reserved events = %d", n)
			}
			// Both gateway failure and hosted-session expiry produce payment.failed.
			for _, message := range []string{order.String() + "/failed/1", order.String() + "/failed/1", order.String() + "/failed/2"} {
				if err := svc.HandlePaymentOutcome(ctx, message, messaging.EventTypePaymentFailed, paymentOutcomePayload(order)); err != nil {
					t.Fatal(err)
				}
			}
			assertStock(t, svc, id, 5, 0)
			if n := scalar(t, db, "SELECT count(*) FROM inventory_reservations WHERE order_id=$1 AND status='RELEASED'", order); n != 1 {
				t.Fatalf("released reservations = %d", n)
			}
		})
	}
}

func TestReserveFullOrderFailureIsAtomic(t *testing.T) {
	db := newInventoryDB(t)
	svc := service.New(db)
	for _, missing := range []bool{false, true} {
		first, second, order := uuid.New(), uuid.New(), uuid.New()
		if _, err := svc.EnsureStock(t.Context(), first, 5); err != nil {
			t.Fatal(err)
		}
		if !missing {
			if _, err := svc.EnsureStock(t.Context(), second, 1); err != nil {
				t.Fatal(err)
			}
		}
		err := svc.ReserveStock(t.Context(), order, uuid.New(), 1000, []service.ReservationItemInput{{ProductID: first, Quantity: 2}, {ProductID: second, Quantity: 2}})
		want := service.ErrInsufficientStock
		if missing {
			want = service.ErrInventoryNotFound
		}
		if !errors.Is(err, want) {
			t.Fatalf("reserve error = %v, want %v", err, want)
		}
		assertStock(t, svc, first, 5, 0)
		if n := scalar(t, db, "SELECT count(*) FROM inventory_reservations WHERE order_id=$1", order); n != 0 {
			t.Fatalf("partial reservations = %d", n)
		}
		if n := scalar(t, db, "SELECT count(*) FROM inventory_outbox WHERE aggregate_id=$1 AND event_type=$2", order, messaging.EventTypeInventoryReservationFailed); n != 1 {
			t.Fatalf("failure events = %d", n)
		}
	}
}

func TestStockReadRPCs(t *testing.T) {
	svc := newInventoryService(t)
	server := grpcserver.New(svc)
	known, missing := uuid.New(), uuid.New()
	if _, err := svc.EnsureStock(t.Context(), known, 4); err != nil {
		t.Fatal(err)
	}
	row, err := server.GetStock(t.Context(), &inventoryv1.GetStockRequest{ProductId: known.String()})
	if err != nil || row.GetAvailableQty() != 4 || row.GetReservedQty() != 0 {
		t.Fatalf("stock = %v, %v", row, err)
	}
	if _, err := server.GetStock(t.Context(), &inventoryv1.GetStockRequest{ProductId: missing.String()}); status.Code(err) != codes.NotFound {
		t.Fatalf("missing error = %v", err)
	}
	batch, err := server.GetStocksByIDs(t.Context(), &inventoryv1.GetStocksByIDsRequest{ProductIds: []string{known.String(), missing.String(), known.String()}})
	if err != nil || len(batch.GetStocks()) != 1 || batch.Stocks[0].GetProductId() != known.String() {
		t.Fatalf("batch = %v, %v", batch, err)
	}
	ids := make([]string, 101)
	for i := range ids {
		ids[i] = uuid.NewString()
	}
	if _, err := server.GetStocksByIDs(t.Context(), &inventoryv1.GetStocksByIDsRequest{ProductIds: ids}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("batch limit error = %v", err)
	}
}

func creationEvent(qty *int32) *productsv1.ProductCreated {
	return &productsv1.ProductCreated{EventId: uuid.NewString(), ProductId: uuid.NewString(), SchemaVersion: 1, ProductVersion: 1, OccurredAt: timestamppb.Now(), Name: "Phone", Description: "Refurbished", MerchantId: uuid.NewString(), PriceCents: 1000, InitialQty: qty}
}

func consumeCreation(t *testing.T, svc *service.Service, event *productsv1.ProductCreated) error {
	t.Helper()
	payload, err := proto.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	return svc.KafkaReservationHandler()(t.Context(), messaging.KafkaMessage{Topic: messaging.EventTypeProductCreated, Value: payload})
}

func TestProductCreatedIdempotencyAndValidation(t *testing.T) {
	db := newInventoryDB(t)
	svc := service.New(db)
	for _, qty := range []int32{0, 5} {
		event := creationEvent(proto.Int32(qty))
		id := uuid.MustParse(event.ProductId)
		if err := consumeCreation(t, svc, event); err != nil {
			t.Fatal(err)
		}
		assertStock(t, svc, id, qty, 0)
		if qty > 0 {
			if err := svc.ReserveStock(t.Context(), uuid.New(), uuid.New(), 1000, []service.ReservationItemInput{{ProductID: id, Quantity: 2}}); err != nil {
				t.Fatal(err)
			}
		}
		// Retry after commit (lost acknowledgement) and replay under a new event id.
		if err := consumeCreation(t, svc, event); err != nil {
			t.Fatal(err)
		}
		event.EventId = uuid.NewString()
		if err := consumeCreation(t, svc, event); err != nil {
			t.Fatal(err)
		}
		if qty > 0 {
			assertStock(t, svc, id, 3, 2)
		} else {
			assertStock(t, svc, id, 0, 0)
		}
		conflict := proto.Clone(event).(*productsv1.ProductCreated)
		conflict.EventId = uuid.NewString()
		conflict.InitialQty = proto.Int32(qty + 1)
		if err := consumeCreation(t, svc, conflict); !errors.Is(err, service.ErrConflictingSeed) {
			t.Fatalf("conflict = %v", err)
		}
		if n := scalar(t, db, "SELECT count(*) FROM inventory_inbox WHERE message_id=$1", messaging.EventTypeProductCreated+"/"+conflict.EventId); n != 0 {
			t.Fatal("conflicting event acknowledged")
		}
	}
	for _, qty := range []*int32{nil, proto.Int32(-1)} {
		event := creationEvent(qty)
		if err := consumeCreation(t, svc, event); !errors.Is(err, service.ErrInvalidQuantity) {
			t.Fatalf("invalid seed = %v", err)
		}
		if n := scalar(t, db, "SELECT count(*) FROM inventory WHERE product_id=$1", event.ProductId); n != 0 {
			t.Fatal("invalid seed persisted")
		}
		if n := scalar(t, db, "SELECT count(*) FROM inventory_inbox WHERE message_id=$1", messaging.EventTypeProductCreated+"/"+event.EventId); n != 0 {
			t.Fatal("invalid event acknowledged")
		}
	}
}

func TestProductCreatedTransactionRollbackAndRetry(t *testing.T) {
	db := newInventoryDB(t)
	svc := service.New(db)
	event := creationEvent(proto.Int32(3))
	// Fail the last write, after inbox and intent insertion.
	if _, err := db.ExecContext(t.Context(), "ALTER TABLE inventory ADD CONSTRAINT reject_seed CHECK (available_qty < 0)"); err != nil {
		t.Fatal(err)
	}
	if err := consumeCreation(t, svc, event); err == nil {
		t.Fatal("expected seed failure")
	}
	for _, table := range []string{"inventory", "inventory_seed_intents", "inventory_inbox"} {
		if n := scalar(t, db, "SELECT count(*) FROM "+table); n != 0 {
			t.Fatalf("%s committed despite failure", table)
		}
	}
	if _, err := db.ExecContext(t.Context(), "ALTER TABLE inventory DROP CONSTRAINT reject_seed"); err != nil {
		t.Fatal(err)
	}
	if err := consumeCreation(t, svc, event); err != nil {
		t.Fatal(err)
	}
	assertStock(t, svc, uuid.MustParse(event.ProductId), 3, 0)
}
