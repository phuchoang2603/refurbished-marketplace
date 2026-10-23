package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/phuchoang2603/refurbished-marketplace/services/web/internal/auth"
	"github.com/phuchoang2603/refurbished-marketplace/services/web/tests/fakes"
	ordersv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/orders/v1"
	usersv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/users/v1"
)

func TestExpiredAccessIsRefreshedFromRefreshCookie(t *testing.T) {
	newAccess := signedAccessToken(t, "user-1")
	usersSvc := &fakes.UsersService{
		RefreshFn: func(ctx context.Context, refreshToken string) (*usersv1.TokenResponse, error) {
			if refreshToken != "old-refresh" {
				t.Fatalf("refresh token = %q, want old-refresh", refreshToken)
			}
			return &usersv1.TokenResponse{AccessToken: newAccess, RefreshToken: "new-refresh", ExpiresIn: 60, RefreshExpiresIn: 3600}, nil
		},
	}
	ordersSvc := &fakes.OrdersService{
		ListFn: func(ctx context.Context, buyerUserID string, limit, offset int32) (*ordersv1.ListOrdersByBuyerResponse, error) {
			if buyerUserID != "user-1" {
				t.Fatalf("buyerUserID = %q, want user-1", buyerUserID)
			}
			return &ordersv1.ListOrdersByBuyerResponse{}, nil
		},
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/orders", nil)
	req.AddCookie(&http.Cookie{Name: auth.AccessCookieName, Value: "expired"})
	req.AddCookie(&http.Cookie{Name: auth.RefreshCookieName, Value: "old-refresh"})

	newTestRouter(t, routerDeps{users: usersSvc, orders: ordersSvc}).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	assertCookieSet(t, rec.Result().Cookies(), auth.AccessCookieName)
	assertCookieSet(t, rec.Result().Cookies(), auth.RefreshCookieName)
}

func TestFailedRefreshClearsCookies(t *testing.T) {
	usersSvc := &fakes.UsersService{
		RefreshFn: func(ctx context.Context, refreshToken string) (*usersv1.TokenResponse, error) {
			return nil, context.Canceled
		},
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/orders", nil)
	req.AddCookie(&http.Cookie{Name: auth.RefreshCookieName, Value: "revoked"})

	newTestRouter(t, routerDeps{users: usersSvc}).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	assertCookieCleared(t, rec.Result().Cookies(), auth.RefreshCookieName)
}

func TestCrossOriginPostIsRejected(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(""))
	req.Header.Set("Sec-Fetch-Site", "cross-site")

	newTestRouter(t, routerDeps{}).ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}
