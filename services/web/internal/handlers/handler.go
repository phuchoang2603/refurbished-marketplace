package handlers

import (
	"net/http"

	authhandlers "refurbished-marketplace/services/web/internal/handlers/auth"
	carthandlers "refurbished-marketplace/services/web/internal/handlers/cart"
	orderhandlers "refurbished-marketplace/services/web/internal/handlers/orders"
	paymenthandlers "refurbished-marketplace/services/web/internal/handlers/payment"
	producthandlers "refurbished-marketplace/services/web/internal/handlers/products"
	shared "refurbished-marketplace/services/web/internal/handlers/shared"

	webAuth "refurbished-marketplace/services/web/internal/auth"
	sharedviews "refurbished-marketplace/services/web/internal/views/shared"
	authconfig "refurbished-marketplace/shared/auth/config"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	auth     *authhandlers.Handler
	products *producthandlers.Handler
	cart     *carthandlers.Handler
	orders   *orderhandlers.Handler
	payment  *paymenthandlers.Handler
	authCfg  authconfig.Config
}

func New(
	users shared.UsersService,
	products shared.ProductsService,
	orders shared.OrdersService,
	cart shared.CartService,
	payment shared.PaymentService,
	authCfg authconfig.Config,
) *Handler {
	deps := &shared.Dependencies{
		Users:    users,
		Products: products,
		Orders:   orders,
		Cart:     cart,
		Payment:  payment,
	}
	return &Handler{
		auth:     authhandlers.New(deps),
		products: producthandlers.New(deps),
		cart:     carthandlers.New(deps),
		orders:   orderhandlers.New(deps),
		payment:  paymenthandlers.New(deps),
		authCfg:  authCfg,
	}
}

func (h *Handler) requireAccessToken(resumePath string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		unauthorized := func(w http.ResponseWriter, r *http.Request) {
			shared.RedirectBrowserToLogin(w, r, resumePath)
		}

		return webAuth.RequireAccessToken(
			h.authCfg,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				state := sharedviews.AuthState{Authenticated: true}
				next.ServeHTTP(w, r.WithContext(sharedviews.WithAuthState(r.Context(), state)))
			}),
			unauthorized,
		)
	}
}

func (h *Handler) viewAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		state := sharedviews.AuthState{Authenticated: webAuth.HasValidAccessToken(h.authCfg, r)}
		next.ServeHTTP(w, r.WithContext(sharedviews.WithAuthState(r.Context(), state)))
	})
}

func (h *Handler) Register(router chi.Router) {
	router.Group(func(r chi.Router) {
		r.Use(h.viewAuth)
		h.auth.RegisterPages(r)
		h.auth.RegisterActions(r)
		h.products.RegisterPages(r)
		h.cart.RegisterPages(r)
		h.cart.RegisterActions(r)
	})

	router.With(h.requireAccessToken("/cart")).Group(h.cart.RegisterProtectedActions)
	router.With(h.requireAccessToken("/products")).Group(func(r chi.Router) {
		h.orders.RegisterPages(r)
		h.orders.RegisterActions(r)
	})
	router.Group(func(r chi.Router) {
		h.registerStatusRoutes(r)
		h.payment.RegisterActions(r)
	})
}
