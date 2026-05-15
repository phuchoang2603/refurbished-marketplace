package handlers

import (
	"net/http"

	"refurbished-marketplace/services/web/internal/views"
	cartproto "refurbished-marketplace/shared/proto/cart"
	ordersproto "refurbished-marketplace/shared/proto/orders"
	paymentproto "refurbished-marketplace/shared/proto/payment"
	productsproto "refurbished-marketplace/shared/proto/products"
	usersproto "refurbished-marketplace/shared/proto/users"

	webAuth "refurbished-marketplace/services/web/internal/auth"

	authconfig "refurbished-marketplace/shared/auth/config"
)

const staticDir = "/static"

type Handler struct {
	users    *usersproto.Client
	products *productsproto.Client
	orders   *ordersproto.Client
	cart     *cartproto.Client
	payment  *paymentproto.Client
	auth     authconfig.Config
}

func New(users *usersproto.Client, products *productsproto.Client, orders *ordersproto.Client, cart *cartproto.Client, payment *paymentproto.Client, authCfg authconfig.Config) *Handler {
	return &Handler{
		users:    users,
		products: products,
		orders:   orders,
		cart:     cart,
		payment:  payment,
		auth:     authCfg,
	}
}

func (h *Handler) requireAccessToken(next http.Handler) http.Handler {
	return webAuth.RequireAccessToken(h.auth, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		state := views.AuthState{Authenticated: true}
		next.ServeHTTP(w, r.WithContext(views.WithAuthState(r.Context(), state)))
	}), writeUnauthorized)
}

func (h *Handler) withViewAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state := views.AuthState{Authenticated: webAuth.HasValidAccessToken(h.auth, r)}
		next(w, r.WithContext(views.WithAuthState(r.Context(), state)))
	}
}

func (h *Handler) Register(mux *http.ServeMux) {
	view := h.withViewAuth
	mux.HandleFunc("GET /healthz", h.handleHealthz)
	mux.HandleFunc("GET /", view(h.handleHome))
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))

	mux.HandleFunc("POST /users", view(h.handleCreateUser))
	mux.HandleFunc("GET /users/{id}", view(h.handleGetUserByID))

	mux.HandleFunc("GET /auth/login", view(h.handleLoginPage))
	mux.HandleFunc("POST /auth/login", view(h.handleLogin))
	mux.HandleFunc("GET /auth/register", view(h.handleRegisterPage))
	mux.HandleFunc("POST /auth/logout", view(h.handleLogout))

	mux.HandleFunc("GET /products", view(h.handleListProducts))
	mux.HandleFunc("GET /products/{id}", view(h.handleGetProductByID))

	mux.HandleFunc("POST /cart/items", view(h.handleAddCartItem))
	mux.HandleFunc("PATCH /cart/items/{product_id}", view(h.handleSetCartItemQuantity))
	mux.HandleFunc("DELETE /cart/items/{product_id}", view(h.handleRemoveCartItem))
	mux.Handle("POST /cart/checkout", h.requireAccessToken(http.HandlerFunc(h.handleCheckoutCart)))
	mux.HandleFunc("GET /cart", view(h.handleGetCart))

	mux.Handle("POST /orders", h.requireAccessToken(http.HandlerFunc(h.handleCreateOrder)))
	mux.Handle("GET /orders", h.requireAccessToken(http.HandlerFunc(h.handleListOrdersByBuyer)))
	mux.Handle("GET /orders/{id}", h.requireAccessToken(http.HandlerFunc(h.handleGetOrderByID)))

	mux.HandleFunc("POST /webhooks/stripe-simulator", h.handleStripeSimWebhook)
}

func (h *Handler) handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) handleHome(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/products", http.StatusSeeOther)
}
