package handlers

import (
	authhandlers "refurbished-marketplace/services/web/internal/handlers/auth"
	carthandlers "refurbished-marketplace/services/web/internal/handlers/cart"
	orderhandlers "refurbished-marketplace/services/web/internal/handlers/orders"
	producthandlers "refurbished-marketplace/services/web/internal/handlers/products"
	shared "refurbished-marketplace/services/web/internal/handlers/shared"

	authconfig "refurbished-marketplace/shared/auth/config"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	deps     *shared.Dependencies
	auth     *authhandlers.Handler
	products *producthandlers.Handler
	cart     *carthandlers.Handler
	orders   *orderhandlers.Handler
	authCfg  authconfig.Config
}

func New(
	users shared.UsersService,
	products shared.ProductsService,
	orders shared.OrdersService,
	cart shared.CartService,
	payment shared.PaymentService,
	hostedPayment shared.HostedPaymentConfig,
	authCfg authconfig.Config,
) *Handler {
	deps := &shared.Dependencies{
		Users:         users,
		Products:      products,
		Orders:        orders,
		Cart:          cart,
		Payment:       payment,
		HostedPayment: hostedPayment,
	}
	return &Handler{
		deps:     deps,
		auth:     authhandlers.New(deps),
		products: producthandlers.New(deps),
		cart:     carthandlers.New(deps),
		orders:   orderhandlers.New(deps),
		authCfg:  authCfg,
	}
}

func (h *Handler) Register(router chi.Router) {
	router.Group(func(r chi.Router) {
		r.Use(h.viewAuth)
		h.auth.RegisterPages(r)
		h.auth.RegisterActions(r)
		h.products.RegisterPages(r)
		h.cart.RegisterPages(r)
		h.cart.RegisterActions(r)

		r.Group(func(r chi.Router) {
			r.Use(h.requireAccessToken())
			h.products.RegisterProtectedPages(r)
			h.products.RegisterProtectedActions(r)
			h.cart.RegisterProtectedActions(r)
			h.orders.RegisterPages(r)
			h.orders.RegisterActions(r)
		})
	})

	router.Group(func(r chi.Router) {
		h.registerStatusRoutes(r)
		h.registerCallbackRoutes(r)
	})
}
