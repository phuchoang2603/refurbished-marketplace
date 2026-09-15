package tests

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/meilisearch/meilisearch-go"
	"github.com/phuchoang2603/refurbished-marketplace/services/search/internal/service"
	"github.com/phuchoang2603/refurbished-marketplace/shared/messaging"
	productsv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/products/v1"
	searchv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/search/v1"
	testkafka "github.com/phuchoang2603/refurbished-marketplace/shared/testutil/kafka"
	testmeili "github.com/phuchoang2603/refurbished-marketplace/shared/testutil/meilisearch"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func newSearchService(t *testing.T) *service.Service {
	t.Helper()
	meili := testmeili.Setup(t)
	client := meilisearch.New(meili.URL, meilisearch.WithAPIKey(meili.MasterKey))
	svc := service.New(client, "listings")
	if err := svc.EnsureIndexSettings(t.Context()); err != nil {
		t.Fatalf("ensure index settings: %v", err)
	}
	return svc
}

func TestSearchProductsBrowseAndMerchantFilter(t *testing.T) {
	svc := newSearchService(t)
	ctx := t.Context()
	merchantA := uuid.NewString()
	merchantB := uuid.NewString()
	phoneID := uuid.NewString()
	laptopID := uuid.NewString()

	if err := svc.HandleProductCreated(ctx, mustMarshal(t, &productsv1.ProductCreated{
		EventId: uuid.NewString(), ProductId: phoneID, SchemaVersion: 1, ProductVersion: 1,
		OccurredAt: timestamppb.New(time.Now().Add(-time.Minute)), Name: "Refurbished Phone",
		Description: "Battery replaced", PriceCents: 25999, MerchantId: merchantA, InitialQty: proto.Int32(4),
	})); err != nil {
		t.Fatalf("index phone: %v", err)
	}
	if err := svc.HandleProductCreated(ctx, mustMarshal(t, &productsv1.ProductCreated{
		EventId: uuid.NewString(), ProductId: laptopID, SchemaVersion: 1, ProductVersion: 1,
		OccurredAt: timestamppb.Now(), Name: "Other Laptop", Description: "Seller B listing",
		PriceCents: 99999, MerchantId: merchantB, InitialQty: proto.Int32(1),
	})); err != nil {
		t.Fatalf("index laptop: %v", err)
	}

	browse, err := svc.SearchProducts(ctx, "", "", 20, 0)
	if err != nil {
		t.Fatalf("browse: %v", err)
	}
	if !containsListing(browse, phoneID) || !containsListing(browse, laptopID) {
		t.Fatalf("browse missing listings: %+v", browse)
	}
	if browse[0].GetId() != laptopID {
		t.Fatalf("browse should rank newest first, got %+v", browse)
	}

	seller, err := svc.SearchProducts(ctx, "", merchantA, 20, 0)
	if err != nil {
		t.Fatalf("merchant filter: %v", err)
	}
	if !containsListing(seller, phoneID) || containsListing(seller, laptopID) {
		t.Fatalf("merchant filter = %+v", seller)
	}
}

func TestSearchProductsTextMatchesNameAndDescription(t *testing.T) {
	svc := newSearchService(t)
	ctx := t.Context()
	phoneID := uuid.NewString()
	laptopID := uuid.NewString()

	if err := svc.HandleProductCreated(ctx, mustMarshal(t, &productsv1.ProductCreated{
		EventId: uuid.NewString(), ProductId: phoneID, SchemaVersion: 1, ProductVersion: 1,
		OccurredAt: timestamppb.New(time.Now().Add(-time.Hour)), Name: "Refurbished Phone",
		Description: "Battery replaced", PriceCents: 25999, MerchantId: uuid.NewString(), InitialQty: proto.Int32(4),
	})); err != nil {
		t.Fatalf("index phone: %v", err)
	}
	if err := svc.HandleProductCreated(ctx, mustMarshal(t, &productsv1.ProductCreated{
		EventId: uuid.NewString(), ProductId: laptopID, SchemaVersion: 1, ProductVersion: 1,
		OccurredAt: timestamppb.Now(), Name: "Other Laptop", Description: "Seller B listing",
		PriceCents: 99999, MerchantId: uuid.NewString(), InitialQty: proto.Int32(1),
	})); err != nil {
		t.Fatalf("index laptop: %v", err)
	}

	byName, err := svc.SearchProducts(ctx, "Phone", "", 20, 0)
	if err != nil {
		t.Fatalf("name query: %v", err)
	}
	if !containsListing(byName, phoneID) || containsListing(byName, laptopID) {
		t.Fatalf("name query = %+v", byName)
	}

	byDesc, err := svc.SearchProducts(ctx, "Battery", "", 20, 0)
	if err != nil {
		t.Fatalf("description query: %v", err)
	}
	if !containsListing(byDesc, phoneID) || containsListing(byDesc, laptopID) {
		t.Fatalf("description query = %+v", byDesc)
	}
}

func TestCreateIndexesViaKafka(t *testing.T) {
	svc := newSearchService(t)
	event := &productsv1.ProductCreated{
		EventId: uuid.NewString(), ProductId: uuid.NewString(), SchemaVersion: 1, ProductVersion: 1,
		OccurredAt: timestamppb.Now(), Name: "Kafka Phone", Description: "From products.created",
		PriceCents: 1000, MerchantId: uuid.NewString(), InitialQty: proto.Int32(2),
	}
	payload, err := proto.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}

	k := testkafka.SetupKafka(t)
	brokers, err := k.Brokers(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	topic := messaging.EventTypeProductCreated
	testkafka.ProduceKafkaRecord(t, t.Context(), brokers, topic, payload)
	cancel, runErr := testkafka.StartKafkaConsumer(t, t.Context(), brokers, "search-product-created-"+uuid.NewString(), []string{topic}, svc.KafkaProductCreatedHandler())
	defer cancel()
	testkafka.WaitForKafkaCondition(t, runErr, cancel, 30*time.Second, 200*time.Millisecond, "listing was not indexed", func() (bool, error) {
		hits, err := svc.SearchProducts(t.Context(), "", "", 20, 0)
		if err != nil {
			return false, err
		}
		return containsListing(hits, event.ProductId), nil
	})
}

func mustMarshal(t *testing.T, msg proto.Message) []byte {
	t.Helper()
	payload, err := proto.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}
	return payload
}

func containsListing(hits []*searchv1.ListingHit, id string) bool {
	for _, hit := range hits {
		if hit.GetId() == id {
			return true
		}
	}
	return false
}
