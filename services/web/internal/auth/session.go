package auth

import (
	"net/http"

	authconfig "refurbished-marketplace/shared/auth/config"
	sharedjwt "refurbished-marketplace/shared/auth/jwt"
)

func HasValidAccessToken(cfg authconfig.Config, r *http.Request) bool {
	raw := AccessTokenFromRequest(r)
	if raw == "" {
		return false
	}
	_, err := sharedjwt.ParseAndValidate(raw, cfg.JWTSecret, "access", cfg.JWTIssuer, cfg.JWTAudience)
	return err == nil
}
