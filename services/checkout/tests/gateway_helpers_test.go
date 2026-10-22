package tests

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/phuchoang2603/refurbished-marketplace/services/checkout/internal/service"
	checkoutv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/checkout/v1"
	usersv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/users/v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var accountCreatedAt = time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)

type testUsers struct{}

func (testUsers) GetUserByID(_ context.Context, request *usersv1.GetUserByIDRequest, _ ...grpc.CallOption) (*usersv1.User, error) {
	return &usersv1.User{Id: request.GetId(), CreatedAt: timestamppb.New(accountCreatedAt)}, nil
}

func testConfig(endpoint string) service.Config {
	return service.Config{PaymentGatewayURL: endpoint, PaymentGatewayAPIKey: "test-api-key", PaymentGatewayWebhookSecret: "test-webhook-secret", CheckoutReturnBaseURL: "https://example.com", PaymentDeadlineGrace: 2 * time.Minute}
}

func newTestService(db *sql.DB) *service.Service {
	return service.New(db, testConfig("http://gateway.invalid"), testUsers{})
}

func gatewayService(t *testing.T, db *sql.DB, handler http.HandlerFunc) *service.Service {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return service.New(db, testConfig(server.URL), testUsers{})
}

func writePayment(t *testing.T, w http.ResponseWriter) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]any{"payment_id": "payment-1", "session_id": "session-1", "checkout_token": "not-logged", "checkout_url": "https://pay.example.com/checkout/session-1?token=private", "expires_at": time.Now().Add(15 * time.Minute).UTC().Format(time.RFC3339), "status": "OPEN", "merchant_id": "merchant-derived", "customer_id": "customer-derived"}); err != nil {
		t.Error(err)
	}
}

func signWebhook(body []byte, secret string, now time.Time) string {
	timestamp := fmt.Sprint(now.Unix())
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(timestamp + "."))
	_, _ = mac.Write(body)
	return "t=" + timestamp + ",v1=" + hex.EncodeToString(mac.Sum(nil))
}

func webhookRequest(t *testing.T, eventID, eventType, orderID string) *checkoutv1.HandleGatewayWebhookRequest {
	t.Helper()
	status := "DECLINED"
	if eventType == "payment.succeeded" {
		status = "SUCCEEDED"
	}
	if eventType == "payment.expired" {
		status = "EXPIRED"
	}
	body, err := json.Marshal(map[string]any{"event_id": eventID, "type": eventType, "payment_id": "payment-1", "order_reference": orderID, "attempt_id": "attempt-1", "status": status, "occurred_at": time.Now().UTC().Format(time.RFC3339)})
	if err != nil {
		t.Fatal(err)
	}
	return &checkoutv1.HandleGatewayWebhookRequest{RawBody: body, SignatureHeader: signWebhook(body, "test-webhook-secret", time.Now())}
}

func sendWebhook(t *testing.T, svc *service.Service, eventID, eventType, orderID string) *checkoutv1.HandleGatewayWebhookResponse {
	t.Helper()
	response, err := svc.HandleGatewayWebhook(t.Context(), webhookRequest(t, eventID, eventType, orderID))
	if err != nil {
		t.Fatal(err)
	}
	if response.GetResult() != checkoutv1.GatewayWebhookResult_GATEWAY_WEBHOOK_RESULT_ACCEPTED {
		t.Fatalf("webhook result %v", response)
	}
	return response
}
