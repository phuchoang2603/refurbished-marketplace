package auth

import (
	"net/http"

	webAuth "refurbished-marketplace/services/web/internal/auth"
	shared "refurbished-marketplace/services/web/internal/handlers/shared"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) RegisterActions(r chi.Router) {
	r.Post("/auth/login", h.handleLogin)
	r.Post("/auth/register", h.handleRegister)
	r.Post("/auth/logout", h.handleLogout)
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	email, password, err := shared.EmailPasswordFromForm(r)
	if err != nil || email == "" || password == "" {
		shared.WriteBadRequest(w, r, "invalid request body")
		return
	}

	tokens, err := h.deps.Users.Login(r.Context(), email, password)
	if err != nil {
		shared.WriteGRPCError(w, r, err)
		return
	}
	webAuth.SetTokenCookies(w, r, tokens.AccessToken, tokens.RefreshToken, tokens.ExpiresIn, tokens.RefreshExpiresIn)
	http.Redirect(w, r, shared.NextTargetFromRequest(r, "/products"), http.StatusSeeOther)
}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	email, password, err := shared.EmailPasswordFromForm(r)
	if err != nil || email == "" || password == "" {
		shared.WriteBadRequest(w, r, "invalid request body")
		return
	}

	_, err = h.deps.Users.CreateUser(r.Context(), email, password)
	if err != nil {
		shared.WriteGRPCError(w, r, err)
		return
	}
	tokens, err := h.deps.Users.Login(r.Context(), email, password)
	if err != nil {
		shared.WriteGRPCError(w, r, err)
		return
	}
	webAuth.SetTokenCookies(w, r, tokens.AccessToken, tokens.RefreshToken, tokens.ExpiresIn, tokens.RefreshExpiresIn)
	http.Redirect(w, r, shared.NextTargetFromRequest(r, "/products"), http.StatusSeeOther)
}

func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := shared.RefreshTokenFromForm(r)
	if err != nil || refreshToken == "" {
		refreshToken = webAuth.RefreshTokenFromRequest(r)
	}
	if refreshToken == "" {
		webAuth.ClearTokenCookies(w, r)
		http.Redirect(w, r, "/products", http.StatusSeeOther)
		return
	}

	_, err = h.deps.Users.Logout(r.Context(), refreshToken)
	webAuth.ClearTokenCookies(w, r)
	if err != nil {
		shared.WriteGRPCError(w, r, err)
		return
	}

	http.Redirect(w, r, "/products", http.StatusSeeOther)
}
