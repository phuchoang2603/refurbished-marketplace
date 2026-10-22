package tests

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/phuchoang2603/refurbished-marketplace/services/products/internal/grpcserver"
	"github.com/phuchoang2603/refurbished-marketplace/services/products/internal/service"
	productsv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/products/v1"
	"go.mongodb.org/mongo-driver/v2/bson"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func TestInvalidGradingRejectedBeforePersistence(t *testing.T) {
	server := grpcserver.New(service.New(nil))
	for _, tc := range []struct{ category, condition string }{
		{"", "good"},
		{"computer", "good"},
		{"Phone", "good"},
		{"phone ", "good"},
		{"phone", ""},
		{"phone", "new"},
		{"phone", "GOOD"},
		{"phone", "good "},
	} {
		t.Run(tc.category+"/"+tc.condition, func(t *testing.T) {
			_, err := server.CreateProduct(t.Context(), &productsv1.CreateProductRequest{Name: "Phone", PriceCents: 1000, MerchantId: uuid.NewString(), InitialStock: proto.Int32(1), Category: tc.category, Condition: tc.condition})
			if status.Code(err) != codes.InvalidArgument {
				t.Fatalf("error=%v", err)
			}
		})
	}
}

func TestGradingPersistsInReadsAndEvents(t *testing.T) {
	store := newProductsStore(t)
	server := grpcserver.New(service.New(store))
	for _, category := range []string{"phone", "laptop", "tablet", "accessory", "headphones", "smartwatch", "game_console", "camera"} {
		for _, condition := range []string{"excellent", "good", "fair"} {
			req := &productsv1.CreateProductRequest{Name: "Listing", PriceCents: 1000, MerchantId: uuid.NewString(), InitialStock: proto.Int32(1), Category: category, Condition: condition}
			created, err := server.CreateProduct(t.Context(), req)
			if err != nil {
				t.Fatal(err)
			}
			read, err := server.GetProductByID(t.Context(), &productsv1.GetProductByIDRequest{Id: created.Id})
			if err != nil {
				t.Fatal(err)
			}
			batch, err := server.GetProductsByIDs(t.Context(), &productsv1.GetProductsByIDsRequest{Ids: []string{created.Id}})
			if err != nil || len(batch.GetProducts()) != 1 {
				t.Fatalf("batch=%v error=%v", batch, err)
			}
			for _, p := range []*productsv1.Product{created, read, batch.Products[0]} {
				if p.GetCategory() != category || p.GetCondition() != condition {
					t.Fatalf("grading lost: %v", p)
				}
			}
			var row outboxRow
			if err := store.Database().Collection("catalog_outbox").FindOne(t.Context(), bson.M{"aggregate_id": created.Id}).Decode(&row); err != nil {
				t.Fatal(err)
			}
			var event productsv1.ProductCreated
			if err := proto.Unmarshal(row.Payload, &event); err != nil {
				t.Fatal(err)
			}
			if event.GetCategory() != category || event.GetCondition() != condition {
				t.Fatalf("event grading lost: %v", &event)
			}
		}
	}
}

func TestGradingBackfillDryRunAndIdempotency(t *testing.T) {
	store := newProductsStore(t)
	collection := store.Database().Collection("listings")
	legacyID, partialID, gradedID := uuid.NewString(), uuid.NewString(), uuid.NewString()
	legacy := func(id string) bson.M {
		return bson.M{"_id": id, "name": "Legacy", "price_cents": 1000, "merchant_id": uuid.NewString(), "created_at": time.Now(), "updated_at": time.Now()}
	}
	partial := legacy(partialID)
	partial["category"] = "camera"
	graded := legacy(gradedID)
	graded["category"] = "phone"
	graded["condition"] = "excellent"
	if _, err := collection.InsertMany(t.Context(), []any{legacy(legacyID), partial, graded}); err != nil {
		t.Fatal(err)
	}
	dry, err := store.BackfillGrading(t.Context(), false)
	if err != nil || dry.Category != 1 || dry.Condition != 2 {
		t.Fatalf("dry=%v err=%v", dry, err)
	}
	var unchanged bson.M
	if err := collection.FindOne(t.Context(), bson.M{"_id": legacyID}).Decode(&unchanged); err != nil {
		t.Fatal(err)
	}
	if _, ok := unchanged["category"]; ok {
		t.Fatal("dry run wrote category")
	}
	if _, ok := unchanged["condition"]; ok {
		t.Fatal("dry run wrote condition")
	}
	applied, err := store.BackfillGrading(t.Context(), true)
	if err != nil || applied != dry {
		t.Fatalf("apply=%v err=%v", applied, err)
	}
	for id, want := range map[string][2]string{legacyID: {"accessory", "good"}, partialID: {"camera", "good"}, gradedID: {"phone", "excellent"}} {
		p, err := store.GetListingByID(t.Context(), uuid.MustParse(id))
		if err != nil || p.Category != want[0] || p.Condition != want[1] {
			t.Fatalf("read=%v err=%v", p, err)
		}
	}
	repeated, err := store.BackfillGrading(t.Context(), true)
	if err != nil || repeated.Category != 0 || repeated.Condition != 0 {
		t.Fatalf("repeat=%v err=%v", repeated, err)
	}
}
