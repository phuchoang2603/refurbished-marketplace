package tests

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	webAuth "refurbished-marketplace/services/web/internal/auth"
	"refurbished-marketplace/services/web/internal/handlers"
	authconfig "refurbished-marketplace/shared/auth/config"
)

func TestRegisterRouteReturnsHTMLForm(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/auth/register", nil)
	mux := http.NewServeMux()
	handlers.New(nil, nil, nil, nil, nil, authconfig.DefaultConfig("secret")).Register(mux)

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	for _, want := range []string{`action="/users"`, `data-on-submit__prevent=`, `contentType`, `data-indicator-fetching`, `autocomplete="new-password"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("body %q missing %q", body, want)
		}
	}
}

func TestLogoutWithoutAccessTokenClearsCookies(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	mux := http.NewServeMux()
	handlers.New(nil, nil, nil, nil, nil, authconfig.DefaultConfig("secret")).Register(mux)

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var clearedAccess, clearedRefresh bool
	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == webAuth.AccessCookieName && cookie.MaxAge < 0 {
			clearedAccess = true
		}
		if cookie.Name == webAuth.RefreshCookieName && cookie.MaxAge < 0 {
			clearedRefresh = true
		}
	}
	if !clearedAccess || !clearedRefresh {
		t.Fatalf("cleared access=%v refresh=%v, want both true", clearedAccess, clearedRefresh)
	}
}

func TestLoginRouteReturnsHTMLForm(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/auth/login", nil)
	mux := http.NewServeMux()
	handlers.New(nil, nil, nil, nil, nil, authconfig.DefaultConfig("secret")).Register(mux)

	mux.ServeHTTP(rec, req)

	if got := rec.Header().Get("Content-Type"); got != "text/html; charset=utf-8" {
		t.Fatalf("content-type = %q, want text/html; charset=utf-8", got)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `action="/auth/login"`) || !strings.Contains(body, `data-on-submit__prevent=`) {
		t.Fatalf("body %q missing login form Datastar markup", body)
	}
}

func TestRequireAccessTokenUnauthorizedReturnsHTML(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/orders", nil)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called")
	})

	webAuth.RequireAccessToken(authconfig.DefaultConfig("secret"), next, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`<h1>Unauthorized</h1>`))
	}).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	if got := rec.Header().Get("Content-Type"); got != "text/html; charset=utf-8" {
		t.Fatalf("content-type = %q, want text/html; charset=utf-8", got)
	}
	if !strings.Contains(rec.Body.String(), "Unauthorized") {
		t.Fatalf("body %q missing unauthorized page", rec.Body.String())
	}
}

func TestRequireAccessTokenAcceptsCookie(t *testing.T) {
	const secret = "secret"
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/orders", nil)
	req.AddCookie(&http.Cookie{Name: webAuth.AccessCookieName, Value: signedAccessToken(t, secret, "user-1")})
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := webAuth.UserIDFromContext(r.Context())
		if !ok || userID != "user-1" {
			t.Fatalf("userID = %q, %v; want user-1, true", userID, ok)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	webAuth.RequireAccessToken(authconfig.DefaultConfig(secret), next, nil).ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestSetTokenCookiesUsesRefreshExpiry(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)

	webAuth.SetTokenCookies(rec, req, "access", "refresh", 60, 3600)

	var accessMaxAge, refreshMaxAge int
	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == webAuth.AccessCookieName {
			accessMaxAge = cookie.MaxAge
		}
		if cookie.Name == webAuth.RefreshCookieName {
			refreshMaxAge = cookie.MaxAge
		}
	}
	if accessMaxAge != 60 {
		t.Fatalf("access MaxAge = %d, want 60", accessMaxAge)
	}
	if refreshMaxAge != 3600 {
		t.Fatalf("refresh MaxAge = %d, want 3600", refreshMaxAge)
	}
}
