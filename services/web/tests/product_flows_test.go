package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/phuchoang2603/refurbished-marketplace/services/web/internal/auth"
	"github.com/phuchoang2603/refurbished-marketplace/services/web/tests/fakes"
	productsv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/products/v1"
	searchv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/search/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCreateProductRedirectsToProductDetail(t *testing.T) {
	productsSvc := &fakes.ProductsService{
		CreateFn: func(ctx context.Context, name, description string, priceCents int64, merchantID string, initialStock int32) (*productsv1.Product, error) {
			if name != "Refurbished Phone" {
				t.Fatalf("name = %q, want Refurbished Phone", name)
			}
			if description != "Battery replaced and tested." {
				t.Fatalf("description = %q, want expected description", description)
			}
			if priceCents != 25999 {
				t.Fatalf("priceCents = %d, want 25999", priceCents)
			}
			if merchantID != "11111111-1111-1111-1111-111111111111" {
				t.Fatalf("merchantID = %q, want UUID subject", merchantID)
			}
			if initialStock != 4 {
				t.Fatalf("initial stock = %d", initialStock)
			}
			return &productsv1.Product{Id: "prod-1"}, nil
		},
	}
	form := url.Values{
		"name":          {"Refurbished Phone"},
		"description":   {"Battery replaced and tested."},
		"price":         {"259.99"},
		"initial_stock": {"4"},
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/seller/products", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: auth.AccessCookieName, Value: signedAccessToken(t, "11111111-1111-1111-1111-111111111111")})

	newTestRouter(t, routerDeps{products: productsSvc}).ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusSeeOther)
	}
	if got := rec.Header().Get("Location"); got != "/products/prod-1?created=1" {
		t.Fatalf("location = %q, want /products/prod-1", got)
	}
}

func TestCreateProductReturnsUnavailableWhenProductsServiceFails(t *testing.T) {
	productsSvc := &fakes.ProductsService{
		CreateFn: func(ctx context.Context, name, description string, priceCents int64, merchantID string, initialStock int32) (*productsv1.Product, error) {
			return nil, status.Error(codes.Unavailable, "products service unavailable")
		},
	}
	form := url.Values{
		"name":          {"Refurbished Phone"},
		"description":   {"Battery replaced and tested."},
		"price":         {"259.99"},
		"initial_stock": {"4"},
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/seller/products", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: auth.AccessCookieName, Value: signedAccessToken(t, "11111111-1111-1111-1111-111111111111")})

	newTestRouter(t, routerDeps{products: productsSvc}).ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
}

func TestSellerProductsPageListsOnlyCurrentSellerProducts(t *testing.T) {
	searchSvc := &fakes.SearchService{
		SearchFn: func(ctx context.Context, query, merchantID string, limit, offset int32) (*searchv1.SearchProductsResponse, error) {
			if merchantID != "11111111-1111-1111-1111-111111111111" {
				t.Fatalf("merchantID = %q, want authenticated seller", merchantID)
			}
			return &searchv1.SearchProductsResponse{Listings: []*searchv1.ListingHit{
				{Id: "prod-1", MerchantId: merchantID, Name: "Seller Phone", PriceCents: 25999},
			}}, nil
		},
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/seller/products", nil)
	req.AddCookie(&http.Cookie{Name: auth.AccessCookieName, Value: signedAccessToken(t, "11111111-1111-1111-1111-111111111111")})

	newTestRouter(t, routerDeps{search: searchSvc}).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Seller Phone") {
		t.Fatalf("body missing seller product in %q", body)
	}
	if strings.Contains(body, "Other Laptop") {
		t.Fatalf("body should not include another seller product in %q", body)
	}
}

func TestProductsPageRendersSearchHitsWithoutStock(t *testing.T) {
	searchSvc := &fakes.SearchService{
		SearchFn: func(ctx context.Context, query, merchantID string, limit, offset int32) (*searchv1.SearchProductsResponse, error) {
			if query != "" || merchantID != "" {
				t.Fatalf("browse query=%q merchant=%q", query, merchantID)
			}
			return &searchv1.SearchProductsResponse{Listings: []*searchv1.ListingHit{
				{Id: "prod-1", MerchantId: "22222222-2222-2222-2222-222222222222", Name: "Browse Phone", Description: "Ready to ship", PriceCents: 25999},
			}}, nil
		},
	}
	rec := httptest.NewRecorder()
	newTestRouter(t, routerDeps{search: searchSvc}).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/products", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Browse Phone") {
		t.Fatalf("body missing listing in %q", body)
	}
	if strings.Contains(body, "In stock") || strings.Contains(body, "Out of stock") {
		t.Fatalf("browse cards should omit stock in %q", body)
	}
}

func TestProductsPageForwardsCatalogQuery(t *testing.T) {
	searchSvc := &fakes.SearchService{
		SearchFn: func(ctx context.Context, query, merchantID string, limit, offset int32) (*searchv1.SearchProductsResponse, error) {
			if query != "pixel" || merchantID != "" {
				t.Fatalf("query=%q merchant=%q", query, merchantID)
			}
			return &searchv1.SearchProductsResponse{Listings: []*searchv1.ListingHit{
				{Id: "prod-1", MerchantId: "22222222-2222-2222-2222-222222222222", Name: "Pixel 8", PriceCents: 24900},
			}}, nil
		},
	}
	rec := httptest.NewRecorder()
	newTestRouter(t, routerDeps{search: searchSvc}).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/products?q=pixel", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Pixel 8") {
		t.Fatalf("body missing listing in %q", body)
	}
	if !strings.Contains(body, `value="pixel"`) {
		t.Fatalf("search input should keep query in %q", body)
	}
}

func TestCatalogSuggestReturnsNamesWithoutGrid(t *testing.T) {
	searchSvc := &fakes.SearchService{
		SearchFn: func(ctx context.Context, query, merchantID string, limit, offset int32) (*searchv1.SearchProductsResponse, error) {
			if query != "phone" || merchantID != "" || limit != 8 {
				t.Fatalf("query=%q merchant=%q limit=%d", query, merchantID, limit)
			}
			return &searchv1.SearchProductsResponse{Listings: []*searchv1.ListingHit{
				{Id: "prod-1", Name: "Refurbished Phone", PriceCents: 25999},
			}}, nil
		},
	}
	rec := httptest.NewRecorder()
	newTestRouter(t, routerDeps{search: searchSvc}).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/products/suggest?q=phone", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Refurbished Phone") || !strings.Contains(body, "/products/prod-1") {
		t.Fatalf("body missing suggestion in %q", body)
	}
	if strings.Contains(body, "Browse quality refurbished") {
		t.Fatalf("suggest must not render the catalog grid in %q", body)
	}
}

func TestCatalogSuggestSkipsShortQuery(t *testing.T) {
	searchSvc := &fakes.SearchService{
		SearchFn: func(ctx context.Context, query, merchantID string, limit, offset int32) (*searchv1.SearchProductsResponse, error) {
			t.Fatalf("search called for short query=%q", query)
			return nil, nil
		},
	}
	rec := httptest.NewRecorder()
	newTestRouter(t, routerDeps{search: searchSvc}).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/products/suggest?q=p", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), `id="catalog-suggest"`) {
		t.Fatalf("missing suggest root in %q", rec.Body.String())
	}
}

func TestProductDetailHidesCartFormForOwner(t *testing.T) {
	stock := int32(4)
	productsSvc := &fakes.ProductsService{
		GetByIDFn: func(ctx context.Context, id string) (*productsv1.Product, error) {
			return &productsv1.Product{Id: id, MerchantId: "11111111-1111-1111-1111-111111111111", Name: "Seller Phone", Description: "Owned item", PriceCents: 25999, AvailableQty: &stock}, nil
		},
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/products/prod-1", nil)
	req.AddCookie(&http.Cookie{Name: auth.AccessCookieName, Value: signedAccessToken(t, "11111111-1111-1111-1111-111111111111")})

	newTestRouter(t, routerDeps{products: productsSvc}).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if strings.Contains(body, "Add to cart") {
		t.Fatalf("body should not include add-to-cart form for owner in %q", body)
	}
	if !strings.Contains(body, "This is your product.") {
		t.Fatalf("body missing owner notice in %q", body)
	}
}
