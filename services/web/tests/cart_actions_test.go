package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"refurbished-marketplace/services/web/internal/auth"
	cartv1 "refurbished-marketplace/shared/proto/cart/v1"
	ordersv1 "refurbished-marketplace/shared/proto/orders/v1"
	paymentv1 "refurbished-marketplace/shared/proto/payment/v1"
	productsv1 "refurbished-marketplace/shared/proto/products/v1"
)

func TestAddCartItemRedirectsToCart(t *testing.T) {
	cartSvc := &fakeCartService{
		addFn: func(ctx context.Context, cartID, productID, merchantID string, quantity int32) (*cartv1.Cart, error) {
			if merchantID != "merchant-1" {
				t.Fatalf("merchantID = %q, want merchant-1", merchantID)
			}
			return &cartv1.Cart{CartId: cartID, Items: []*cartv1.CartItem{{ProductId: productID, Quantity: quantity, MerchantId: merchantID}}}, nil
		},
	}
	form := url.Values{"product_id": {"prod-1"}, "merchant_id": {"merchant-1"}, "quantity": {"2"}}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/cart/items", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	newTestRouter(t, routerDeps{cart: cartSvc}).ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusSeeOther)
	}
	if got := rec.Header().Get("Location"); got != "/cart" {
		t.Fatalf("location = %q, want /cart", got)
	}
}

func TestCheckoutClearsCartCookieAndRedirectsToOrder(t *testing.T) {
	cartSvc := &fakeCartService{
		getFn: func(ctx context.Context, cartID string) (*cartv1.Cart, error) {
			return &cartv1.Cart{
				CartId: cartID,
				Items:  []*cartv1.CartItem{{ProductId: "prod-1", Quantity: 1, MerchantId: "merchant-1"}},
			}, nil
		},
		clearCartFn: func(ctx context.Context, cartID string) error { return nil },
	}
	productsSvc := &fakeProductsService{
		getByIDFn: func(ctx context.Context, id string) (*productsv1.Product, error) {
			return &productsv1.Product{Id: id, Name: "Phone", PriceCents: 1200, MerchantId: "merchant-1"}, nil
		},
	}
	ordersSvc := &fakeOrdersService{
		createFn: func(ctx context.Context, buyerUserID, merchantID string, items []*ordersv1.CreateOrderItem, totalCents int64) (*ordersv1.Order, error) {
			if buyerUserID != "user-1" {
				t.Fatalf("buyerUserID = %q, want user-1", buyerUserID)
			}
			return &ordersv1.Order{Id: "order-1", BuyerUserId: buyerUserID, TotalCents: totalCents}, nil
		},
	}
	paymentSvc := &fakePaymentService{
		createSessionFn: func(ctx context.Context, req *paymentv1.CreateHostedPaymentSessionRequest) (*paymentv1.CreateHostedPaymentSessionResponse, error) {
			if req.GetOrderId() != "order-1" {
				t.Fatalf("orderID = %q, want order-1", req.GetOrderId())
			}
			if req.GetReturnUrl() != "http://localhost:8080/orders/order-1" {
				t.Fatalf("return_url = %q", req.GetReturnUrl())
			}
			return &paymentv1.CreateHostedPaymentSessionResponse{
				OrderId:          "order-1",
				PaymentSessionId: "sess-1",
				ReturnUrl:        req.GetReturnUrl(),
				CancelUrl:        req.GetCancelUrl(),
			}, nil
		},
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/cart/checkout", strings.NewReader(url.Values{"merchant_id": {"merchant-1"}}.Encode()))
	req.Host = "localhost:8080"
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: auth.AccessCookieName, Value: signedAccessToken(t, "user-1")})
	req.AddCookie(&http.Cookie{Name: "cart_id", Value: "cart-1"})

	newTestRouter(t, routerDeps{cart: cartSvc, products: productsSvc, orders: ordersSvc, payment: paymentSvc}).ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusSeeOther)
	}
	wantLocation := "http://localhost:8097/pay?callback_url=http%3A%2F%2Flocalhost%3A8080%2Fcallbacks%2Fhosted-payment&cancel_url=http%3A%2F%2Flocalhost%3A8080%2Forders%2Forder-1&order_id=order-1&payment_session_id=sess-1&return_url=http%3A%2F%2Flocalhost%3A8080%2Forders%2Forder-1"
	if got := rec.Header().Get("Location"); got != wantLocation {
		t.Fatalf("location = %q, want %q", got, wantLocation)
	}
	assertCookieCleared(t, rec.Result().Cookies(), "cart_id")
}
