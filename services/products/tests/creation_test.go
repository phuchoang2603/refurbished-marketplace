package tests

import (
	"testing"

	"github.com/google/uuid"
	"github.com/phuchoang2603/refurbished-marketplace/services/products/internal/grpcserver"
	"github.com/phuchoang2603/refurbished-marketplace/services/products/internal/service"
	"github.com/phuchoang2603/refurbished-marketplace/shared/messaging"
	productsv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/products/v1"
	"go.mongodb.org/mongo-driver/v2/bson"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type outboxRow struct {
	ID          string `bson:"id"`
	AggregateID string `bson:"aggregate_id"`
	EventType   string `bson:"event_type"`
	Payload     []byte `bson:"payload"`
}

func TestCreateProductDurableEvent(t *testing.T) {
	store := newProductsStore(t)
	svc := service.New(store)
	server := grpcserver.New(svc)
	req := &productsv1.CreateProductRequest{Name: "Phone", Description: "Refurbished", MerchantId: uuid.NewString(), PriceCents: 1000, InitialStock: proto.Int32(0)}
	created, err := server.CreateProduct(t.Context(), req)
	if err != nil {
		t.Fatal(err)
	}
	var row outboxRow
	if err := store.Database().Collection("catalog_outbox").FindOne(t.Context(), bson.M{"aggregate_id": created.GetId()}).Decode(&row); err != nil {
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
	if err := store.Database().Collection("catalog_outbox").FindOne(t.Context(), bson.M{"aggregate_id": created.Id}).Decode(&retained); err != nil {
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
	req := &productsv1.CreateProductRequest{Name: "Phone", MerchantId: uuid.NewString(), PriceCents: 1000}
	for _, qty := range []*int32{nil, proto.Int32(-1)} {
		req.InitialStock = qty
		if _, err := server.CreateProduct(t.Context(), req); status.Code(err) != codes.InvalidArgument {
			t.Fatalf("quantity error = %v", err)
		}
	}
	db := store.Database()
	if err := db.CreateCollection(t.Context(), "catalog_outbox"); err != nil {
		t.Fatal(err)
	}
	if err := db.RunCommand(t.Context(), bson.D{
		{Key: "collMod", Value: "catalog_outbox"},
		{Key: "validator", Value: bson.M{"event_type": "disabled"}},
		{Key: "validationLevel", Value: "strict"},
		{Key: "validationAction", Value: "error"},
	}).Err(); err != nil {
		t.Fatal(err)
	}
	req.InitialStock = proto.Int32(3)
	if _, err := server.CreateProduct(t.Context(), req); err == nil {
		t.Fatal("expected outbox failure")
	}
	for _, coll := range []string{"listings", "catalog_outbox"} {
		n, err := db.Collection(coll).CountDocuments(t.Context(), bson.M{})
		if err != nil {
			t.Fatal(err)
		}
		if n != 0 {
			t.Fatalf("%s committed despite outbox failure", coll)
		}
	}
	if err := db.RunCommand(t.Context(), bson.D{
		{Key: "collMod", Value: "catalog_outbox"},
		{Key: "validator", Value: bson.M{}},
		{Key: "validationLevel", Value: "off"},
	}).Err(); err != nil {
		t.Fatal(err)
	}
	if _, err := server.CreateProduct(t.Context(), req); err != nil {
		t.Fatal(err)
	}
}
