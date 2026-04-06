package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	authconfig "refurbished-marketplace/shared/auth/config"
	sharedjwt "refurbished-marketplace/shared/auth/jwt"
)

type contextKey string

const userIDKey contextKey = "authUserID"

func UserIDFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(userIDKey).(string)
	return v, ok
}

func RequireAccessToken(cfg authconfig.Config, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw := bearerToken(r.Header.Get("Authorization"))
		if raw == "" {
			writeUnauthorized(w)
			return
		}

		claims, err := sharedjwt.ParseAndValidate(raw, cfg.JWTSecret, "access", cfg.JWTIssuer, cfg.JWTAudience)
		if err != nil {
			if errors.Is(err, sharedjwt.ErrExpiredToken) || errors.Is(err, sharedjwt.ErrInvalidToken) {
				writeUnauthorized(w)
				return
			}
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, claims.Subject)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func bearerToken(header string) string {
	header = strings.TrimSpace(header)
	if header == "" {
		return ""
	}

	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return ""
	}

	token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	if token == "" {
		return ""
	}

	return token
}

func writeUnauthorized(w http.ResponseWriter) {
	http.Error(w, "unauthorized", http.StatusUnauthorized)
}
