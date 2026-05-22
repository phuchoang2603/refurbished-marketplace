package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"refurbished-marketplace/services/web/internal/auth"
	ordersv1 "refurbished-marketplace/shared/proto/orders/v1"
)

func TestProtectedGetShowsUnauthorizedPopup(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/orders", nil)

	newTestRouter(t, routerDeps{}).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestProtectedPostShowsUnauthorizedPopup(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/cart/checkout", nil)

	newTestRouter(t, routerDeps{}).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestAuthenticatedProtectedRouteProceeds(t *testing.T) {
	ordersSvc := &fakeOrdersService{
		listFn: func(ctx context.Context, buyerUserID string, limit, offset int32) (*ordersv1.ListOrdersByBuyerResponse, error) {
			if buyerUserID != "user-1" {
				t.Fatalf("buyerUserID = %q, want user-1", buyerUserID)
			}
			return &ordersv1.ListOrdersByBuyerResponse{}, nil
		},
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/orders", nil)
	req.AddCookie(&http.Cookie{Name: auth.AccessCookieName, Value: signedAccessToken(t, "user-1")})

	newTestRouter(t, routerDeps{orders: ordersSvc}).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}
