package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"refurbished-marketplace/services/web/internal/auth"
	"refurbished-marketplace/services/web/internal/handlers"
	authconfig "refurbished-marketplace/shared/auth/config"
	cartv1 "refurbished-marketplace/shared/proto/cart/v1"
	ordersv1 "refurbished-marketplace/shared/proto/orders/v1"
	paymentv1 "refurbished-marketplace/shared/proto/payment/v1"
	productsv1 "refurbished-marketplace/shared/proto/products/v1"
	usersv1 "refurbished-marketplace/shared/proto/users/v1"

	"github.com/go-chi/chi/v5"
	jwtlib "github.com/golang-jwt/jwt/v5"
)

const testJWTSecret = "secret"

type fakeUsersService struct {
	loginFn      func(context.Context, string, string) (*usersv1.TokenResponse, error)
	logoutFn     func(context.Context, string) (*usersv1.LogoutResponse, error)
	createUserFn func(context.Context, string, string) (*usersv1.User, error)
}

func (f *fakeUsersService) Login(ctx context.Context, email, password string) (*usersv1.TokenResponse, error) {
	if f.loginFn != nil {
		return f.loginFn(ctx, email, password)
	}
	return nil, nil
}

func (f *fakeUsersService) Logout(ctx context.Context, refreshToken string) (*usersv1.LogoutResponse, error) {
	if f.logoutFn != nil {
		return f.logoutFn(ctx, refreshToken)
	}
	return &usersv1.LogoutResponse{}, nil
}

func (f *fakeUsersService) CreateUser(ctx context.Context, email, password string) (*usersv1.User, error) {
	if f.createUserFn != nil {
		return f.createUserFn(ctx, email, password)
	}
	return &usersv1.User{}, nil
}

type fakeProductsService struct {
	getByIDFn func(context.Context, string) (*productsv1.Product, error)
	listFn    func(context.Context, int32, int32) (*productsv1.ListProductsResponse, error)
}

func (f *fakeProductsService) GetProductByID(ctx context.Context, id string) (*productsv1.Product, error) {
	if f.getByIDFn != nil {
		return f.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (f *fakeProductsService) ListProducts(ctx context.Context, limit, offset int32) (*productsv1.ListProductsResponse, error) {
	if f.listFn != nil {
		return f.listFn(ctx, limit, offset)
	}
	return &productsv1.ListProductsResponse{}, nil
}

type fakeOrdersService struct {
	createFn func(context.Context, string, string, []*ordersv1.CreateOrderItem, int64) (*ordersv1.Order, error)
	getFn    func(context.Context, string) (*ordersv1.Order, error)
	listFn   func(context.Context, string, int32, int32) (*ordersv1.ListOrdersByBuyerResponse, error)
}

func (f *fakeOrdersService) CreateOrder(ctx context.Context, buyerUserID, merchantID string, items []*ordersv1.CreateOrderItem, totalCents int64) (*ordersv1.Order, error) {
	if f.createFn != nil {
		return f.createFn(ctx, buyerUserID, merchantID, items, totalCents)
	}
	return nil, nil
}

func (f *fakeOrdersService) GetOrderByID(ctx context.Context, id string) (*ordersv1.Order, error) {
	if f.getFn != nil {
		return f.getFn(ctx, id)
	}
	return nil, nil
}

func (f *fakeOrdersService) ListOrdersByBuyer(ctx context.Context, buyerUserID string, limit, offset int32) (*ordersv1.ListOrdersByBuyerResponse, error) {
	if f.listFn != nil {
		return f.listFn(ctx, buyerUserID, limit, offset)
	}
	return &ordersv1.ListOrdersByBuyerResponse{}, nil
}

type fakeCartService struct {
	getFn       func(context.Context, string) (*cartv1.Cart, error)
	addFn       func(context.Context, string, string, int32) (*cartv1.Cart, error)
	setQtyFn    func(context.Context, string, string, int32) (*cartv1.Cart, error)
	removeFn    func(context.Context, string, string) (*cartv1.Cart, error)
	clearCartFn func(context.Context, string) error
}

func (f *fakeCartService) GetCart(ctx context.Context, cartID string) (*cartv1.Cart, error) {
	if f.getFn != nil {
		return f.getFn(ctx, cartID)
	}
	return &cartv1.Cart{}, nil
}

func (f *fakeCartService) AddCartItem(ctx context.Context, cartID, productID string, quantity int32) (*cartv1.Cart, error) {
	if f.addFn != nil {
		return f.addFn(ctx, cartID, productID, quantity)
	}
	return &cartv1.Cart{}, nil
}

func (f *fakeCartService) SetCartItemQuantity(ctx context.Context, cartID, productID string, quantity int32) (*cartv1.Cart, error) {
	if f.setQtyFn != nil {
		return f.setQtyFn(ctx, cartID, productID, quantity)
	}
	return &cartv1.Cart{}, nil
}

func (f *fakeCartService) RemoveCartItem(ctx context.Context, cartID, productID string) (*cartv1.Cart, error) {
	if f.removeFn != nil {
		return f.removeFn(ctx, cartID, productID)
	}
	return &cartv1.Cart{}, nil
}

func (f *fakeCartService) ClearCart(ctx context.Context, cartID string) error {
	if f.clearCartFn != nil {
		return f.clearCartFn(ctx, cartID)
	}
	return nil
}

type fakePaymentService struct {
	handleWebhookFn func(context.Context, *paymentv1.HandleGatewayWebhookRequest) (*paymentv1.HandleGatewayWebhookResponse, error)
}

func (f *fakePaymentService) HandleGatewayWebhook(ctx context.Context, req *paymentv1.HandleGatewayWebhookRequest) (*paymentv1.HandleGatewayWebhookResponse, error) {
	if f.handleWebhookFn != nil {
		return f.handleWebhookFn(ctx, req)
	}
	return &paymentv1.HandleGatewayWebhookResponse{}, nil
}

type routerDeps struct {
	users    *fakeUsersService
	products *fakeProductsService
	orders   *fakeOrdersService
	cart     *fakeCartService
	payment  *fakePaymentService
}

func newTestRouter(t *testing.T, deps routerDeps) http.Handler {
	t.Helper()
	if deps.users == nil {
		deps.users = &fakeUsersService{}
	}
	if deps.products == nil {
		deps.products = &fakeProductsService{}
	}
	if deps.orders == nil {
		deps.orders = &fakeOrdersService{}
	}
	if deps.cart == nil {
		deps.cart = &fakeCartService{}
	}
	if deps.payment == nil {
		deps.payment = &fakePaymentService{}
	}

	h := handlers.New(deps.users, deps.products, deps.orders, deps.cart, deps.payment, authconfig.DefaultConfig(testJWTSecret))
	router := chi.NewRouter()
	h.Register(router)
	return router
}

func signedAccessToken(t *testing.T, subject string) string {
	t.Helper()
	claims := jwtlib.MapClaims{
		"typ": "access",
		"iss": authconfig.DefaultJWTIssuer,
		"aud": authconfig.DefaultJWTAudience,
		"sub": subject,
		"jti": "test-token",
		"exp": time.Now().Add(time.Hour).Unix(),
	}
	token, err := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims).SignedString([]byte(testJWTSecret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return token
}

func TestProtectedGetRedirectsToLoginWithNext(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/orders", nil)

	newTestRouter(t, routerDeps{}).ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusSeeOther)
	}
	if got := rec.Header().Get("Location"); got != "/auth/login?next=%2Forders" {
		t.Fatalf("location = %q, want /auth/login?next=%%2Forders", got)
	}
}

func TestProtectedPostRedirectsToLoginWithSafeResume(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/cart/checkout", nil)

	newTestRouter(t, routerDeps{}).ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusSeeOther)
	}
	if got := rec.Header().Get("Location"); got != "/auth/login?next=%2Fcart" {
		t.Fatalf("location = %q, want /auth/login?next=%%2Fcart", got)
	}
}

func TestAuthenticatedProtectedRouteProceeds(t *testing.T) {
	ordersSvc := &fakeOrdersService{
		listFn: func(ctx context.Context, buyerUserID string, limit, offset int32) (*ordersv1.ListOrdersByBuyerResponse, error) {
			if buyerUserID != "user-1" {
				t.Fatalf("buyerUserID = %q, want user-1", buyerUserID)
			}
			return &ordersv1.ListOrdersByBuyerResponse{}, nil
		},
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/orders", nil)
	req.AddCookie(&http.Cookie{Name: auth.AccessCookieName, Value: signedAccessToken(t, "user-1")})

	newTestRouter(t, routerDeps{orders: ordersSvc}).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}
