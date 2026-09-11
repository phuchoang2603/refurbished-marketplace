package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/phuchoang2603/refurbished-marketplace/services/web/internal/auth"
	"github.com/phuchoang2603/refurbished-marketplace/services/web/tests/fakes"
	cartv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/cart/v1"
	ordersv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/orders/v1"
	paymentv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/payment/v1"
	productsv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/products/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func TestCheckoutRedirectsToHostedPaymentWithoutClearingCart(t *testing.T) {
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
		CreateFn: func(ctx context.Context, buyerUserID, merchantID string, items []*ordersv1.CreateOrderItem, totalCents int64, idempotencyKey string) (*ordersv1.Order, error) {
			if idempotencyKey != "intent-1" {
				t.Fatalf("idempotencyKey = %q, want intent-1", idempotencyKey)
			}
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
			return &ordersv1.Order{
				Id:          "order-1",
				BuyerUserId: buyerUserID,
				MerchantId:  merchantID,
				Status:      ordersv1.OrderStatus_ORDER_STATUS_PENDING,
				TotalCents:  totalCents,
				Items: []*ordersv1.OrderItem{
					{ProductId: items[0].GetProductId(), Quantity: items[0].GetQuantity()},
					{ProductId: items[1].GetProductId(), Quantity: items[1].GetQuantity()},
				},
			}, nil
		},
	}
	var reservedOrderID string
	productsSvc.ReserveFn = func(ctx context.Context, orderID, merchantID string, totalCents int64, items []*productsv1.ReserveStockItem) error {
		reservedOrderID = orderID
		if merchantID != "merchant-1" {
			t.Fatalf("reserve merchantID = %q, want merchant-1", merchantID)
		}
		if len(items) != 2 {
			t.Fatalf("reserve items = %d, want 2", len(items))
		}
		return nil
	}
	paymentSvc := &fakes.PaymentService{
		CreateSessionFn: func(ctx context.Context, req *paymentv1.CreateHostedPaymentSessionRequest) (*paymentv1.CreateHostedPaymentSessionResponse, error) {
			if req.GetOrderId() != "order-1" {
				t.Fatalf("orderID = %q, want order-1", req.GetOrderId())
			}
			if req.GetMerchant().GetId() != "merchant-1" {
				t.Fatalf("merchant_id = %q, want merchant-1", req.GetMerchant().GetId())
			}
			if req.GetBuyer().GetId() != "user-1" {
				t.Fatalf("buyer_id = %q, want user-1", req.GetBuyer().GetId())
			}
			if req.GetBuyer().GetEmail() != "buyer@example.com" {
				t.Fatalf("buyer email = %q", req.GetBuyer().GetEmail())
			}
			if req.GetShippingAddress().GetCountry() != "US" || req.GetShippingAddress().GetLine1() == "" {
				t.Fatalf("shipping = %+v", req.GetShippingAddress())
			}
			if req.GetTotalCents() != 1700 {
				t.Fatalf("total_cents = %d, want 1700", req.GetTotalCents())
			}
			if len(req.GetItems()) != 2 {
				t.Fatalf("items = %d, want 2", len(req.GetItems()))
			}
			if req.GetItems()[0].GetName() != "Phone" || req.GetItems()[1].GetName() != "Case" {
				t.Fatalf("item names = %q %q", req.GetItems()[0].GetName(), req.GetItems()[1].GetName())
			}
			return &paymentv1.CreateHostedPaymentSessionResponse{
				OrderId:          "order-1",
				PaymentSessionId: "sess-1",
				ReturnUrl:        req.GetReturnUrl(),
			}, nil
		},
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/cart/checkout", strings.NewReader(checkoutForm(url.Values{"merchant_id": {"merchant-1"}, "checkout_intent_key": {"intent-1"}})))
	req.Host = "localhost:8080"
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: auth.AccessCookieName, Value: signedAccessToken(t, "user-1")})
	req.AddCookie(&http.Cookie{Name: "cart_id", Value: "cart-1"})

	newTestRouter(t, routerDeps{cart: cartSvc, products: productsSvc, orders: ordersSvc, payment: paymentSvc}).ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusSeeOther)
	}
	wantLocation := "http://localhost:8097/pay?callback_url=http%3A%2F%2Flocalhost%3A8080%2Fcallbacks%2Fhosted-payment&order_id=order-1&payment_session_id=sess-1&return_url=http%3A%2F%2Flocalhost%3A8080%2Forders%2Forder-1"
	if got := rec.Header().Get("Location"); got != wantLocation {
		t.Fatalf("location = %q, want %q", got, wantLocation)
	}
	if len(batchIDs) != 2 || batchIDs[0] != "prod-1" || batchIDs[1] != "prod-2" {
		t.Fatalf("batch IDs = %v, want [prod-1 prod-2]", batchIDs)
	}
	if len(removed) != 0 {
		t.Fatalf("removed = %v, want none on checkout", removed)
	}
	if reservedOrderID != "order-1" {
		t.Fatalf("reserved order = %q, want order-1", reservedOrderID)
	}
}

func TestCartCheckoutIntentStableAcrossReloads(t *testing.T) {
	cartSvc := &fakes.CartService{
		GetFn: func(ctx context.Context, cartID string) (*cartv1.Cart, error) {
			return &cartv1.Cart{
				CartId: cartID,
				Items: []*cartv1.CartItem{
					{ProductId: "prod-1", Quantity: 1, MerchantId: "merchant-1", ProductName: "Phone", UnitPriceCents: 1200},
				},
			}, nil
		},
	}
	router := newTestRouter(t, routerDeps{cart: cartSvc})

	first := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodGet, "/cart", nil)
	req1.AddCookie(&http.Cookie{Name: "cart_id", Value: "cart-1"})
	router.ServeHTTP(first, req1)
	if first.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", first.Code, http.StatusOK)
	}
	key1 := checkoutIntentFromHTML(first.Body.String())
	if key1 == "" {
		t.Fatal("missing checkout_intent_key on first cart render")
	}

	second := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/cart", nil)
	req2.AddCookie(&http.Cookie{Name: "cart_id", Value: "cart-1"})
	for _, c := range first.Result().Cookies() {
		req2.AddCookie(c)
	}
	router.ServeHTTP(second, req2)
	if second.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", second.Code, http.StatusOK)
	}
	key2 := checkoutIntentFromHTML(second.Body.String())
	if key2 != key1 {
		t.Fatalf("checkout intent rotated on reload: %q vs %q", key1, key2)
	}
}

func checkoutIntentFromHTML(body string) string {
	const marker = `name="checkout_intent_key" value="`
	i := strings.Index(body, marker)
	if i < 0 {
		return ""
	}
	rest := body[i+len(marker):]
	j := strings.Index(rest, `"`)
	if j < 0 {
		return ""
	}
	return rest[:j]
}

func TestCartCheckoutMarksOrderFailedWhenReserveFails(t *testing.T) {
	cartSvc := &fakes.CartService{
		GetFn: func(ctx context.Context, cartID string) (*cartv1.Cart, error) {
			return &cartv1.Cart{
				CartId: cartID,
				Items: []*cartv1.CartItem{
					{ProductId: "prod-1", Quantity: 1, MerchantId: "merchant-1", ProductName: "Phone", UnitPriceCents: 1200},
				},
			}, nil
		},
	}
	productsSvc := &fakes.ProductsService{
		GetByIDsFn: func(ctx context.Context, ids []string) (*productsv1.GetProductsByIDsResponse, error) {
			return &productsv1.GetProductsByIDsResponse{Products: []*productsv1.Product{
				{Id: "prod-1", Name: "Phone", PriceCents: 1200, MerchantId: "merchant-1"},
			}}, nil
		},
		ReserveFn: func(ctx context.Context, orderID, merchantID string, totalCents int64, items []*productsv1.ReserveStockItem) error {
			return status.Error(codes.FailedPrecondition, "insufficient stock")
		},
	}
	var failedStatus ordersv1.OrderStatus
	ordersSvc := &fakes.OrdersService{
		CreateFn: func(ctx context.Context, buyerUserID, merchantID string, items []*ordersv1.CreateOrderItem, totalCents int64, idempotencyKey string) (*ordersv1.Order, error) {
			return &ordersv1.Order{
				Id:          "order-fail",
				BuyerUserId: buyerUserID,
				MerchantId:  merchantID,
				Status:      ordersv1.OrderStatus_ORDER_STATUS_PENDING,
				TotalCents:  totalCents,
				Items:       []*ordersv1.OrderItem{{ProductId: items[0].GetProductId(), Quantity: items[0].GetQuantity()}},
			}, nil
		},
		UpdateStatusFn: func(ctx context.Context, id string, st ordersv1.OrderStatus) (*ordersv1.Order, error) {
			failedStatus = st
			return &ordersv1.Order{Id: id, Status: st}, nil
		},
	}
	paymentCalled := false
	paymentSvc := &fakes.PaymentService{
		CreateSessionFn: func(ctx context.Context, req *paymentv1.CreateHostedPaymentSessionRequest) (*paymentv1.CreateHostedPaymentSessionResponse, error) {
			paymentCalled = true
			return &paymentv1.CreateHostedPaymentSessionResponse{OrderId: req.GetOrderId()}, nil
		},
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/cart/checkout", strings.NewReader(checkoutForm(url.Values{"merchant_id": {"merchant-1"}, "checkout_intent_key": {"intent-fail"}})))
	req.Host = "localhost:8080"
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: auth.AccessCookieName, Value: signedAccessToken(t, "user-1")})
	req.AddCookie(&http.Cookie{Name: "cart_id", Value: "cart-1"})

	newTestRouter(t, routerDeps{cart: cartSvc, products: productsSvc, orders: ordersSvc, payment: paymentSvc}).ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusConflict)
	}
	if failedStatus != ordersv1.OrderStatus_ORDER_STATUS_FAILED {
		t.Fatalf("order status = %v, want FAILED", failedStatus)
	}
	if paymentCalled {
		t.Fatal("hosted payment session should not be created when reserve fails")
	}
}

func TestCheckoutRejectsMissingShipping(t *testing.T) {
	created := false
	cartSvc := &fakes.CartService{
		GetFn: func(ctx context.Context, cartID string) (*cartv1.Cart, error) {
			return &cartv1.Cart{
				CartId: cartID,
				Items:  []*cartv1.CartItem{{ProductId: "prod-1", Quantity: 1, MerchantId: "merchant-1"}},
			}, nil
		},
	}
	ordersSvc := &fakes.OrdersService{
		CreateFn: func(ctx context.Context, buyerUserID, merchantID string, items []*ordersv1.CreateOrderItem, totalCents int64, idempotencyKey string) (*ordersv1.Order, error) {
			created = true
			return &ordersv1.Order{Id: "order-1"}, nil
		},
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/cart/checkout", strings.NewReader(url.Values{"merchant_id": {"merchant-1"}, "checkout_intent_key": {"intent-1"}}.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: auth.AccessCookieName, Value: signedAccessToken(t, "user-1")})
	req.AddCookie(&http.Cookie{Name: "cart_id", Value: "cart-1"})

	newTestRouter(t, routerDeps{cart: cartSvc, orders: ordersSvc}).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if created {
		t.Fatal("order should not be created without shipping")
	}
}

func checkoutForm(extra url.Values) string {
	v := url.Values{
		"shipping_line1":       {"1 Main St"},
		"shipping_city":        {"New York"},
		"shipping_postal_code": {"10001"},
		"shipping_country":     {"US"},
	}
	for k, vals := range extra {
		v[k] = vals
	}
	return v.Encode()
}
