package shared_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	shared "refurbished-marketplace/services/web/internal/handlers/shared"
	paymentv1 "refurbished-marketplace/shared/proto/payment/v1"
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

func TestBuildHostedPaymentURLUsesCallbackBaseURL(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/cart/checkout", nil)
	req.Host = "shop.example"
	req.Header.Set("X-Forwarded-Proto", "https")

	got := shared.BuildHostedPaymentURL(shared.HostedPaymentConfig{
		GatewayBaseURL:  "https://pay.example",
		PublicBaseURL:   "https://shop.example",
		CallbackBaseURL: "http://web:8080",
	}, req, &paymentv1.CreateHostedPaymentSessionResponse{
		OrderId:          "order-1",
		PaymentSessionId: "sess-1",
		ReturnUrl:        "https://shop.example/orders/order-1",
		CancelUrl:        "https://shop.example/orders/order-1",
	})

	want := "https://pay.example/pay?callback_url=http%3A%2F%2Fweb%3A8080%2Fcallbacks%2Fhosted-payment&cancel_url=https%3A%2F%2Fshop.example%2Forders%2Forder-1&order_id=order-1&payment_session_id=sess-1&return_url=https%3A%2F%2Fshop.example%2Forders%2Forder-1"
	if got != want {
		t.Fatalf("BuildHostedPaymentURL = %q, want %q", got, want)
	}
}
