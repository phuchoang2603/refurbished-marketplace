package handlers

import (
	"net/http"

	webAuth "refurbished-marketplace/services/web/internal/auth"
	shared "refurbished-marketplace/services/web/internal/handlers/shared"
	sharedviews "refurbished-marketplace/services/web/internal/views/shared"
)

func (h *Handler) authUserFromRequest(r *http.Request) (string, sharedviews.AuthState, bool) {
	userID, ok := webAuth.AccessUserIDFromRequest(h.authCfg, r)
	return userID, sharedviews.AuthState{Authenticated: ok}, ok
}

func (h *Handler) requireAccessToken() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, _, ok := h.authUserFromRequest(r)
			if !ok {
				shared.WritePopup(w, r, http.StatusUnauthorized, "Unauthorized", "you are not authenticated")
				return
			}

			ctx := webAuth.ContextWithUserID(r.Context(), userID)
			ctx = sharedviews.WithAuthState(ctx, sharedviews.AuthState{Authenticated: true})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (h *Handler) viewAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, state, ok := h.authUserFromRequest(r)
		ctx := sharedviews.WithAuthState(r.Context(), state)
		if ok {
			ctx = webAuth.ContextWithUserID(ctx, userID)
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
