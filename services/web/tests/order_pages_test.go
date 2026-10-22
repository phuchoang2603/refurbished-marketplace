package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/phuchoang2603/refurbished-marketplace/services/web/internal/auth"
	"github.com/phuchoang2603/refurbished-marketplace/services/web/tests/fakes"
	cartv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/cart/v1"
	checkoutv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/checkout/v1"
	ordersv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/orders/v1"
)

func TestOrderPageUsesCheckoutTerminalStatus(t *testing.T) {
	for _, paymentStatus := range []string{"SUCCEEDED", "EXPIRED", "PENDING", ""} {
		t.Run(paymentStatus, func(t *testing.T) {
			removed := false
			orders := &fakes.OrdersService{GetFn: func(_ context.Context, id string) (*ordersv1.Order, error) {
				return &ordersv1.Order{Id: id, BuyerUserId: "user-1", MerchantId: "merchant-1", Status: ordersv1.OrderStatus_ORDER_STATUS_PENDING, Items: []*ordersv1.OrderItem{{ProductId: "prod-1"}}}, nil
			}}
			checkout := &fakes.CheckoutService{GetFn: func(_ context.Context, r *checkoutv1.GetCheckoutRequest) (*checkoutv1.CheckoutStatus, error) {
				if r.GetOrderId() != "order-1" || r.GetCheckoutId() != "" || r.GetBuyerUserId() != "user-1" {
					t.Fatalf("status request=%v", r)
				}
				return &checkoutv1.CheckoutStatus{OrderId: "order-1", PaymentStatus: paymentStatus}, nil
			}}
			cart := &fakes.CartService{RemoveManyFn: func(_ context.Context, id string, ids []string) (*cartv1.Cart, error) {
				if id != "cart-1" || len(ids) != 1 || ids[0] != "prod-1" {
					t.Fatalf("wrong cart drain: %s %v", id, ids)
				}
				removed = true
				return &cartv1.Cart{CartId: id, Items: []*cartv1.CartItem{{ProductId: "other-merchant-product"}}}, nil
			}}
			r := httptest.NewRequest(http.MethodGet, "/orders/order-1", nil)
			r.AddCookie(&http.Cookie{Name: auth.AccessCookieName, Value: signedAccessToken(t, "user-1")})
			r.AddCookie(&http.Cookie{Name: "cart_id", Value: "cart-1"})
			w := httptest.NewRecorder()
			newTestRouter(t, routerDeps{orders: orders, checkout: checkout, cart: cart}).ServeHTTP(w, r)
			if w.Code != 200 || removed != (paymentStatus == "SUCCEEDED" || paymentStatus == "EXPIRED") {
				t.Fatalf("status=%d removed=%t", w.Code, removed)
			}
			if paymentStatus != "" && !strings.Contains(w.Body.String(), paymentStatus) {
				t.Fatalf("missing payment status")
			}
		})
	}
}
