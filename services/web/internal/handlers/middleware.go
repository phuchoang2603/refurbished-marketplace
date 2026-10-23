package handlers

import (
	"context"
	"errors"
	"net/http"

	"golang.org/x/sync/singleflight"

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

var errEmptyRefresh = errors.New("empty refresh response")

type refreshResult struct {
	accessToken      string
	refreshToken     string
	expiresIn        int64
	refreshExpiresIn int64
}

func (h *Handler) refreshAccess(next http.Handler) http.Handler {
	var group singleflight.Group
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := webAuth.AccessClaimsFromRequest(h.authCfg, r); ok {
			next.ServeHTTP(w, r)
			return
		}
		refreshToken := webAuth.RefreshTokenFromRequest(r)
		if refreshToken == "" || r.URL.Path == "/auth/logout" {
			next.ServeHTTP(w, r)
			return
		}

		v, err, _ := group.Do(refreshToken, func() (any, error) {
			tokens, err := h.deps.Users.Refresh(context.WithoutCancel(r.Context()), refreshToken)
			if err != nil {
				return nil, err
			}
			if tokens == nil {
				return nil, errEmptyRefresh
			}
			return refreshResult{tokens.AccessToken, tokens.RefreshToken, tokens.ExpiresIn, tokens.RefreshExpiresIn}, nil
		})
		if err != nil {
			webAuth.ClearTokenCookies(w, r)
			next.ServeHTTP(w, r)
			return
		}

		res := v.(refreshResult)
		webAuth.SetTokenCookies(w, r, res.accessToken, res.refreshToken, res.expiresIn, res.refreshExpiresIn)
		next.ServeHTTP(w, r.WithContext(webAuth.ContextWithAccessToken(r.Context(), res.accessToken)))
	})
}
