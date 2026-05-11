package handlers

import (
	"net/http"

	webAuth "refurbished-marketplace/services/web/internal/auth"
	"refurbished-marketplace/services/web/internal/views"
)

func sessionTokenView(tokenType string, expiresIn int64) views.TokenView {
	return views.TokenView{TokenType: tokenType, ExpiresIn: expiresIn}
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
		writeGRPCError(w, r, err)
		return
	}
	webAuth.SetTokenCookies(w, r, tokens.AccessToken, tokens.RefreshToken, tokens.ExpiresIn, tokens.RefreshExpiresIn)
	r = r.WithContext(views.WithAuthState(r.Context(), views.AuthState{Authenticated: true}))

	writeHTML(w, r, http.StatusOK, views.TokensPage(sessionTokenView(tokens.TokenType, tokens.ExpiresIn)))
}

func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := refreshTokenFromForm(r)
	if err != nil || refreshToken == "" {
		refreshToken = webAuth.RefreshTokenFromRequest(r)
	}
	if refreshToken == "" {
		webAuth.ClearTokenCookies(w, r)
		r = r.WithContext(views.WithAuthState(r.Context(), views.AuthState{Authenticated: false}))
		writeHTML(w, r, http.StatusOK, views.MessagePage("Logged out", "Your browser session has been cleared."))
		return
	}

	_, err = h.users.Logout(r.Context(), refreshToken)
	webAuth.ClearTokenCookies(w, r)
	r = r.WithContext(views.WithAuthState(r.Context(), views.AuthState{Authenticated: false}))
	if err != nil {
		writeGRPCError(w, r, err)
		return
	}

	writeHTML(w, r, http.StatusOK, views.MessagePage("Logged out", "Your session has been cleared."))
}
