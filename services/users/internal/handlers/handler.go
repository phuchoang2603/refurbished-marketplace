// Package handlers provides HTTP handlers for the users service.
package handlers

import (
	"net/http"

	"refurbished-marketplace/services/users/internal/service"
)

type Handler struct {
	svc *service.Service
}

type createUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func New(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /healthz", h.handleHealthz)
	mux.HandleFunc("POST /users", h.handleCreateUser)
	mux.HandleFunc("POST /auth/login", h.handleLogin)
	mux.HandleFunc("POST /auth/refresh", h.handleRefresh)
	mux.HandleFunc("POST /auth/logout", h.handleLogout)
	mux.HandleFunc("GET /users/{id}", h.handleGetUserByID)
}
