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
	inventoryv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/inventory/v1"
	productsv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/products/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCreateProductQuantityAndInventoryIndependence(t *testing.T) {
	for _, qty := range []string{"", "-1", "0", "4"} {
		t.Run("quantity="+qty, func(t *testing.T) {
			called := false
			products := &fakes.ProductsService{CreateFn: func(_ context.Context, _, _ string, _ int64, _ string, initial int32) (*productsv1.Product, error) {
				called = true
				if qty == "0" && initial != 0 {
					t.Fatalf("initial = %d", initial)
				}
				return &productsv1.Product{Id: "new-product"}, nil
			}}
			inventory := &fakes.InventoryService{
				GetFn: func(context.Context, string) (*inventoryv1.Stock, error) {
					t.Fatal("creation read inventory")
					return nil, nil
				},
				ReserveFn: func(context.Context, string, string, int64, []*inventoryv1.ReserveStockItem) error {
					t.Fatal("creation reserved inventory")
					return nil
				},
			}
			form := url.Values{"name": {"Phone"}, "description": {"Refurbished"}, "price": {"10.00"}}
			if qty != "" {
				form.Set("initial_stock", qty)
			}
			req := httptest.NewRequest(http.MethodPost, "/seller/products", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.AddCookie(&http.Cookie{Name: auth.AccessCookieName, Value: signedAccessToken(t, "11111111-1111-1111-1111-111111111111")})
			rec := httptest.NewRecorder()
			newTestRouter(t, routerDeps{products: products, inventory: inventory}).ServeHTTP(rec, req)
			valid := qty == "0" || qty == "4"
			want := http.StatusBadRequest
			if valid {
				want = http.StatusSeeOther
			}
			if rec.Code != want || called != valid {
				t.Fatalf("response=%d called=%t, want %d/%t", rec.Code, called, want, valid)
			}
		})
	}
}

func TestProductAvailabilityStates(t *testing.T) {
	for _, tc := range []struct {
		name   string
		stock  *inventoryv1.Stock
		err    error
		text   string
		canBuy bool
	}{
		{name: "pending", err: status.Error(codes.NotFound, "missing"), text: "Availability processing"},
		{name: "unavailable", err: status.Error(codes.Unavailable, "offline"), text: "Availability unavailable"},
		{name: "zero", stock: &inventoryv1.Stock{}, text: "Out of stock"},
		{name: "ready", stock: &inventoryv1.Stock{AvailableQty: 5}, text: "In stock 5", canBuy: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			products := &fakes.ProductsService{GetByIDFn: func(context.Context, string) (*productsv1.Product, error) {
				return &productsv1.Product{Id: "product", Name: "Phone", MerchantId: "merchant", PriceCents: 1000}, nil
			}}
			reads := 0
			inventory := &fakes.InventoryService{GetFn: func(context.Context, string) (*inventoryv1.Stock, error) { reads++; return tc.stock, tc.err }}
			rec := httptest.NewRecorder()
			newTestRouter(t, routerDeps{products: products, inventory: inventory}).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/products/product", nil))
			body := rec.Body.String()
			if rec.Code != http.StatusOK || reads != 1 || !strings.Contains(body, tc.text) {
				t.Fatalf("response=%d reads=%d body=%s", rec.Code, reads, body)
			}
			if strings.Contains(body, `action="/cart/items"`) != tc.canBuy {
				t.Fatalf("purchase controls mismatch: %s", body)
			}
			if tc.err != nil && strings.Contains(body, "Out of stock") {
				t.Fatal("invented zero quantity")
			}
		})
	}
}

func TestCreatedListingNotice(t *testing.T) {
	owner := "11111111-1111-1111-1111-111111111111"
	products := &fakes.ProductsService{GetByIDFn: func(context.Context, string) (*productsv1.Product, error) {
		return &productsv1.Product{Id: "product", Name: "Phone", MerchantId: owner}, nil
	}}
	inventory := &fakes.InventoryService{GetFn: func(context.Context, string) (*inventoryv1.Stock, error) {
		return nil, status.Error(codes.NotFound, "pending")
	}}
	req := httptest.NewRequest(http.MethodGet, "/products/product?created=1", nil)
	req.AddCookie(&http.Cookie{Name: auth.AccessCookieName, Value: signedAccessToken(t, owner)})
	rec := httptest.NewRecorder()
	newTestRouter(t, routerDeps{products: products, inventory: inventory}).ServeHTTP(rec, req)
	if !strings.Contains(rec.Body.String(), "Listing created.") || !strings.Contains(rec.Body.String(), "Availability processing") {
		t.Fatal(rec.Body.String())
	}
}
