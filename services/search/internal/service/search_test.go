package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/meilisearch/meilisearch-go"
)

func TestSearchProductsResponse(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		wantError  bool
		wantCount  int
	}{
		{name: "typed fields", body: `{"hits":[{"id":"product-1","name":"Laptop 日本語","description":"Battery replaced","price_cents":9007199254740993,"merchant_id":"550e8400-e29b-41d4-a716-446655440000","created_at":1793939373,"extra":"ignored"}],"processingTimeMs":2}`, wantCount: 1},
		{name: "empty", body: `{"hits":[]}`},
		{name: "invalid price", body: `{"hits":[{"price_cents":"invalid"}]}`, wantError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var req meilisearch.SearchRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Error(err)
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				if r.Method != http.MethodPost || r.URL.Path != "/indexes/listings/search" || req.Query != "Laptop" || req.Limit != 100 || req.Offset != 17 {
					t.Errorf("request: %s %s %+v", r.Method, r.URL.Path, req)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(tc.body))
			}))
			defer server.Close()
			svc := New(meilisearch.New(server.URL), "listings")
			hits, err := svc.SearchProducts(t.Context(), "Laptop", "", 100, 17)
			if (err != nil) != tc.wantError {
				t.Fatalf("error: %v", err)
			}
			if len(hits) != tc.wantCount {
				t.Fatalf("hits: %+v", hits)
			}
			if tc.wantCount == 1 {
				h := hits[0]
				if h.Id != "product-1" || h.Name != "Laptop 日本語" || h.Description != "Battery replaced" || h.PriceCents != 9007199254740993 || h.MerchantId != "550e8400-e29b-41d4-a716-446655440000" || !h.CreatedAt.AsTime().Equal(time.Unix(1793939373, 0)) {
					t.Fatalf("decoded fields: %+v", h)
				}
			}
		})
	}
}
