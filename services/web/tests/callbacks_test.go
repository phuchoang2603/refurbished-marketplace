package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/phuchoang2603/refurbished-marketplace/services/web/tests/fakes"
	cartv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/cart/v1"
	ordersv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/orders/v1"
	paymentv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/payment/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestHostedPaymentCallbackForwardsToPaymentService(t *testing.T) {
	var got *paymentv1.HandleGatewayWebhookRequest
	paymentSvc := &fakes.PaymentService{
		HandleWebhookFn: func(ctx context.Context, req *paymentv1.HandleGatewayWebhookRequest) (*paymentv1.HandleGatewayWebhookResponse, error) {
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

func TestHostedPaymentCallbackLeavesCartUntilOrderPaid(t *testing.T) {
	var removed [][]string
	cartSvc := &fakes.CartService{
		RemoveManyFn: func(ctx context.Context, cartID string, productIDs []string) (*cartv1.Cart, error) {
			removed = append(removed, productIDs)
			return &cartv1.Cart{CartId: cartID}, nil
		},
	}
	ordersSvc := &fakes.OrdersService{
		GetFn: func(ctx context.Context, id string) (*ordersv1.Order, error) {
			return &ordersv1.Order{
				Id:          id,
				BuyerUserId: "user-1",
				MerchantId:  "merchant-1",
				Status:      ordersv1.OrderStatus_ORDER_STATUS_PENDING,
				Items:       []*ordersv1.OrderItem{{ProductId: "prod-1", Quantity: 1}},
			}, nil
		},
	}
	paymentSvc := &fakes.PaymentService{
		HandleWebhookFn: func(ctx context.Context, req *paymentv1.HandleGatewayWebhookRequest) (*paymentv1.HandleGatewayWebhookResponse, error) {
			return &paymentv1.HandleGatewayWebhookResponse{}, nil
		},
	}
	router := newTestRouter(t, routerDeps{orders: ordersSvc, cart: cartSvc, payment: paymentSvc})

	for _, callbackStatus := range []string{"SUCCEEDED", "SUCCEEDED", "FAILED", "EXPIRED"} {
		body, err := json.Marshal(map[string]string{
			"order_id":           "11111111-1111-1111-1111-111111111111",
			"payment_session_id": "sess-1",
			"status":             callbackStatus,
		})
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/callbacks/hosted-payment", bytes.NewReader(body))
		req.AddCookie(&http.Cookie{Name: "cart_id", Value: "cart-1"})
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("%s callback status = %d, want %d", callbackStatus, rec.Code, http.StatusNoContent)
		}
	}
	if len(removed) != 0 {
		t.Fatalf("callbacks removed cart lines before the order was paid: %v", removed)
	}
}

func TestHostedPaymentCallbackMapsNotFound(t *testing.T) {
	paymentSvc := &fakes.PaymentService{
		HandleWebhookFn: func(ctx context.Context, req *paymentv1.HandleGatewayWebhookRequest) (*paymentv1.HandleGatewayWebhookResponse, error) {
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
