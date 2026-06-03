package tests

import (
	"context"
	"net/http"
	"testing"
	"time"

	"refurbished-marketplace/services/web/internal/handlers"
	shared "refurbished-marketplace/services/web/internal/handlers/shared"
	cartv1 "refurbished-marketplace/shared/proto/cart/v1"
	ordersv1 "refurbished-marketplace/shared/proto/orders/v1"
	paymentv1 "refurbished-marketplace/shared/proto/payment/v1"
	productsv1 "refurbished-marketplace/shared/proto/products/v1"
	usersv1 "refurbished-marketplace/shared/proto/users/v1"

	authconfig "refurbished-marketplace/shared/auth/config"

	jwtlib "github.com/golang-jwt/jwt/v5"

	"github.com/go-chi/chi/v5"
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
	createFn  func(context.Context, string, string, int64, string, int32) (*productsv1.Product, error)
	getByIDFn func(context.Context, string) (*productsv1.Product, error)
	listFn    func(context.Context, int32, int32) (*productsv1.ListProductsResponse, error)
}

func (f *fakeProductsService) CreateProduct(ctx context.Context, name, description string, priceCents int64, merchantID string, initialStock int32) (*productsv1.Product, error) {
	if f.createFn != nil {
		return f.createFn(ctx, name, description, priceCents, merchantID, initialStock)
	}
	return &productsv1.Product{}, nil
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
	addFn       func(context.Context, string, string, string, int32) (*cartv1.Cart, error)
	setQtyFn    func(context.Context, string, string, string, int32) (*cartv1.Cart, error)
	removeFn    func(context.Context, string, string) (*cartv1.Cart, error)
	clearCartFn func(context.Context, string) error
}

func (f *fakeCartService) GetCart(ctx context.Context, cartID string) (*cartv1.Cart, error) {
	if f.getFn != nil {
		return f.getFn(ctx, cartID)
	}
	return &cartv1.Cart{}, nil
}

func (f *fakeCartService) AddCartItem(ctx context.Context, cartID, productID, merchantID string, quantity int32) (*cartv1.Cart, error) {
	if f.addFn != nil {
		return f.addFn(ctx, cartID, productID, merchantID, quantity)
	}
	return &cartv1.Cart{}, nil
}

func (f *fakeCartService) SetCartItemQuantity(ctx context.Context, cartID, productID, merchantID string, quantity int32) (*cartv1.Cart, error) {
	if f.setQtyFn != nil {
		return f.setQtyFn(ctx, cartID, productID, merchantID, quantity)
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
	createSessionFn func(context.Context, *paymentv1.CreateHostedPaymentSessionRequest) (*paymentv1.CreateHostedPaymentSessionResponse, error)
	getSessionFn    func(context.Context, string) (*paymentv1.HostedPaymentSession, error)
	handleWebhookFn func(context.Context, *paymentv1.HandleGatewayWebhookRequest) (*paymentv1.HandleGatewayWebhookResponse, error)
}

func (f *fakePaymentService) CreateHostedPaymentSession(ctx context.Context, req *paymentv1.CreateHostedPaymentSessionRequest) (*paymentv1.CreateHostedPaymentSessionResponse, error) {
	if f.createSessionFn != nil {
		return f.createSessionFn(ctx, req)
	}
	return &paymentv1.CreateHostedPaymentSessionResponse{}, nil
}

func (f *fakePaymentService) GetHostedPaymentSessionByOrder(ctx context.Context, orderID string) (*paymentv1.HostedPaymentSession, error) {
	if f.getSessionFn != nil {
		return f.getSessionFn(ctx, orderID)
	}
	return nil, nil
}

func (f *fakePaymentService) HandleGatewayWebhook(ctx context.Context, req *paymentv1.HandleGatewayWebhookRequest) (*paymentv1.HandleGatewayWebhookResponse, error) {
	if f.handleWebhookFn != nil {
		return f.handleWebhookFn(ctx, req)
	}
	return &paymentv1.HandleGatewayWebhookResponse{}, nil
}

type routerDeps struct {
	users         *fakeUsersService
	products      *fakeProductsService
	orders        *fakeOrdersService
	cart          *fakeCartService
	payment       *fakePaymentService
	hostedPayment shared.HostedPaymentConfig
}

var defaultTestHostedPayment = shared.HostedPaymentConfig{
	GatewayBaseURL: "http://localhost:8097",
}

func newTestRouter(t *testing.T, deps routerDeps) http.Handler {
	t.Helper()
	hostedPayment := deps.hostedPayment
	if hostedPayment.GatewayBaseURL == "" {
		hostedPayment = defaultTestHostedPayment
	}
	h := handlers.New(deps.users, deps.products, deps.orders, deps.cart, deps.payment, hostedPayment, authconfig.DefaultConfig(testJWTSecret))
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
