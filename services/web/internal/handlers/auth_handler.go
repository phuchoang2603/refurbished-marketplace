package handlers

import (
	"encoding/json"
	"net/http"
)

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

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

func mapTokens(access, refresh, tokenType string, expiresIn int64) tokenResponse {
	return tokenResponse{AccessToken: access, RefreshToken: refresh, TokenType: tokenType, ExpiresIn: expiresIn}
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	tokens, err := h.users.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, mapTokens(tokens.AccessToken, tokens.RefreshToken, tokens.TokenType, tokens.ExpiresIn))
}

func (h *Handler) handleRefresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	tokens, err := h.users.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, mapTokens(tokens.AccessToken, tokens.RefreshToken, tokens.TokenType, tokens.ExpiresIn))
}

func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	var req logoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	_, err := h.users.Logout(r.Context(), req.RefreshToken)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
