package tests

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/a-h/templ"
	webAuth "refurbished-marketplace/services/web/internal/auth"
	"refurbished-marketplace/services/web/internal/handlers"
	"refurbished-marketplace/services/web/internal/views"
	authconfig "refurbished-marketplace/shared/auth/config"
)

func renderToString(t *testing.T, c templ.Component) string {
	t.Helper()
	var buf bytes.Buffer
	if err := c.Render(context.Background(), &buf); err != nil {
		t.Fatalf("render failed: %v", err)
	}
	return buf.String()
}

func TestProductsPageRendersHTML(t *testing.T) {
	html := renderToString(t, views.ProductsPage([]views.ProductView{{ID: "p1", Name: "Phone", PriceCents: 1000}}))
	if !strings.Contains(html, "<title>Products</title>") {
		t.Fatalf("missing title in %q", html)
	}
	if !strings.Contains(html, "Phone") {
		t.Fatalf("missing product name in %q", html)
	}
}

func TestCartSectionRendersEmptyState(t *testing.T) {
	html := renderToString(t, views.CartSection(views.CartView{CartID: "c1"}))
	if !strings.Contains(html, "Your cart is empty.") {
		t.Fatalf("missing empty state in %q", html)
	}
}

func TestLoginPageIncludesDatastarForm(t *testing.T) {
	html := renderToString(t, views.LoginPage())
	if !strings.Contains(html, `data-on-submit__prevent="@post('/auth/login')"`) {
		t.Fatalf("missing Datastar login action in %q", html)
	}
}

func TestProductDetailIncludesAddToCartAction(t *testing.T) {
	html := renderToString(t, views.ProductDetailSection(views.ProductView{ID: "p1", Name: "Phone", PriceCents: 1000}))
	if !strings.Contains(html, `data-on-submit__prevent="@post('/cart/items')"`) {
		t.Fatalf("missing Datastar cart action in %q", html)
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
	if !strings.Contains(body, `action="/auth/login"`) || !strings.Contains(body, `data-on-submit__prevent="@post('/auth/login')"`) {
		t.Fatalf("body %q missing login form Datastar markup", body)
	}
}

func TestRequireAccessTokenUnauthorizedReturnsHTML(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/orders", nil)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called")
	})

	webAuth.RequireAccessToken(authconfig.DefaultConfig("secret"), next).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	if got := rec.Header().Get("Content-Type"); got != "text/html; charset=utf-8" {
		t.Fatalf("content-type = %q, want text/html; charset=utf-8", got)
	}
	if !strings.Contains(rec.Body.String(), "<h1>Unauthorized</h1>") {
		t.Fatalf("body %q missing unauthorized page", rec.Body.String())
	}
}
