package tests

import (
	"context"
	"encoding/base64"
	"encoding/json"
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
			if w.Code != 200 || removed != (paymentStatus == "SUCCEEDED") {
				t.Fatalf("status=%d removed=%t", w.Code, removed)
			}
			if paymentStatus != "" && !strings.Contains(w.Body.String(), paymentStatus) {
				t.Fatalf("missing payment status")
			}
		})
	}
}

func TestOrderPageOffersRetryPaymentForOpenPendingOrder(t *testing.T) {
	const paymentURL = "https://pay.example/checkout/open-1"
	for _, tc := range []struct {
		name          string
		status        ordersv1.OrderStatus
		paymentStatus string
		paymentURL    string
		wantButton    bool
	}{
		{name: "open pending order", status: ordersv1.OrderStatus_ORDER_STATUS_PENDING, paymentStatus: "OPEN", paymentURL: paymentURL, wantButton: true},
		{name: "open settled order", status: ordersv1.OrderStatus_ORDER_STATUS_PAID, paymentStatus: "OPEN", paymentURL: paymentURL, wantButton: false},
		{name: "open payment without hosted url", status: ordersv1.OrderStatus_ORDER_STATUS_PENDING, paymentStatus: "OPEN", wantButton: false},
		{name: "expired payment", status: ordersv1.OrderStatus_ORDER_STATUS_PENDING, paymentStatus: "EXPIRED", paymentURL: paymentURL, wantButton: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			orders := &fakes.OrdersService{GetFn: func(_ context.Context, id string) (*ordersv1.Order, error) {
				return &ordersv1.Order{Id: id, BuyerUserId: "user-1", MerchantId: "merchant-1", Status: tc.status, Items: []*ordersv1.OrderItem{{ProductId: "prod-1"}}}, nil
			}}
			checkout := &fakes.CheckoutService{GetFn: func(_ context.Context, r *checkoutv1.GetCheckoutRequest) (*checkoutv1.CheckoutStatus, error) {
				return &checkoutv1.CheckoutStatus{OrderId: r.GetOrderId(), PaymentStatus: tc.paymentStatus, HostedCheckoutUrl: tc.paymentURL}, nil
			}}
			r := httptest.NewRequest(http.MethodGet, "/orders/order-1", nil)
			r.AddCookie(&http.Cookie{Name: auth.AccessCookieName, Value: signedAccessToken(t, "user-1")})
			w := httptest.NewRecorder()
			newTestRouter(t, routerDeps{orders: orders, checkout: checkout}).ServeHTTP(w, r)
			if w.Code != http.StatusOK {
				t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
			}
			body := w.Body.String()
			if strings.Contains(body, "Retry payment") != tc.wantButton {
				t.Fatalf("retry button present=%t, want %t", strings.Contains(body, "Retry payment"), tc.wantButton)
			}
			if tc.wantButton && !strings.Contains(body, `href="`+paymentURL+`"`) {
				t.Fatalf("missing hosted checkout link in %s", body)
			}
		})
	}
}

func TestExpiredOrderRetainsCartAndRotatesOnlyMatchingIntent(t *testing.T) {
	for _, tc := range []struct {
		name         string
		cartID       string
		intentKey    string
		wantRotation bool
	}{
		{"expired intent", "cart-1", "expired-intent", true},
		{"newer checkout", "cart-1", "newer-intent", false},
		{"different cart", "cart-2", "expired-intent", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			orders := &fakes.OrdersService{GetFn: func(_ context.Context, id string) (*ordersv1.Order, error) {
				return &ordersv1.Order{Id: id, BuyerUserId: "user-1", MerchantId: "merchant-1", Items: []*ordersv1.OrderItem{{ProductId: "prod-1"}}}, nil
			}}
			checkout := &fakes.CheckoutService{GetFn: func(_ context.Context, r *checkoutv1.GetCheckoutRequest) (*checkoutv1.CheckoutStatus, error) {
				return &checkoutv1.CheckoutStatus{OrderId: r.GetOrderId(), PaymentStatus: "EXPIRED", IntentKey: "expired-intent"}, nil
			}}
			cart := &fakes.CartService{RemoveManyFn: func(context.Context, string, []string) (*cartv1.Cart, error) {
				t.Fatal("expired payment must retain cart items")
				return nil, nil
			}}
			router := newTestRouter(t, routerDeps{orders: orders, checkout: checkout, cart: cart})
			store := struct {
				CartID string            `json:"c"`
				Keys   map[string]string `json:"k"`
			}{tc.cartID, map[string]string{"merchant-1": tc.intentKey, "merchant-2": "other-intent"}}
			raw, err := json.Marshal(store)
			if err != nil {
				t.Fatal(err)
			}
			cookie := &http.Cookie{Name: "checkout_intent", Value: base64.RawURLEncoding.EncodeToString(raw)}
			for visit := 0; visit < 2; visit++ {
				r := httptest.NewRequest(http.MethodGet, "/orders/order-1", nil)
				r.AddCookie(&http.Cookie{Name: auth.AccessCookieName, Value: signedAccessToken(t, "user-1")})
				r.AddCookie(&http.Cookie{Name: "cart_id", Value: "cart-1"})
				r.AddCookie(cookie)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, r)
				if w.Code != http.StatusOK {
					t.Fatalf("status=%d", w.Code)
				}
				cookies := w.Result().Cookies()
				if !tc.wantRotation || visit > 0 {
					if len(cookies) != 0 {
						t.Fatalf("old order changed current intent: %v", cookies)
					}
					continue
				}
				if len(cookies) != 1 || cookies[0].Name != "checkout_intent" {
					t.Fatalf("missing rotated intent: %v", cookies)
				}
				cookie = cookies[0]
				raw, err = base64.RawURLEncoding.DecodeString(cookie.Value)
				if err != nil {
					t.Fatal(err)
				}
				if err := json.Unmarshal(raw, &store); err != nil {
					t.Fatal(err)
				}
				if store.CartID != "cart-1" || store.Keys["merchant-1"] == "" || store.Keys["merchant-1"] == "expired-intent" || store.Keys["merchant-2"] != "other-intent" {
					t.Fatalf("incorrect rotated intent: %+v", store)
				}
			}
		})
	}
}
