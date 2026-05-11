package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"refurbished-marketplace/services/web/internal/handlers"
	"refurbished-marketplace/services/web/internal/views"
	authconfig "refurbished-marketplace/shared/auth/config"
)

func TestHomeRedirectsToCatalog(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mux := http.NewServeMux()
	handlers.New(nil, nil, nil, nil, nil, authconfig.DefaultConfig("secret")).Register(mux)

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusSeeOther)
	}
	if got := rec.Header().Get("Location"); got != "/products" {
		t.Fatalf("Location = %q, want /products", got)
	}
}

func TestProductsPageRendersHTML(t *testing.T) {
	html := renderToString(t, views.ProductsPage([]views.ProductView{{ID: "p1", Name: "Phone", PriceCents: 1000}}))
	if !strings.Contains(html, "<title>Products</title>") {
		t.Fatalf("missing title in %q", html)
	}
	if !strings.Contains(html, "Phone") {
		t.Fatalf("missing product name in %q", html)
	}
	if !strings.Contains(html, `href="/static/app.css"`) || !strings.Contains(html, `class="product-grid"`) {
		t.Fatalf("missing shell stylesheet or product grid in %q", html)
	}
	for _, want := range []string{`href="/cart"`, `href="/orders"`, `href="/auth/login"`} {
		if !strings.Contains(html, want) {
			t.Fatalf("missing navigation link %q in %q", want, html)
		}
	}
}

func TestShellReflectsGuestAndAuthenticatedNavState(t *testing.T) {
	guestHTML := renderToString(t, views.LoginPage())
	for _, want := range []string{`data-signals=`, `authenticated`, `data-show="!$session.authenticated"`, `data-show="$session.authenticated"`} {
		if !strings.Contains(guestHTML, want) {
			t.Fatalf("guest shell missing %q in %q", want, guestHTML)
		}
	}

	ctx := views.WithAuthState(context.Background(), views.AuthState{Authenticated: true})
	authHTML := renderWithContext(t, ctx, views.LoginPage())
	if !strings.Contains(authHTML, `authenticated`) || !strings.Contains(authHTML, `true`) {
		t.Fatalf("authenticated shell missing true auth signal in %q", authHTML)
	}
}

func TestProductDetailIncludesAddToCartAction(t *testing.T) {
	html := renderToString(t, views.ProductDetailSection(views.ProductView{ID: "p1", Name: "Phone", PriceCents: 1000}))
	if !strings.Contains(html, `data-on-submit__prevent=`) || !strings.Contains(html, `contentType`) || !strings.Contains(html, `data-indicator-fetching`) {
		t.Fatalf("missing Datastar cart action in %q", html)
	}
}

func TestCartSectionRendersEmptyState(t *testing.T) {
	html := renderToString(t, views.CartSection(views.CartView{CartID: "c1"}))
	if !strings.Contains(html, "Your cart is empty.") {
		t.Fatalf("missing empty state in %q", html)
	}
}

func TestCartSectionRendersProductDetailsAndTotal(t *testing.T) {
	html := renderToString(t, views.CartSection(views.CartView{
		CartID: "c1",
		Items: []views.CartItemView{{
			ProductID:      "p1",
			ProductName:    "Phone",
			ProductPrice:   1000,
			LineTotalCents: 2000,
			Available:      true,
			Quantity:       2,
		}},
		EstimatedTotalCents: 2000,
	}))
	for _, want := range []string{"Phone", "Estimated total", "$2000", `id="cart-item-p1"`, `class="responsive-table"`, `data-label="Product"`} {
		if !strings.Contains(html, want) {
			t.Fatalf("body %q missing %q", html, want)
		}
	}
}

func TestLoginPageIncludesDatastarForm(t *testing.T) {
	html := renderToString(t, views.LoginPage())
	if !strings.Contains(html, `data-on-submit__prevent=`) || !strings.Contains(html, `contentType`) || !strings.Contains(html, `data-indicator-fetching`) || !strings.Contains(html, `data-attr-disabled`) {
		t.Fatalf("missing Datastar login action in %q", html)
	}
}
