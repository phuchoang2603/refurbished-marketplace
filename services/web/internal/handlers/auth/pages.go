package auth

import (
	"net/http"

	shared "refurbished-marketplace/services/web/internal/handlers/shared"
	authviews "refurbished-marketplace/services/web/internal/views/auth"

	"github.com/go-chi/chi/v5"
)

type Handler struct{ deps *shared.Dependencies }

func New(deps *shared.Dependencies) *Handler { return &Handler{deps: deps} }

func (h *Handler) RegisterPages(r chi.Router) {
	r.Get("/auth/login", h.handleLoginPage)
	r.Get("/auth/register", h.handleRegisterPage)
}

func (h *Handler) handleLoginPage(w http.ResponseWriter, r *http.Request) {
	shared.WriteHTML(w, r, http.StatusOK, authviews.LoginPage(shared.NextTargetFromRequest(r, "/products")))
}

func (h *Handler) handleRegisterPage(w http.ResponseWriter, r *http.Request) {
	shared.WriteHTML(w, r, http.StatusOK, authviews.RegisterPage(shared.NextTargetFromRequest(r, "/products")))
}
