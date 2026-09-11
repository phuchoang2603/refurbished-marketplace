package handlers

import (
	"net/http"

	webAuth "github.com/phuchoang2603/refurbished-marketplace/services/web/internal/auth"
	shared "github.com/phuchoang2603/refurbished-marketplace/services/web/internal/handlers/shared"
	sharedviews "github.com/phuchoang2603/refurbished-marketplace/services/web/internal/views/shared"
)

func (h *Handler) authUserFromRequest(r *http.Request) (string, string, sharedviews.AuthState, bool) {
	claims, ok := webAuth.AccessClaimsFromRequest(h.authCfg, r)
	return claims.Subject, claims.Email, sharedviews.AuthState{Authenticated: ok}, ok
}

func (h *Handler) requireAccessToken() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, email, _, ok := h.authUserFromRequest(r)
			if !ok {
				shared.WritePopup(w, r, http.StatusUnauthorized, "Unauthorized", "you are not authenticated")
				return
			}

			ctx := webAuth.ContextWithUserID(r.Context(), userID)
			ctx = webAuth.ContextWithEmail(ctx, email)
			ctx = sharedviews.WithAuthState(ctx, sharedviews.AuthState{Authenticated: true})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (h *Handler) viewAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, email, state, ok := h.authUserFromRequest(r)
		ctx := sharedviews.WithAuthState(r.Context(), state)
		if ok {
			ctx = webAuth.ContextWithUserID(ctx, userID)
			ctx = webAuth.ContextWithEmail(ctx, email)
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
