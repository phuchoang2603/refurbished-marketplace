package service

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/meilisearch/meilisearch-go"
)

func BenchmarkSearchDecode(b *testing.B) {
	docs := make([]listingDocument, 100)
	for i := range docs {
		docs[i] = listingDocument{ID: fmt.Sprintf("product-%d", i), Name: fmt.Sprintf("Benchmark Laptop %d", i), Description: "Synthetic refurbished marketplace benchmark product", PriceCents: 19999, MerchantID: "550e8400-e29b-41d4-a716-446655440000", CreatedAt: 1793939373}
	}
	payload, err := json.Marshal(struct {
		Hits []listingDocument `json:"hits"`
	}{docs})
	if err != nil {
		b.Fatal(err)
	}
	b.Run("SDK", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			var response meilisearch.SearchResponse
			if err := json.Unmarshal(payload, &response); err != nil {
				b.Fatal(err)
			}
			var hits []listingDocument
			if err := response.Hits.DecodeInto(&hits); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("Typed", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			var response struct {
				Hits []listingDocument `json:"hits"`
			}
			if err := json.Unmarshal(payload, &response); err != nil {
				b.Fatal(err)
			}
		}
	})
}
