// Package handlers provides HTTP handlers for the web service. It defines the Handler struct that contains a users client for communicating with the users gRPC service.
package handlers

import (
	"net/http"

	webAuth "refurbished-marketplace/services/web/internal/auth"
	authconfig "refurbished-marketplace/shared/auth/config"
	"refurbished-marketplace/shared/proto/ordersclient"
	"refurbished-marketplace/shared/proto/productsclient"
	"refurbished-marketplace/shared/proto/usersclient"
)

type Handler struct {
	users    *usersclient.Client
	products *productsclient.Client
	orders   *ordersclient.Client
	auth     authconfig.Config
}

func New(users *usersclient.Client, products *productsclient.Client, orders *ordersclient.Client, authCfg authconfig.Config) *Handler {
	return &Handler{users: users, products: products, orders: orders, auth: authCfg}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /healthz", h.handleHealthz)
	mux.HandleFunc("POST /users", h.handleCreateUser)
	mux.HandleFunc("GET /users/{id}", h.handleGetUserByID)

	mux.Handle("POST /products", webAuth.RequireAccessToken(h.auth, http.HandlerFunc(h.handleCreateProduct)))
	mux.Handle("PATCH /products/{id}", webAuth.RequireAccessToken(h.auth, http.HandlerFunc(h.handleUpdateProduct)))
	mux.Handle("DELETE /products/{id}", webAuth.RequireAccessToken(h.auth, http.HandlerFunc(h.handleDeleteProduct)))
	mux.HandleFunc("GET /products", h.handleListProducts)
	mux.HandleFunc("GET /products/{id}", h.handleGetProductByID)

	mux.Handle("POST /orders", webAuth.RequireAccessToken(h.auth, http.HandlerFunc(h.handleCreateOrder)))
	mux.Handle("GET /orders", webAuth.RequireAccessToken(h.auth, http.HandlerFunc(h.handleListOrdersByBuyer)))
	mux.Handle("POST /orders/{id}/confirm", webAuth.RequireAccessToken(h.auth, http.HandlerFunc(h.handleConfirmOrder)))
	mux.Handle("POST /orders/{id}/cancel", webAuth.RequireAccessToken(h.auth, http.HandlerFunc(h.handleCancelOrder)))
	mux.HandleFunc("GET /orders/{id}", h.handleGetOrderByID)

	mux.HandleFunc("POST /auth/login", h.handleLogin)
	mux.HandleFunc("POST /auth/refresh", h.handleRefresh)
	mux.Handle("POST /auth/logout", webAuth.RequireAccessToken(h.auth, http.HandlerFunc(h.handleLogout)))
}
