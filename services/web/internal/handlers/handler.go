// Package handlers provides HTTP handlers for the web service. It defines the Handler struct that contains a users client for communicating with the users gRPC service.
package handlers

import (
	"net/http"

	"refurbished-marketplace/shared/proto/productsclient"
	"refurbished-marketplace/shared/proto/usersclient"
)

type Handler struct {
	users    *usersclient.Client
	products *productsclient.Client
}

func New(users *usersclient.Client, products *productsclient.Client) *Handler {
	return &Handler{users: users, products: products}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /healthz", h.handleHealthz)
	mux.HandleFunc("POST /users", h.handleCreateUser)
	mux.HandleFunc("GET /users/{id}", h.handleGetUserByID)
	mux.HandleFunc("POST /products", h.handleCreateProduct)
	mux.HandleFunc("GET /products", h.handleListProducts)
	mux.HandleFunc("GET /products/{id}", h.handleGetProductByID)
	mux.HandleFunc("POST /auth/login", h.handleLogin)
	mux.HandleFunc("POST /auth/refresh", h.handleRefresh)
	mux.HandleFunc("POST /auth/logout", h.handleLogout)
}
