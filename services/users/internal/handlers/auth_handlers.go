package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"refurbished-marketplace/services/users/internal/service"
)

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	tokens, err := h.svc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, mapTokens(tokens))
}

func (h *Handler) handleRefresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	tokens, err := h.svc.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidToken), errors.Is(err, service.ErrTokenExpired), errors.Is(err, service.ErrTokenRevoked):
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid refresh token"})
		default:
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		}
		return
	}

	writeJSON(w, http.StatusOK, mapTokens(tokens))
}

func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	var req logoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	err := h.svc.Logout(r.Context(), req.RefreshToken)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidToken), errors.Is(err, service.ErrTokenExpired), errors.Is(err, service.ErrTokenRevoked):
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid refresh token"})
		default:
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
