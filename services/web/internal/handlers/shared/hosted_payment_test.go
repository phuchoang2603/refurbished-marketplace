package shared_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	shared "github.com/phuchoang2603/refurbished-marketplace/services/web/internal/handlers/shared"
)

func TestRequestBaseURLUsesForwardedProto(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/cart/checkout", nil)
	req.Host = "shop.example"
	req.Header.Set("X-Forwarded-Proto", "https")

	if got := shared.RequestBaseURL(req); got != "https://shop.example" {
		t.Fatalf("RequestBaseURL = %q, want https://shop.example", got)
	}
}

func TestRequestBaseURLUsesCfVisitor(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/cart/checkout", nil)
	req.Host = "shop.example"
	req.Header.Set("Cf-Visitor", `{"scheme":"https"}`)

	if got := shared.RequestBaseURL(req); got != "https://shop.example" {
		t.Fatalf("RequestBaseURL = %q, want https://shop.example", got)
	}
}
