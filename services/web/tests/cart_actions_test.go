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
	checkoutv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/checkout/v1"
	inventoryv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/inventory/v1"
	ordersv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/orders/v1"
	productsv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/products/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestAddCartItemRedirectsToCart(t *testing.T) {
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
	form := url.Values{"product_id": {"prod-1"}, "merchant_id": {"merchant-1"}, "quantity": {"2"}, "product_name": {"Phone"}, "unit_price_cents": {"1200"}}
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

func TestCheckoutSubmitsSnapshotAndRedirectsToProgress(t *testing.T) {
	var removed []string
	var batchIDs []string
	batchCalls := 0
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
			batchCalls++
			batchIDs = append([]string{}, ids...)
			return &productsv1.GetProductsByIDsResponse{Products: []*productsv1.Product{
				{Id: "prod-1", Name: "Phone", Category: "phone", Condition: "excellent", PriceCents: 1200, MerchantId: "merchant-1"},
				{Id: "prod-2", Name: "Case", Category: "accessory", Condition: "fair", PriceCents: 250, MerchantId: "merchant-1"},
			}}, nil
		},
	}
	ordersSvc := &fakes.OrdersService{
		CreateFn: func(ctx context.Context, buyerUserID, merchantID string, items []*ordersv1.CreateOrderItem, totalCents int64, idempotencyKey string) (*ordersv1.Order, error) {
			if idempotencyKey != "intent-1" {
				t.Fatalf("idempotencyKey = %q, want intent-1", idempotencyKey)
			}
			t.Fatal("checkout POST must not call Orders directly")
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
	inventorySvc := &fakes.InventoryService{
		ReserveFn: func(ctx context.Context, orderID, merchantID string, totalCents int64, items []*inventoryv1.ReserveStockItem) error {
			reservedOrderID = orderID
			if merchantID != "merchant-1" {
				t.Fatalf("reserve merchantID = %q, want merchant-1", merchantID)
			}
			if len(items) != 2 {
				t.Fatalf("reserve items = %d, want 2", len(items))
			}
			return nil
		},
	}
	rec := httptest.NewRecorder()
	checkoutSvc := &fakes.CheckoutService{SubmitFn: func(ctx context.Context, request *checkoutv1.SubmitCheckoutRequest) (*checkoutv1.SubmitCheckoutResponse, error) {
		if request.GetBuyerUserId() != "user-1" || request.GetIntentKey() != "intent-1" || request.GetMerchantId() != "merchant-1" {
			t.Fatalf("wrong checkout identity: %v", request)
		}
		if request.GetTotalCents() != 1700 || len(request.GetItems()) != 2 || request.GetItems()[0].GetName() != "Phone" || request.GetItems()[1].GetName() != "Case" {
			t.Fatalf("checkout did not use authoritative prices and names: %v", request)
		}
		if request.GetItems()[0].GetCategory() != "phone" || request.GetItems()[0].GetCondition() != "excellent" || request.GetItems()[1].GetCategory() != "accessory" || request.GetItems()[1].GetCondition() != "fair" {
			t.Fatalf("wrong grading: %v", request)
		}
		if request.GetShippingAddress().GetCountry() != "US" {
			t.Fatalf("missing buyer and shipping context: %v", request)
		}
		return &checkoutv1.SubmitCheckoutResponse{CheckoutId: "checkout-1", State: checkoutv1.CheckoutState_CHECKOUT_STATE_CREATING_ORDER}, nil
	}}
	req := httptest.NewRequest(http.MethodPost, "/cart/checkout", strings.NewReader(checkoutForm(url.Values{"merchant_id": {"merchant-1"}, "checkout_intent_key": {"intent-1"}})))
	req.Host = "localhost:8080"
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: auth.AccessCookieName, Value: signedAccessToken(t, "user-1")})
	req.AddCookie(&http.Cookie{Name: "cart_id", Value: "cart-1"})

	newTestRouter(t, routerDeps{cart: cartSvc, products: productsSvc, inventory: inventorySvc, orders: ordersSvc, checkout: checkoutSvc}).ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusSeeOther)
	}
	wantLocation := "/checkouts/checkout-1"
	if got := rec.Header().Get("Location"); got != wantLocation {
		t.Fatalf("location = %q, want %q", got, wantLocation)
	}
	if batchCalls != 1 || len(batchIDs) != 2 || batchIDs[0] != "prod-1" || batchIDs[1] != "prod-2" {
		t.Fatalf("batch IDs = %v, want [prod-1 prod-2]", batchIDs)
	}
	if len(removed) != 0 {
		t.Fatalf("removed = %v, want none on checkout", removed)
	}
	if reservedOrderID != "" {
		t.Fatalf("web reserved stock directly for %q", reservedOrderID)
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

func TestCartCheckoutDoesNotReserveBeforeAsynchronousResult(t *testing.T) {
	for _, reason := range []string{"insufficient stock", "stock not initialized"} {
		t.Run(reason, func(t *testing.T) {
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
			}
			inventorySvc := &fakes.InventoryService{
				ReserveFn: func(ctx context.Context, orderID, merchantID string, totalCents int64, items []*inventoryv1.ReserveStockItem) error {
					return status.Error(codes.FailedPrecondition, reason)
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
			checkoutSvc := &fakes.CheckoutService{SubmitFn: func(ctx context.Context, request *checkoutv1.SubmitCheckoutRequest) (*checkoutv1.SubmitCheckoutResponse, error) {
				return &checkoutv1.SubmitCheckoutResponse{CheckoutId: "checkout-pending"}, nil
			}}
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/cart/checkout", strings.NewReader(checkoutForm(url.Values{"merchant_id": {"merchant-1"}, "checkout_intent_key": {"intent-fail"}})))
			req.Host = "localhost:8080"
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.AddCookie(&http.Cookie{Name: auth.AccessCookieName, Value: signedAccessToken(t, "user-1")})
			req.AddCookie(&http.Cookie{Name: "cart_id", Value: "cart-1"})

			newTestRouter(t, routerDeps{cart: cartSvc, products: productsSvc, inventory: inventorySvc, orders: ordersSvc, checkout: checkoutSvc}).ServeHTTP(rec, req)

			if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != "/checkouts/checkout-pending" {
				t.Fatalf("checkout should be pending, got %d %s", rec.Code, rec.Header().Get("Location"))
			}
			if failedStatus != ordersv1.OrderStatus_ORDER_STATUS_UNSPECIFIED {
				t.Fatalf("web finalized an order directly: %v", failedStatus)
			}
		})
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
		"shipping_name":        {"Test Buyer"},
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

func TestCheckoutRevalidationFailsClosed(t *testing.T) {
	for _, tc := range []struct {
		name     string
		products []*productsv1.Product
		err      error
		want     int
	}{
		{"missing", nil, nil, 409},
		{"wrong merchant", []*productsv1.Product{{Id: "prod-1", MerchantId: "other-merchant", PriceCents: 1000}}, nil, 409},
		{"products unavailable", nil, status.Error(codes.Unavailable, "products unavailable"), 503},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cart := &fakes.CartService{GetFn: func(context.Context, string) (*cartv1.Cart, error) {
				return &cartv1.Cart{Items: []*cartv1.CartItem{{ProductId: "prod-1", MerchantId: "merchant-1", Quantity: 1}}}, nil
			}}
			products := &fakes.ProductsService{GetByIDFn: func(context.Context, string) (*productsv1.Product, error) {
				t.Fatal("per-item lookup")
				return nil, nil
			}, GetByIDsFn: func(context.Context, []string) (*productsv1.GetProductsByIDsResponse, error) {
				return &productsv1.GetProductsByIDsResponse{Products: tc.products}, tc.err
			}}
			checkout := &fakes.CheckoutService{SubmitFn: func(context.Context, *checkoutv1.SubmitCheckoutRequest) (*checkoutv1.SubmitCheckoutResponse, error) {
				t.Fatal("submitted invalid snapshot")
				return nil, nil
			}}
			r := httptest.NewRequest(http.MethodPost, "/cart/checkout", strings.NewReader(checkoutForm(url.Values{"merchant_id": {"merchant-1"}, "checkout_intent_key": {"intent-1"}})))
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			r.AddCookie(&http.Cookie{Name: auth.AccessCookieName, Value: signedAccessToken(t, "user-1")})
			r.AddCookie(&http.Cookie{Name: "cart_id", Value: "cart-1"})
			w := httptest.NewRecorder()
			newTestRouter(t, routerDeps{cart: cart, products: products, checkout: checkout}).ServeHTTP(w, r)
			if w.Code != tc.want {
				t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
			}
		})
	}
}

func TestCheckoutUnavailableAndShippingName(t *testing.T) {
	for _, tc := range []struct {
		name, shippingName, accept string
		want                       int
		calls                      int
	}{
		{"missing name", "", "", 400, 0},
		{"unavailable", "Buyer", "", 503, 1},
		{"unavailable popup", "Buyer", "text/event-stream", 200, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			cart := &fakes.CartService{GetFn: func(context.Context, string) (*cartv1.Cart, error) {
				return &cartv1.Cart{Items: []*cartv1.CartItem{{ProductId: "prod-1", MerchantId: "merchant-1", Quantity: 1}}}, nil
			}}
			products := &fakes.ProductsService{GetByIDsFn: func(context.Context, []string) (*productsv1.GetProductsByIDsResponse, error) {
				return &productsv1.GetProductsByIDsResponse{Products: []*productsv1.Product{{Id: "prod-1", MerchantId: "merchant-1", PriceCents: 1000, Category: "phone", Condition: "good"}}}, nil
			}}
			checkout := &fakes.CheckoutService{SubmitFn: func(context.Context, *checkoutv1.SubmitCheckoutRequest) (*checkoutv1.SubmitCheckoutResponse, error) {
				calls++
				return nil, status.Error(codes.Unavailable, "private configuration detail")
			}}
			r := httptest.NewRequest(http.MethodPost, "/cart/checkout", strings.NewReader(checkoutForm(url.Values{"merchant_id": {"merchant-1"}, "checkout_intent_key": {"intent-1"}, "shipping_name": {tc.shippingName}})))
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			r.Header.Set("Accept", tc.accept)
			r.AddCookie(&http.Cookie{Name: auth.AccessCookieName, Value: signedAccessToken(t, "user-1")})
			r.AddCookie(&http.Cookie{Name: "cart_id", Value: "cart-1"})
			w := httptest.NewRecorder()
			newTestRouter(t, routerDeps{cart: cart, products: products, checkout: checkout}).ServeHTTP(w, r)
			if w.Code != tc.want || calls != tc.calls {
				t.Fatalf("status=%d calls=%d body=%s", w.Code, calls, w.Body.String())
			}
			if tc.calls > 0 && (!strings.Contains(w.Body.String(), "Checkout unavailable") || strings.Contains(w.Body.String(), "private configuration detail")) {
				t.Fatalf("unfriendly error: %s", w.Body.String())
			}
		})
	}
}
