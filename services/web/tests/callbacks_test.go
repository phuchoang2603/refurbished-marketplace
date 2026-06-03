package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	paymentv1 "refurbished-marketplace/shared/proto/payment/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestHostedPaymentCallbackForwardsToPaymentService(t *testing.T) {
	var got *paymentv1.HandleGatewayWebhookRequest
	paymentSvc := &fakePaymentService{
		handleWebhookFn: func(ctx context.Context, req *paymentv1.HandleGatewayWebhookRequest) (*paymentv1.HandleGatewayWebhookResponse, error) {
			got = req
			return &paymentv1.HandleGatewayWebhookResponse{}, nil
		},
	}

	body, err := json.Marshal(map[string]string{
		"order_id":           "11111111-1111-1111-1111-111111111111",
		"payment_session_id": "sess-1",
		"status":             "SUCCEEDED",
	})
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/callbacks/hosted-payment", bytes.NewReader(body))
	newTestRouter(t, routerDeps{payment: paymentSvc}).ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d body=%q", rec.Code, http.StatusNoContent, rec.Body.String())
	}
	if got == nil {
		t.Fatal("expected HandleGatewayWebhook to be called")
	}
	if got.GetOrderId() != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("order_id = %q", got.GetOrderId())
	}
	if got.GetPaymentSessionId() != "sess-1" {
		t.Fatalf("payment_session_id = %q", got.GetPaymentSessionId())
	}
	if got.GetStatus() != paymentv1.HostedPaymentSessionStatus_HOSTED_PAYMENT_SESSION_STATUS_SUCCEEDED {
		t.Fatalf("status = %v", got.GetStatus())
	}
}

func TestHostedPaymentCallbackMapsNotFound(t *testing.T) {
	paymentSvc := &fakePaymentService{
		handleWebhookFn: func(ctx context.Context, req *paymentv1.HandleGatewayWebhookRequest) (*paymentv1.HandleGatewayWebhookResponse, error) {
			return nil, status.Error(codes.NotFound, "payment session not found")
		},
	}

	body, err := json.Marshal(map[string]string{
		"order_id":           "11111111-1111-1111-1111-111111111111",
		"payment_session_id": "sess-1",
		"status":             "FAILED",
		"failure_reason":     "Card declined",
	})
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/callbacks/hosted-payment", bytes.NewReader(body))
	newTestRouter(t, routerDeps{payment: paymentSvc}).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}
