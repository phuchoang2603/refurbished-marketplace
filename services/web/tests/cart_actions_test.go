package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"refurbished-marketplace/services/web/internal/auth"
	cartv1 "refurbished-marketplace/shared/proto/cart/v1"
	ordersv1 "refurbished-marketplace/shared/proto/orders/v1"
	productsv1 "refurbished-marketplace/shared/proto/products/v1"
)

func TestAddCartItemReturnsHTMLFragmentContract(t *testing.T) {
	cartSvc := &fakeCartService{
		addFn: func(ctx context.Context, cartID, productID string, quantity int32) (*cartv1.Cart, error) {
			return &cartv1.Cart{
				CartId: cartID,
				Items:  []*cartv1.CartItem{{ProductId: productID, Quantity: quantity}},
			}, nil
		},
	}
	productsSvc := &fakeProductsService{
		getByIDFn: func(ctx context.Context, id string) (*productsv1.Product, error) {
			return &productsv1.Product{Id: id, Name: "Phone", PriceCents: 1200, MerchantId: "merchant-1"}, nil
		},
	}
	form := url.Values{"product_id": {"prod-1"}, "quantity": {"2"}}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/cart/items", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	newTestRouter(t, routerDeps{cart: cartSvc, products: productsSvc}).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Content-Type"); got != "text/html; charset=utf-8" {
		t.Fatalf("content-type = %q, want text/html; charset=utf-8", got)
	}
	if got := rec.Header().Get("datastar-selector"); got != "#cart" {
		t.Fatalf("datastar-selector = %q, want #cart", got)
	}
	if got := rec.Header().Get("datastar-mode"); got != "outer" {
		t.Fatalf("datastar-mode = %q, want outer", got)
	}
	body := rec.Body.String()
	for _, want := range []string{`id="cart"`, "Phone", "Estimated total"} {
		if !strings.Contains(body, want) {
			t.Fatalf("body missing %q in %q", want, body)
		}
	}
}

func TestCheckoutClearsCartCookieAndRedirectsToOrder(t *testing.T) {
	cartSvc := &fakeCartService{
		getFn: func(ctx context.Context, cartID string) (*cartv1.Cart, error) {
			return &cartv1.Cart{
				CartId: cartID,
				Items:  []*cartv1.CartItem{{ProductId: "prod-1", Quantity: 1}},
			}, nil
		},
		clearCartFn: func(ctx context.Context, cartID string) error { return nil },
	}
	productsSvc := &fakeProductsService{
		getByIDFn: func(ctx context.Context, id string) (*productsv1.Product, error) {
			return &productsv1.Product{Id: id, Name: "Phone", PriceCents: 1200, MerchantId: "merchant-1"}, nil
		},
	}
	ordersSvc := &fakeOrdersService{
		createFn: func(ctx context.Context, buyerUserID, merchantID string, items []*ordersv1.CreateOrderItem, totalCents int64) (*ordersv1.Order, error) {
			if buyerUserID != "user-1" {
				t.Fatalf("buyerUserID = %q, want user-1", buyerUserID)
			}
			return &ordersv1.Order{Id: "order-1", BuyerUserId: buyerUserID, TotalCents: totalCents}, nil
		},
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/cart/checkout", nil)
	req.AddCookie(&http.Cookie{Name: auth.AccessCookieName, Value: signedAccessToken(t, "user-1")})
	req.AddCookie(&http.Cookie{Name: "cart_id", Value: "cart-1"})

	newTestRouter(t, routerDeps{cart: cartSvc, products: productsSvc, orders: ordersSvc}).ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusSeeOther)
	}
	if got := rec.Header().Get("Location"); got != "/orders/order-1" {
		t.Fatalf("location = %q, want /orders/order-1", got)
	}
	assertCookieCleared(t, rec.Result().Cookies(), "cart_id")
}

func TestAddCartItemDatastarValidationErrorOpensDialog(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/cart/items", strings.NewReader(url.Values{"product_id": {"prod-1"}, "quantity": {"0"}}.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "text/event-stream")

	newTestRouter(t, routerDeps{}).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Content-Type"); !strings.Contains(got, "text/event-stream") {
		t.Fatalf("content-type = %q, want text/event-stream", got)
	}
	body := rec.Body.String()
	for _, want := range []string{"id=\"dialog-root\"", "id=\"error-dialog\"", "Bad request", "invalid request body", "replaceChildren()"} {
		if !strings.Contains(body, want) {
			t.Fatalf("body missing %q in %q", want, body)
		}
	}
}
