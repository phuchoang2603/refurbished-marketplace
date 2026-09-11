package tests

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/phuchoang2603/refurbished-marketplace/services/web/internal/auth"
	"github.com/phuchoang2603/refurbished-marketplace/services/web/tests/fakes"
	cartv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/cart/v1"
	ordersv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/orders/v1"
	paymentv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/payment/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestOrderPageShowsHostedPaymentStatus(t *testing.T) {
	var removed []string
	ordersSvc := &fakes.OrdersService{
		GetFn: func(ctx context.Context, id string) (*ordersv1.Order, error) {
			return &ordersv1.Order{
				Id:          id,
				BuyerUserId: "user-1",
				MerchantId:  "merchant-1",
				Status:      ordersv1.OrderStatus_ORDER_STATUS_PENDING,
				TotalCents:  1200,
				Items:       []*ordersv1.OrderItem{{ProductId: "prod-1", Quantity: 1}},
				CreatedAt:   timestamppb.New(time.Now()),
				UpdatedAt:   timestamppb.New(time.Now()),
			}, nil
		},
	}
	paymentSvc := &fakes.PaymentService{
		GetSessionFn: func(ctx context.Context, orderID string) (*paymentv1.HostedPaymentSession, error) {
			return &paymentv1.HostedPaymentSession{Status: paymentv1.HostedPaymentSessionStatus_HOSTED_PAYMENT_SESSION_STATUS_FAILED, FailureReason: "Card declined"}, nil
		},
	}
	cartSvc := &fakes.CartService{
		RemoveManyFn: func(ctx context.Context, cartID string, productIDs []string) (*cartv1.Cart, error) {
			removed = append([]string{}, productIDs...)
			return &cartv1.Cart{CartId: cartID}, nil
		},
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/orders/order-1", nil)
	req.AddCookie(&http.Cookie{Name: auth.AccessCookieName, Value: signedAccessToken(t, "user-1")})
	req.AddCookie(&http.Cookie{Name: "cart_id", Value: "cart-1"})

	newTestRouter(t, routerDeps{orders: ordersSvc, payment: paymentSvc, cart: cartSvc}).ServeHTTP(rec, req)

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
	if strings.Contains(html, "Resume payment") {
		t.Fatalf("did not expect resume payment after failed hosted session in %q", html)
	}
	if len(removed) != 1 || removed[0] != "prod-1" {
		t.Fatalf("removed = %v, want [prod-1] after failed payment", removed)
	}
}
