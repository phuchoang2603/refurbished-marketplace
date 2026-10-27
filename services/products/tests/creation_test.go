package tests

import (
	"testing"

	"github.com/google/uuid"
	"github.com/phuchoang2603/refurbished-marketplace/services/products/internal/grpcserver"
	"github.com/phuchoang2603/refurbished-marketplace/services/products/internal/service"
	"github.com/phuchoang2603/refurbished-marketplace/shared/messaging"
	productsv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/products/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type outboxRow struct {
	ID          string
	AggregateID string
	EventType   string
	Payload     []byte
}

func TestCreateProductDurableEvent(t *testing.T) {
	store := newProductsStore(t)
	svc := service.New(store)
	server := grpcserver.New(svc)
	req := &productsv1.CreateProductRequest{Category: "phone", Condition: "good", Name: "Phone", Description: "Refurbished", MerchantId: uuid.NewString(), PriceCents: 1000, InitialStock: proto.Int32(0)}
	created, err := server.CreateProduct(t.Context(), req)
	if err != nil {
		t.Fatal(err)
	}
	var row outboxRow
	if err := store.Database().QueryRowContext(t.Context(), "SELECT id, aggregate_id, event_type, payload FROM products_outbox WHERE aggregate_id = $1", created.GetId()).Scan(&row.ID, &row.AggregateID, &row.EventType, &row.Payload); err != nil {
		t.Fatal(err)
	}
	var event productsv1.ProductCreated
	if err := proto.Unmarshal(row.Payload, &event); err != nil {
		t.Fatal(err)
	}
	if row.ID != event.GetEventId() || row.AggregateID != event.GetProductId() || row.AggregateID != created.GetId() || row.EventType != messaging.EventTypeProductCreated {
		t.Fatalf("outbox identity mismatch: %v", &event)
	}
	if event.InitialQty == nil || event.GetInitialQty() != 0 || event.GetName() != req.Name || event.GetDescription() != req.Description || event.GetPriceCents() != req.PriceCents || event.GetMerchantId() != req.MerchantId || event.GetSchemaVersion() != 1 || event.GetProductVersion() != 1 || !event.GetOccurredAt().AsTime().Equal(created.GetCreatedAt().AsTime()) {
		t.Fatalf("unexpected event: %v", &event)
	}
	if _, err := svc.GetProductByID(t.Context(), uuid.MustParse(created.Id)); err != nil {
		t.Fatal(err)
	}
	var retained outboxRow
	if err := store.Database().QueryRowContext(t.Context(), "SELECT id, aggregate_id, event_type, payload FROM products_outbox WHERE aggregate_id = $1", created.Id).Scan(&retained.ID, &retained.AggregateID, &retained.EventType, &retained.Payload); err != nil {
		t.Fatal(err)
	}
	var retainedEvent productsv1.ProductCreated
	if err := proto.Unmarshal(retained.Payload, &retainedEvent); err != nil {
		t.Fatal(err)
	}
	if retained.ID != row.ID || !proto.Equal(&event, &retainedEvent) {
		t.Fatal("delayed publication changed event")
	}
}

func TestCreateProductRollbackAndQuantityPresence(t *testing.T) {
	store := newProductsStore(t)
	server := grpcserver.New(service.New(store))
	req := &productsv1.CreateProductRequest{Category: "phone", Condition: "good", Name: "Phone", MerchantId: uuid.NewString(), PriceCents: 1000}
	for _, qty := range []*int32{nil, proto.Int32(-1)} {
		req.InitialStock = qty
		if _, err := server.CreateProduct(t.Context(), req); status.Code(err) != codes.InvalidArgument {
			t.Fatalf("quantity error = %v", err)
		}
	}
	db := store.Database()
	if _, err := db.ExecContext(t.Context(), "ALTER TABLE products_outbox ADD CONSTRAINT reject_events CHECK (event_type = 'disabled')"); err != nil {
		t.Fatal(err)
	}

	req.InitialStock = proto.Int32(3)
	if _, err := server.CreateProduct(t.Context(), req); err == nil {
		t.Fatal("expected outbox failure")
	}
	for _, table := range []string{"products", "products_outbox"} {
		var n int
		if err := db.QueryRowContext(t.Context(), "SELECT count(*) FROM "+table).Scan(&n); err != nil {
			t.Fatal(err)
		}
		if n != 0 {
			t.Fatalf("%s committed despite outbox failure", table)
		}
	}
	if _, err := db.ExecContext(t.Context(), "ALTER TABLE products_outbox DROP CONSTRAINT reject_events"); err != nil {
		t.Fatal(err)
	}

	if _, err := server.CreateProduct(t.Context(), req); err != nil {
		t.Fatal(err)
	}
}
