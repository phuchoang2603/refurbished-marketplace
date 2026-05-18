package auth

import (
	"errors"
	"net/http"

	authconfig "refurbished-marketplace/shared/auth/config"
	sharedjwt "refurbished-marketplace/shared/auth/jwt"
)

type UnauthorizedHandler func(http.ResponseWriter, *http.Request)

func RequireAccessToken(cfg authconfig.Config, next http.Handler, unauthorized UnauthorizedHandler) http.Handler {
	if unauthorized == nil {
		unauthorized = func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
		}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw := AccessTokenFromRequest(r)
		if raw == "" {
			unauthorized(w, r)
			return
		}

		claims, err := sharedjwt.ParseAndValidate(raw, cfg.JWTSecret, "access", cfg.JWTIssuer, cfg.JWTAudience)
		if err != nil {
			if errors.Is(err, sharedjwt.ErrExpiredToken) || errors.Is(err, sharedjwt.ErrInvalidToken) {
				unauthorized(w, r)
				return
			}
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		ctx := ContextWithUserID(r.Context(), claims.Subject)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
