package handlers

import (
	"net/http"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshTokenRequest struct {
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
	if !decodeJSON(w, r, &req) {
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
	var req refreshTokenRequest
	if !decodeJSON(w, r, &req) {
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
	var req refreshTokenRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	_, err := h.users.Logout(r.Context(), req.RefreshToken)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
