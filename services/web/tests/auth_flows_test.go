package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"refurbished-marketplace/services/web/internal/auth"
	usersv1 "refurbished-marketplace/shared/proto/users/v1"
)

func TestLoginSetsCookiesAndRedirects(t *testing.T) {
	usersSvc := &fakeUsersService{
		loginFn: func(ctx context.Context, email, password string) (*usersv1.TokenResponse, error) {
			if email != "buyer@example.com" || password != "secret123" {
				t.Fatalf("unexpected credentials: %q %q", email, password)
			}
			return &usersv1.TokenResponse{
				AccessToken:      "access-token",
				RefreshToken:     "refresh-token",
				ExpiresIn:        60,
				RefreshExpiresIn: 3600,
			}, nil
		},
	}
	form := url.Values{"email": {"buyer@example.com"}, "password": {"secret123"}}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auth/login?next=%2Forders", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	newTestRouter(t, routerDeps{users: usersSvc}).ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusSeeOther)
	}
	if got := rec.Header().Get("Location"); got != "/orders" {
		t.Fatalf("location = %q, want /orders", got)
	}
	assertCookieSet(t, rec.Result().Cookies(), auth.AccessCookieName)
	assertCookieSet(t, rec.Result().Cookies(), auth.RefreshCookieName)
}

func TestRegisterSetsCookiesAndRedirects(t *testing.T) {
	usersSvc := &fakeUsersService{
		createUserFn: func(ctx context.Context, email, password string) (*usersv1.User, error) {
			if email != "buyer@example.com" || password != "secret123" {
				t.Fatalf("unexpected credentials: %q %q", email, password)
			}
			return &usersv1.User{Id: "user-1", Email: email}, nil
		},
		loginFn: func(ctx context.Context, email, password string) (*usersv1.TokenResponse, error) {
			return &usersv1.TokenResponse{
				AccessToken:      "access-token",
				RefreshToken:     "refresh-token",
				ExpiresIn:        60,
				RefreshExpiresIn: 3600,
			}, nil
		},
	}
	form := url.Values{"email": {"buyer@example.com"}, "password": {"secret123"}}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auth/register?next=%2Fcart", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	newTestRouter(t, routerDeps{users: usersSvc}).ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusSeeOther)
	}
	if got := rec.Header().Get("Location"); got != "/cart" {
		t.Fatalf("location = %q, want /cart", got)
	}
	assertCookieSet(t, rec.Result().Cookies(), auth.AccessCookieName)
	assertCookieSet(t, rec.Result().Cookies(), auth.RefreshCookieName)
}

func TestLogoutClearsCookiesAndRedirects(t *testing.T) {
	var gotRefresh string
	usersSvc := &fakeUsersService{
		logoutFn: func(ctx context.Context, refreshToken string) (*usersv1.LogoutResponse, error) {
			gotRefresh = refreshToken
			return &usersv1.LogoutResponse{}, nil
		},
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: auth.RefreshCookieName, Value: "refresh-token"})

	newTestRouter(t, routerDeps{users: usersSvc}).ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusSeeOther)
	}
	if got := rec.Header().Get("Location"); got != "/products" {
		t.Fatalf("location = %q, want /products", got)
	}
	if gotRefresh != "refresh-token" {
		t.Fatalf("logout refresh token = %q, want refresh-token", gotRefresh)
	}
	assertCookieCleared(t, rec.Result().Cookies(), auth.AccessCookieName)
	assertCookieCleared(t, rec.Result().Cookies(), auth.RefreshCookieName)
}

func assertCookieSet(t *testing.T, cookies []*http.Cookie, name string) {
	t.Helper()
	for _, cookie := range cookies {
		if cookie.Name == name {
			if cookie.Value == "" || cookie.MaxAge <= 0 {
				t.Fatalf("cookie %q not set correctly: value=%q maxAge=%d", name, cookie.Value, cookie.MaxAge)
			}
			return
		}
	}
	t.Fatalf("cookie %q not found", name)
}

func assertCookieCleared(t *testing.T, cookies []*http.Cookie, name string) {
	t.Helper()
	for _, cookie := range cookies {
		if cookie.Name == name {
			if cookie.MaxAge >= 0 {
				t.Fatalf("cookie %q not cleared: maxAge=%d", name, cookie.MaxAge)
			}
			return
		}
	}
	t.Fatalf("cookie %q not found", name)
}
