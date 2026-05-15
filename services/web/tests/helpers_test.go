package tests

import (
	"bytes"
	"context"
	"testing"
	"time"

	authconfig "refurbished-marketplace/shared/auth/config"

	"github.com/a-h/templ"
	jwtlib "github.com/golang-jwt/jwt/v5"
)

func renderToString(t *testing.T, c templ.Component) string {
	t.Helper()
	return renderWithContext(t, context.Background(), c)
}

func renderWithContext(t *testing.T, ctx context.Context, c templ.Component) string {
	t.Helper()
	var buf bytes.Buffer
	if err := c.Render(ctx, &buf); err != nil {
		t.Fatalf("render failed: %v", err)
	}
	return buf.String()
}

func signedAccessToken(t *testing.T, secret, subject string) string {
	t.Helper()
	claims := jwtlib.MapClaims{
		"typ": "access",
		"iss": authconfig.DefaultJWTIssuer,
		"aud": authconfig.DefaultJWTAudience,
		"sub": subject,
		"jti": "test-token",
		"exp": time.Now().Add(time.Hour).Unix(),
	}
	token, err := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return token
}
