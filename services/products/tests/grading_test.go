package tests

import (
	"testing"

	"github.com/google/uuid"
	"github.com/phuchoang2603/refurbished-marketplace/services/products/internal/grpcserver"
	"github.com/phuchoang2603/refurbished-marketplace/services/products/internal/service"
	productsv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/products/v1"
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
			if err := store.Database().QueryRowContext(t.Context(), "SELECT id, aggregate_id, event_type, payload FROM products_outbox WHERE aggregate_id = $1", created.Id).Scan(&row.ID, &row.AggregateID, &row.EventType, &row.Payload); err != nil {
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
