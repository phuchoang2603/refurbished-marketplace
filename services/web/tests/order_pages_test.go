package tests

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
	"refurbished-marketplace/services/web/internal/auth"
	"refurbished-marketplace/services/web/tests/fakes"
	ordersv1 "refurbished-marketplace/shared/proto/orders/v1"
	paymentv1 "refurbished-marketplace/shared/proto/payment/v1"
)

func TestOrderPageShowsHostedPaymentStatus(t *testing.T) {
	ordersSvc := &fakes.OrdersService{
		GetFn: func(ctx context.Context, id string) (*ordersv1.Order, error) {
			return &ordersv1.Order{Id: id, BuyerUserId: "user-1", Status: ordersv1.OrderStatus_ORDER_STATUS_PENDING, TotalCents: 1200, CreatedAt: timestamppb.New(time.Now()), UpdatedAt: timestamppb.New(time.Now())}, nil
		},
	}
	paymentSvc := &fakes.PaymentService{
		GetSessionFn: func(ctx context.Context, orderID string) (*paymentv1.HostedPaymentSession, error) {
			return &paymentv1.HostedPaymentSession{Status: paymentv1.HostedPaymentSessionStatus_HOSTED_PAYMENT_SESSION_STATUS_FAILED, FailureReason: "Card declined"}, nil
		},
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/orders/order-1", nil)
	req.AddCookie(&http.Cookie{Name: auth.AccessCookieName, Value: signedAccessToken(t, "user-1")})

	newTestRouter(t, routerDeps{orders: ordersSvc, payment: paymentSvc}).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body, err := io.ReadAll(rec.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	html := string(body)
	if !strings.Contains(html, "Payment status:") {
		t.Fatalf("expected payment status section in %q", html)
	}
	if !strings.Contains(html, "FAILED") || !strings.Contains(html, "Card declined") {
		t.Fatalf("expected failed hosted payment state in %q", html)
	}
}

func TestResumePaymentRedirectsToHostedSession(t *testing.T) {
	ordersSvc := &fakes.OrdersService{
		GetFn: func(ctx context.Context, id string) (*ordersv1.Order, error) {
			return &ordersv1.Order{Id: id, BuyerUserId: "user-1", Status: ordersv1.OrderStatus_ORDER_STATUS_PENDING, TotalCents: 1200, CreatedAt: timestamppb.New(time.Now()), UpdatedAt: timestamppb.New(time.Now())}, nil
		},
	}
	paymentSvc := &fakes.PaymentService{
		CreateSessionFn: func(ctx context.Context, req *paymentv1.CreateHostedPaymentSessionRequest) (*paymentv1.CreateHostedPaymentSessionResponse, error) {
			return &paymentv1.CreateHostedPaymentSessionResponse{
				OrderId:          req.GetOrderId(),
				PaymentSessionId: "sess-1",
				ReturnUrl:        req.GetReturnUrl(),
				CancelUrl:        req.GetCancelUrl(),
			}, nil
		},
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/orders/order-1/resume-payment", nil)
	req.Host = "localhost:8080"
	req.AddCookie(&http.Cookie{Name: auth.AccessCookieName, Value: signedAccessToken(t, "user-1")})

	newTestRouter(t, routerDeps{orders: ordersSvc, payment: paymentSvc}).ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusSeeOther)
	}
	wantLocation := "http://localhost:8097/pay?callback_url=http%3A%2F%2Flocalhost%3A8080%2Fcallbacks%2Fhosted-payment&cancel_url=http%3A%2F%2Flocalhost%3A8080%2Forders%2Forder-1&order_id=order-1&payment_session_id=sess-1&return_url=http%3A%2F%2Flocalhost%3A8080%2Forders%2Forder-1"
	if got := rec.Header().Get("Location"); got != wantLocation {
		t.Fatalf("location = %q, want %q", got, wantLocation)
	}
}
