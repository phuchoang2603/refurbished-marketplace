package handlers

import (
	"net/http"

	"refurbished-marketplace/services/web/internal/views"
)

func mapTokenView(access, refresh, tokenType string, expiresIn int64) views.TokenView {
	return views.TokenView{AccessToken: access, RefreshToken: refresh, TokenType: tokenType, ExpiresIn: expiresIn}
}

func loginCredentialsFromForm(r *http.Request) (string, string, error) {
	if !parseForm(r) {
		return "", "", errInvalidRequestBody
	}
	return r.FormValue("email"), r.FormValue("password"), nil
}

func refreshTokenFromForm(r *http.Request) (string, error) {
	if !parseForm(r) {
		return "", errInvalidRequestBody
	}
	return r.FormValue("refresh_token"), nil
}

func (h *Handler) handleLoginPage(w http.ResponseWriter, r *http.Request) {
	writeHTML(w, r, http.StatusOK, views.LoginPage())
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	email, password, err := loginCredentialsFromForm(r)
	if err != nil || email == "" || password == "" {
		writeBadRequest(w, r, "invalid request body")
		return
	}

	tokens, err := h.users.Login(r.Context(), email, password)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeHTML(w, r, http.StatusOK, views.TokensPage(mapTokenView(tokens.AccessToken, tokens.RefreshToken, tokens.TokenType, tokens.ExpiresIn)))
}

func (h *Handler) handleRefreshPage(w http.ResponseWriter, r *http.Request) {
	writeHTML(w, r, http.StatusOK, views.RefreshPage())
}

func (h *Handler) handleRefresh(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := refreshTokenFromForm(r)
	if err != nil || refreshToken == "" {
		writeBadRequest(w, r, "invalid request body")
		return
	}

	tokens, err := h.users.Refresh(r.Context(), refreshToken)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeHTML(w, r, http.StatusOK, views.TokensPage(mapTokenView(tokens.AccessToken, tokens.RefreshToken, tokens.TokenType, tokens.ExpiresIn)))
}

func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := refreshTokenFromForm(r)
	if err != nil || refreshToken == "" {
		writeBadRequest(w, r, "invalid request body")
		return
	}

	_, err = h.users.Logout(r.Context(), refreshToken)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeHTML(w, r, http.StatusOK, views.MessagePage("Logged out", "Your session has been cleared."))
}
