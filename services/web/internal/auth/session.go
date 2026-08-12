package auth

import (
	"net/http"

	authconfig "github.com/phuchoang2603/refurbished-marketplace/shared/auth/config"
	sharedjwt "github.com/phuchoang2603/refurbished-marketplace/shared/auth/jwt"
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
