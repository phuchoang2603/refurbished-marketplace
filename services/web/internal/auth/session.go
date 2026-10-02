package auth

import (
	"context"
	"net/http"

	authconfig "github.com/phuchoang2603/refurbished-marketplace/shared/auth/config"
	sharedjwt "github.com/phuchoang2603/refurbished-marketplace/shared/auth/jwt"
	"google.golang.org/grpc/metadata"
)

func OutgoingAccessContext(r *http.Request) context.Context {
	return metadata.AppendToOutgoingContext(r.Context(), "authorization", "Bearer "+AccessTokenFromRequest(r))
}

func AccessClaimsFromRequest(cfg authconfig.Config, r *http.Request) (sharedjwt.Claims, bool) {
	raw := AccessTokenFromRequest(r)
	if raw == "" {
		return sharedjwt.Claims{}, false
	}
	claims, err := sharedjwt.ParseAndValidate(raw, cfg.JWTSecret, "access", cfg.JWTIssuer, cfg.JWTAudience)
	if err != nil {
		return sharedjwt.Claims{}, false
	}
	return claims, claims.Subject != ""
}

func AccessUserIDFromRequest(cfg authconfig.Config, r *http.Request) (string, bool) {
	claims, ok := AccessClaimsFromRequest(cfg, r)
	if !ok {
		return "", false
	}
	return claims.Subject, true
}
