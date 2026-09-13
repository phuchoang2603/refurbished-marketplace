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

func TestCreateProductDurableEvent(t *testing.T) {
	db := newProductsDB(t)
	svc := service.New(db)
	server := grpcserver.New(svc)
	req := &productsv1.CreateProductRequest{Name: "Phone", Description: "Refurbished", MerchantId: uuid.NewString(), PriceCents: 1000, InitialStock: proto.Int32(0)}
	created, err := server.CreateProduct(t.Context(), req)
	if err != nil {
		t.Fatal(err)
	}
	var eventID, productID, eventType string
	var payload []byte
	// No broker or publisher runs here: creation must still commit durable data.
	if err := db.QueryRowContext(t.Context(), "SELECT id, aggregate_id, event_type, payload FROM products_outbox WHERE aggregate_id=$1", created.GetId()).Scan(&eventID, &productID, &eventType, &payload); err != nil {
		t.Fatal(err)
	}
	var event productsv1.ProductCreated
	if err := proto.Unmarshal(payload, &event); err != nil {
		t.Fatal(err)
	}
	if eventID != event.GetEventId() || productID != event.GetProductId() || productID != created.GetId() || eventType != messaging.EventTypeProductCreated {
		t.Fatalf("outbox identity mismatch: %v", &event)
	}
	if event.InitialQty == nil || event.GetInitialQty() != 0 || event.GetName() != req.Name || event.GetDescription() != req.Description || event.GetPriceCents() != req.PriceCents || event.GetMerchantId() != req.MerchantId || event.GetSchemaVersion() != 1 || event.GetProductVersion() != 1 || !event.GetOccurredAt().AsTime().Equal(created.GetCreatedAt().AsTime()) {
		t.Fatalf("unexpected event: %v", &event)
	}
	// Reading catalog while publication is delayed must not consume/change the outbox.
	if _, err := svc.GetProductByID(t.Context(), uuid.MustParse(created.Id)); err != nil {
		t.Fatal(err)
	}
	var retainedID string
	var retainedPayload []byte
	if err := db.QueryRowContext(t.Context(), "SELECT id, payload FROM products_outbox WHERE aggregate_id=$1", created.Id).Scan(&retainedID, &retainedPayload); err != nil {
		t.Fatal(err)
	}
	var retained productsv1.ProductCreated
	if err := proto.Unmarshal(retainedPayload, &retained); err != nil {
		t.Fatal(err)
	}
	if retainedID != eventID || !proto.Equal(&event, &retained) {
		t.Fatal("delayed publication changed event")
	}
}

func TestCreateProductRollbackAndQuantityPresence(t *testing.T) {
	db := newProductsDB(t)
	server := grpcserver.New(service.New(db))
	req := &productsv1.CreateProductRequest{Name: "Phone", MerchantId: uuid.NewString(), PriceCents: 1000}
	for _, qty := range []*int32{nil, proto.Int32(-1)} {
		req.InitialStock = qty
		if _, err := server.CreateProduct(t.Context(), req); status.Code(err) != codes.InvalidArgument {
			t.Fatalf("quantity error = %v", err)
		}
	}
	req.InitialStock = proto.Int32(3)
	if _, err := db.ExecContext(t.Context(), "ALTER TABLE products_outbox ADD CONSTRAINT reject_event CHECK (event_type = 'disabled')"); err != nil {
		t.Fatal(err)
	}
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
	if _, err := db.ExecContext(t.Context(), "ALTER TABLE products_outbox DROP CONSTRAINT reject_event"); err != nil {
		t.Fatal(err)
	}
	if _, err := server.CreateProduct(t.Context(), req); err != nil {
		t.Fatal(err)
	}
}
