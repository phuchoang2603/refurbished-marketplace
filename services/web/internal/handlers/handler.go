// Package handlers provides HTTP handlers for the web service. It defines the Handler struct that contains a users client for communicating with the users gRPC service.
package handlers

import (
	"net/http"

	webAuth "refurbished-marketplace/services/web/internal/auth"
	authconfig "refurbished-marketplace/shared/auth/config"
	"refurbished-marketplace/shared/proto/cartclient"
	"refurbished-marketplace/shared/proto/ordersclient"
	"refurbished-marketplace/shared/proto/productsclient"
	"refurbished-marketplace/shared/proto/usersclient"
)

type Handler struct {
	users    *usersclient.Client
	products *productsclient.Client
	orders   *ordersclient.Client
	cart     *cartclient.Client
	auth     authconfig.Config
}

func New(users *usersclient.Client, products *productsclient.Client, orders *ordersclient.Client, cart *cartclient.Client, authCfg authconfig.Config) *Handler {
	return &Handler{users: users, products: products, orders: orders, cart: cart, auth: authCfg}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /healthz", h.handleHealthz)
	mux.HandleFunc("POST /users", h.handleCreateUser)
	mux.HandleFunc("GET /users/{id}", h.handleGetUserByID)

	mux.HandleFunc("GET /products", h.handleListProducts)
	mux.HandleFunc("GET /products/{id}", h.handleGetProductByID)

	mux.HandleFunc("POST /cart/items", h.handleAddCartItem)
	mux.HandleFunc("PATCH /cart/items/{product_id}", h.handleSetCartItemQuantity)
	mux.HandleFunc("DELETE /cart/items/{product_id}", h.handleRemoveCartItem)
	mux.Handle("POST /cart/checkout", webAuth.RequireAccessToken(h.auth, http.HandlerFunc(h.handleCheckoutCart)))
	mux.HandleFunc("GET /cart", h.handleGetCart)

	mux.Handle("POST /orders", webAuth.RequireAccessToken(h.auth, http.HandlerFunc(h.handleCreateOrder)))
	mux.Handle("GET /orders", webAuth.RequireAccessToken(h.auth, http.HandlerFunc(h.handleListOrdersByBuyer)))
	mux.HandleFunc("GET /orders/{id}", h.handleGetOrderByID)

	mux.HandleFunc("POST /auth/login", h.handleLogin)
	mux.HandleFunc("POST /auth/refresh", h.handleRefresh)
	mux.Handle("POST /auth/logout", webAuth.RequireAccessToken(h.auth, http.HandlerFunc(h.handleLogout)))
}

func (h *Handler) handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
