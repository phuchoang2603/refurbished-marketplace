package tests

import (
	"net/http"
	"testing"
	"time"

	"refurbished-marketplace/services/web/internal/handlers"
	shared "refurbished-marketplace/services/web/internal/handlers/shared"
	"refurbished-marketplace/services/web/tests/fakes"
	authconfig "refurbished-marketplace/shared/auth/config"

	"github.com/go-chi/chi/v5"
	jwtlib "github.com/golang-jwt/jwt/v5"
)

const testJWTSecret = "secret"

type routerDeps struct {
	users         *fakes.UsersService
	products      *fakes.ProductsService
	orders        *fakes.OrdersService
	cart          *fakes.CartService
	payment       *fakes.PaymentService
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
