package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"refurbished-marketplace/services/web/internal/auth"
	"refurbished-marketplace/services/web/tests/fakes"
	cartv1 "refurbished-marketplace/shared/proto/cart/v1"
	ordersv1 "refurbished-marketplace/shared/proto/orders/v1"
	paymentv1 "refurbished-marketplace/shared/proto/payment/v1"
	productsv1 "refurbished-marketplace/shared/proto/products/v1"
)

func TestAddCartItemRedirectsToCart(t *testing.T) {
	productsSvc := &fakes.ProductsService{
		GetByIDFn: func(ctx context.Context, id string) (*productsv1.Product, error) {
			if id != "prod-1" {
				t.Fatalf("product id = %q, want prod-1", id)
			}
			return &productsv1.Product{Id: id, Name: "Phone", PriceCents: 1200, MerchantId: "merchant-1"}, nil
		},
	}
	cartSvc := &fakes.CartService{
		AddFn: func(ctx context.Context, cartID, productID, merchantID, productName string, quantity int32, unitPriceCents int64) (*cartv1.Cart, error) {
			if merchantID != "merchant-1" {
				t.Fatalf("merchantID = %q, want merchant-1", merchantID)
			}
			if productName != "Phone" || unitPriceCents != 1200 {
				t.Fatalf("snapshot name=%q price=%d", productName, unitPriceCents)
			}
			return &cartv1.Cart{CartId: cartID, Items: []*cartv1.CartItem{{ProductId: productID, Quantity: quantity, MerchantId: merchantID, ProductName: productName, UnitPriceCents: unitPriceCents}}}, nil
		},
	}
	form := url.Values{"product_id": {"prod-1"}, "merchant_id": {"merchant-1"}, "quantity": {"2"}}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/cart/items", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	newTestRouter(t, routerDeps{cart: cartSvc, products: productsSvc}).ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusSeeOther)
	}
	if got := rec.Header().Get("Location"); got != "/cart" {
		t.Fatalf("location = %q, want /cart", got)
	}
}

func TestCheckoutClearsCartCookieAndRedirectsToOrder(t *testing.T) {
	var removed []string
	var batchIDs []string
	cartSvc := &fakes.CartService{
		GetFn: func(ctx context.Context, cartID string) (*cartv1.Cart, error) {
			return &cartv1.Cart{
				CartId: cartID,
				Items: []*cartv1.CartItem{
					{ProductId: "prod-1", Quantity: 1, MerchantId: "merchant-1", ProductName: "Phone", UnitPriceCents: 1000},
					{ProductId: "prod-2", Quantity: 2, MerchantId: "merchant-1", ProductName: "Case", UnitPriceCents: 200},
				},
			}, nil
		},
		RemoveManyFn: func(ctx context.Context, cartID string, productIDs []string) (*cartv1.Cart, error) {
			removed = append([]string{}, productIDs...)
			return &cartv1.Cart{CartId: cartID, Items: nil}, nil
		},
	}
	productsSvc := &fakes.ProductsService{
		GetByIDsFn: func(ctx context.Context, ids []string) (*productsv1.GetProductsByIDsResponse, error) {
			batchIDs = append([]string{}, ids...)
			return &productsv1.GetProductsByIDsResponse{Products: []*productsv1.Product{
				{Id: "prod-1", Name: "Phone", PriceCents: 1200, MerchantId: "merchant-1"},
				{Id: "prod-2", Name: "Case", PriceCents: 250, MerchantId: "merchant-1"},
			}}, nil
		},
	}
	ordersSvc := &fakes.OrdersService{
		CreateFn: func(ctx context.Context, buyerUserID, merchantID string, items []*ordersv1.CreateOrderItem, totalCents int64) (*ordersv1.Order, error) {
			if buyerUserID != "user-1" {
				t.Fatalf("buyerUserID = %q, want user-1", buyerUserID)
			}
			if len(items) != 2 {
				t.Fatalf("items = %d, want 2", len(items))
			}
			if items[0].GetUnitPriceCents() != 1200 || items[1].GetUnitPriceCents() != 250 {
				t.Fatalf("expected SoR batch prices, got %d and %d", items[0].GetUnitPriceCents(), items[1].GetUnitPriceCents())
			}
			if totalCents != 1200+500 {
				t.Fatalf("totalCents = %d, want 1700", totalCents)
			}
			return &ordersv1.Order{Id: "order-1", BuyerUserId: buyerUserID, TotalCents: totalCents}, nil
		},
	}
	paymentSvc := &fakes.PaymentService{
		CreateSessionFn: func(ctx context.Context, req *paymentv1.CreateHostedPaymentSessionRequest) (*paymentv1.CreateHostedPaymentSessionResponse, error) {
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
	if len(batchIDs) != 2 || batchIDs[0] != "prod-1" || batchIDs[1] != "prod-2" {
		t.Fatalf("batch IDs = %v, want [prod-1 prod-2]", batchIDs)
	}
	if len(removed) != 2 || removed[0] != "prod-1" || removed[1] != "prod-2" {
		t.Fatalf("removed = %v, want [prod-1 prod-2]", removed)
	}
}
