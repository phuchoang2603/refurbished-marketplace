package tests

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/phuchoang2603/refurbished-marketplace/services/web/tests/fakes"
	checkoutv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/checkout/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGatewayWebhookRawForwardingAndResults(t *testing.T) {
	for _, tc := range []struct {
		name   string
		result checkoutv1.GatewayWebhookResult
		err    error
		want   int
	}{
		{"accepted", checkoutv1.GatewayWebhookResult_GATEWAY_WEBHOOK_RESULT_ACCEPTED, nil, 204},
		{"duplicate", checkoutv1.GatewayWebhookResult_GATEWAY_WEBHOOK_RESULT_DUPLICATE, nil, 204},
		{"rejected", checkoutv1.GatewayWebhookResult_GATEWAY_WEBHOOK_RESULT_REJECTED, nil, 400},
		{"unspecified", checkoutv1.GatewayWebhookResult_GATEWAY_WEBHOOK_RESULT_UNSPECIFIED, nil, 503},
		{"invalid", 0, status.Error(codes.InvalidArgument, "secret detail"), 400},
		{"unavailable", 0, status.Error(codes.Unavailable, "secret detail"), 503},
		{"internal", 0, errors.New("secret detail"), 503},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := []byte(" { \"type\" : \"payment.succeeded\", \"unknown\": [1,2] }\r\n")
			called := false
			checkout := &fakes.CheckoutService{WebhookFn: func(_ context.Context, request *checkoutv1.HandleGatewayWebhookRequest) (*checkoutv1.HandleGatewayWebhookResponse, error) {
				called = true
				if !bytes.Equal(request.RawBody, body) || request.SignatureHeader != "t=123,v1=00" {
					t.Fatalf("request changed: %v", request)
				}
				return &checkoutv1.HandleGatewayWebhookResponse{Result: tc.result}, tc.err
			}}
			r := httptest.NewRequest(http.MethodPost, "/webhooks/payment-gateway", bytes.NewReader(body))
			r.Header.Set("Gateway-Signature", "t=123,v1=00")
			r.AddCookie(&http.Cookie{Name: "cart_id", Value: "cart-1"})
			w := httptest.NewRecorder()
			newTestRouter(t, routerDeps{checkout: checkout}).ServeHTTP(w, r)
			if !called || w.Code != tc.want || strings.Contains(w.Body.String(), "secret detail") || w.Header().Get("Set-Cookie") != "" {
				t.Fatalf("response = %d %s", w.Code, w.Body.String())
			}
		})
	}
}

func TestGatewayWebhookBodyLimit(t *testing.T) {
	for _, size := range []int{64 << 10, (64 << 10) + 1} {
		t.Run(strconv.Itoa(size), func(t *testing.T) {
			called := false
			checkout := &fakes.CheckoutService{WebhookFn: func(_ context.Context, request *checkoutv1.HandleGatewayWebhookRequest) (*checkoutv1.HandleGatewayWebhookResponse, error) {
				called = true
				if len(request.RawBody) != size {
					t.Fatal("body truncated")
				}
				return &checkoutv1.HandleGatewayWebhookResponse{Result: checkoutv1.GatewayWebhookResult_GATEWAY_WEBHOOK_RESULT_ACCEPTED}, nil
			}}
			r := httptest.NewRequest(http.MethodPost, "/webhooks/payment-gateway", strings.NewReader(strings.Repeat("x", size)))
			r.ContentLength = -1
			w := httptest.NewRecorder()
			newTestRouter(t, routerDeps{checkout: checkout}).ServeHTTP(w, r)
			want := http.StatusNoContent
			if size > 64<<10 {
				want = http.StatusBadRequest
			}
			if w.Code != want || called != (size <= 64<<10) {
				t.Fatalf("size=%d response=%d forwarded=%t", size, w.Code, called)
			}
		})
	}
}

type brokenWebhookBody struct{}

func (brokenWebhookBody) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }
func (brokenWebhookBody) Close() error             { return nil }

func TestGatewayWebhookMissingSignatureAndBrokenBody(t *testing.T) {
	checkout := &fakes.CheckoutService{WebhookFn: func(_ context.Context, request *checkoutv1.HandleGatewayWebhookRequest) (*checkoutv1.HandleGatewayWebhookResponse, error) {
		if request.SignatureHeader != "" {
			t.Fatal("invented signature")
		}
		return &checkoutv1.HandleGatewayWebhookResponse{Result: checkoutv1.GatewayWebhookResult_GATEWAY_WEBHOOK_RESULT_REJECTED}, nil
	}}
	router := newTestRouter(t, routerDeps{checkout: checkout})
	for _, body := range []io.ReadCloser{io.NopCloser(strings.NewReader("{}")), brokenWebhookBody{}} {
		r := httptest.NewRequest(http.MethodPost, "/webhooks/payment-gateway", nil)
		r.Body = body
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)
		if w.Code != 400 {
			t.Fatalf("status=%d", w.Code)
		}
	}
}

func TestLegacyCallbackRemovedAndWebhookRequiresCheckout(t *testing.T) {
	router := newTestRouter(t, routerDeps{})
	for _, tc := range []struct {
		path string
		want int
	}{{"/callbacks/hosted-payment", 404}, {"/webhooks/payment-gateway", 503}} {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodPost, tc.path, strings.NewReader("{}")))
		if w.Code != tc.want {
			t.Fatalf("%s status=%d", tc.path, w.Code)
		}
	}
}
