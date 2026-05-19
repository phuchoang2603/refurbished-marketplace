package auth

import (
	"net/http"

	authconfig "refurbished-marketplace/shared/auth/config"
	sharedjwt "refurbished-marketplace/shared/auth/jwt"
)

func AccessUserIDFromRequest(cfg authconfig.Config, r *http.Request) (string, bool) {
	raw := AccessTokenFromRequest(r)
	if raw == "" {
		return "", false
	}
	claims, err := sharedjwt.ParseAndValidate(raw, cfg.JWTSecret, "access", cfg.JWTIssuer, cfg.JWTAudience)
	if err != nil {
		return "", false
	}
	return claims.Subject, claims.Subject != ""
}
