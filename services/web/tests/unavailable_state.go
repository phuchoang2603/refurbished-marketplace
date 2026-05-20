package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	cartv1 "refurbished-marketplace/shared/proto/cart/v1"
	ordersv1 "refurbished-marketplace/shared/proto/orders/v1"
	productsv1 "refurbished-marketplace/shared/proto/products/v1"

	"refurbished-marketplace/services/web/internal/auth"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestProductsPageShowsUnavailableState(t *testing.T) {
	productsSvc := &fakeProductsService{
		listFn: func(ctx context.Context, limit, offset int32) (*productsv1.ListProductsResponse, error) {
			return nil, status.Error(codes.Unavailable, "products service unavailable")
		},
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/products", nil)

	newTestRouter(t, routerDeps{products: productsSvc}).ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
	if body := rec.Body.String(); !strings.Contains(body, "Products unavailable") {
		t.Fatalf("body = %q, want products unavailable state", body)
	}
}

func TestProtectedOrdersPageShowsUnavailableState(t *testing.T) {
	ordersSvc := &fakeOrdersService{
		listFn: func(ctx context.Context, buyerUserID string, limit, offset int32) (*ordersv1.ListOrdersByBuyerResponse, error) {
			return nil, status.Error(codes.Unavailable, "orders service unavailable")
		},
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/orders", nil)
	req.AddCookie(&http.Cookie{Name: auth.AccessCookieName, Value: signedAccessToken(t, "user-1")})

	newTestRouter(t, routerDeps{orders: ordersSvc}).ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
	if body := rec.Body.String(); !strings.Contains(body, "Orders unavailable") {
		t.Fatalf("body = %q, want orders unavailable state", body)
	}
}

func TestCheckoutMutationShowsUnavailablePopup(t *testing.T) {
	cartSvc := &fakeCartService{
		getFn: func(ctx context.Context, cartID string) (*cartv1.Cart, error) {
			return &cartv1.Cart{CartId: cartID, Items: []*cartv1.CartItem{{ProductId: "prod-1", Quantity: 1}}}, nil
		},
	}
	productsSvc := &fakeProductsService{
		getByIDFn: func(ctx context.Context, id string) (*productsv1.Product, error) {
			return nil, status.Error(codes.Unavailable, "products service unavailable")
		},
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/cart/checkout", nil)
	req.AddCookie(&http.Cookie{Name: auth.AccessCookieName, Value: signedAccessToken(t, "user-1")})
	req.AddCookie(&http.Cookie{Name: "cart_id", Value: "cart-1"})

	newTestRouter(t, routerDeps{cart: cartSvc, products: productsSvc}).ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
	if body := rec.Body.String(); !strings.Contains(body, "products service unavailable") {
		t.Fatalf("body = %q, want unavailable popup", body)
	}
}
