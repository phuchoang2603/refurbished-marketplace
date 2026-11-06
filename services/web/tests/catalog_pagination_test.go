package tests

import (
	"context"
	"fmt"
	"html"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/phuchoang2603/refurbished-marketplace/services/web/tests/fakes"
	searchv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/search/v1"
)

func TestCatalogPagination(t *testing.T) {
	for _, tc := range []struct {
		name, page     string
		offset         int32
		hits           int
		previous, next string
	}{
		{name: "first", page: "1", hits: 21, next: "2"},
		{name: "middle", page: "2", offset: 20, hits: 21, previous: "1", next: "3"},
		{name: "full last", page: "3", offset: 40, hits: 20, previous: "2"},
		{name: "partial last", page: "4", offset: 60, hits: 5, previous: "3"},
		{name: "empty", page: "5", offset: 80, previous: "4"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			search := &fakes.SearchService{SearchFn: func(ctx context.Context, query, merchant string, limit, offset int32) (*searchv1.SearchProductsResponse, error) {
				if query != "phone&camera" || merchant != "" || limit != 21 || offset != tc.offset {
					t.Fatalf("query=%q merchant=%q limit=%d offset=%d", query, merchant, limit, offset)
				}
				hits := make([]*searchv1.ListingHit, tc.hits)
				for i := range hits {
					hits[i] = &searchv1.ListingHit{Id: fmt.Sprintf("listing-%d", i), Name: "Phone"}
				}
				return &searchv1.SearchProductsResponse{Listings: hits}, nil
			}}
			rec := httptest.NewRecorder()
			newTestRouter(t, routerDeps{search: search}).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/products?q=phone%26camera&page="+tc.page, nil))
			if rec.Code != http.StatusOK {
				t.Fatalf("status: %d", rec.Code)
			}
			body := html.UnescapeString(rec.Body.String())
			if got, want := strings.Count(body, `id="product-`), min(tc.hits, 20); got != want {
				t.Fatalf("cards=%d, want %d", got, want)
			}
			for _, link := range []struct{ relation, page string }{{"prev", tc.previous}, {"next", tc.next}} {
				if strings.Contains(body, `rel="`+link.relation+`"`) != (link.page != "") {
					t.Fatalf("unexpected %s link", link.relation)
				}
				if link.page != "" && !strings.Contains(body, `href="/products?page=`+link.page+`&q=phone%26camera"`) {
					t.Fatalf("%s link lost query or page", link.relation)
				}
			}
		})
	}
}

func TestCatalogRejectsInvalidPages(t *testing.T) {
	for _, page := range []string{"0", "-1", "abc", "2147483647", "999999999999999999999"} {
		t.Run(page, func(t *testing.T) {
			search := &fakes.SearchService{SearchFn: func(context.Context, string, string, int32, int32) (*searchv1.SearchProductsResponse, error) {
				t.Fatal("invalid page reached search")
				return nil, nil
			}}
			rec := httptest.NewRecorder()
			newTestRouter(t, routerDeps{search: search}).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/products?page="+page, nil))
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status=%d", rec.Code)
			}
		})
	}
}
