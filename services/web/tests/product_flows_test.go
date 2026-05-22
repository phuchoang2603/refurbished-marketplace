package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"refurbished-marketplace/services/web/internal/auth"
	productsv1 "refurbished-marketplace/shared/proto/products/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCreateProductRedirectsToProductDetail(t *testing.T) {
	productsSvc := &fakeProductsService{
		createFn: func(ctx context.Context, name, description string, priceCents int64, merchantID string, initialStock int32) (*productsv1.Product, error) {
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
				t.Fatalf("initialStock = %d, want 4", initialStock)
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
	if got := rec.Header().Get("Location"); got != "/products/prod-1" {
		t.Fatalf("location = %q, want /products/prod-1", got)
	}
}

func TestCreateProductReturnsUnavailableWhenProductsServiceFails(t *testing.T) {
	productsSvc := &fakeProductsService{
		createFn: func(ctx context.Context, name, description string, priceCents int64, merchantID string, initialStock int32) (*productsv1.Product, error) {
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
	stock := int32(4)
	productsSvc := &fakeProductsService{
		listFn: func(ctx context.Context, limit, offset int32) (*productsv1.ListProductsResponse, error) {
			return &productsv1.ListProductsResponse{Products: []*productsv1.Product{
				{Id: "prod-1", MerchantId: "11111111-1111-1111-1111-111111111111", Name: "Seller Phone", PriceCents: 25999, AvailableQty: &stock},
				{Id: "prod-2", MerchantId: "22222222-2222-2222-2222-222222222222", Name: "Other Laptop", PriceCents: 99999, AvailableQty: &stock},
			}}, nil
		},
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/seller/products", nil)
	req.AddCookie(&http.Cookie{Name: auth.AccessCookieName, Value: signedAccessToken(t, "11111111-1111-1111-1111-111111111111")})

	newTestRouter(t, routerDeps{products: productsSvc}).ServeHTTP(rec, req)

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

func TestProductsPageHidesCurrentUsersProducts(t *testing.T) {
	stock := int32(4)
	productsSvc := &fakeProductsService{
		listFn: func(ctx context.Context, limit, offset int32) (*productsv1.ListProductsResponse, error) {
			return &productsv1.ListProductsResponse{Products: []*productsv1.Product{
				{Id: "prod-1", MerchantId: "11111111-1111-1111-1111-111111111111", Name: "Own Phone", Description: "This should be hidden", PriceCents: 25999, AvailableQty: &stock},
				{Id: "prod-2", MerchantId: "22222222-2222-2222-2222-222222222222", Name: "Other Laptop", Description: "Visible", PriceCents: 99999, AvailableQty: &stock},
			}}, nil
		},
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	req.AddCookie(&http.Cookie{Name: auth.AccessCookieName, Value: signedAccessToken(t, "11111111-1111-1111-1111-111111111111")})

	newTestRouter(t, routerDeps{products: productsSvc}).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if strings.Contains(body, "Own Phone") {
		t.Fatalf("body should not include current user's product in %q", body)
	}
	if !strings.Contains(body, "Other Laptop") {
		t.Fatalf("body missing visible product in %q", body)
	}
}

func TestProductDetailHidesCartFormForOwner(t *testing.T) {
	stock := int32(4)
	productsSvc := &fakeProductsService{
		getByIDFn: func(ctx context.Context, id string) (*productsv1.Product, error) {
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
