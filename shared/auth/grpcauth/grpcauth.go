// Package grpcauth carries buyer access tokens across gRPC calls.
//
// Callers attach the raw token once per request with WithAccessToken and dial
// with UnaryClientInterceptor; servers install UnaryServerInterceptor and read
// the validated identity with ClaimsFromContext.
package grpcauth

import (
	"context"
	"strings"

	authconfig "github.com/phuchoang2603/refurbished-marketplace/shared/auth/config"
	sharedjwt "github.com/phuchoang2603/refurbished-marketplace/shared/auth/jwt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	metadataKey  = "authorization"
	bearerPrefix = "Bearer "
)

type accessTokenKey struct{}

type claimsKey struct{}

func WithAccessToken(ctx context.Context, raw string) context.Context {
	return context.WithValue(ctx, accessTokenKey{}, strings.TrimSpace(raw))
}

func AccessTokenFromContext(ctx context.Context) string {
	raw, _ := ctx.Value(accessTokenKey{}).(string)
	return raw
}

func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if raw := AccessTokenFromContext(ctx); raw != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, metadataKey, bearerPrefix+raw)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func ContextWithClaims(ctx context.Context, claims sharedjwt.Claims) context.Context {
	return context.WithValue(ctx, claimsKey{}, claims)
}

func ClaimsFromContext(ctx context.Context) (sharedjwt.Claims, bool) {
	claims, ok := ctx.Value(claimsKey{}).(sharedjwt.Claims)
	return claims, ok && claims.Subject != ""
}

// UnaryServerInterceptor rejects every RPC on the server without a valid access token.
func UnaryServerInterceptor(cfg authconfig.Config) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		claims, err := Authenticate(ctx, cfg)
		if err != nil {
			return nil, err
		}
		return handler(ContextWithClaims(ctx, claims), req)
	}
}

func Authenticate(ctx context.Context, cfg authconfig.Config) (sharedjwt.Claims, error) {
	values := metadata.ValueFromIncomingContext(ctx, metadataKey)
	if len(values) != 1 {
		return sharedjwt.Claims{}, status.Error(codes.Unauthenticated, "access token required")
	}
	raw, ok := strings.CutPrefix(values[0], bearerPrefix)
	if !ok || strings.TrimSpace(raw) == "" {
		return sharedjwt.Claims{}, status.Error(codes.Unauthenticated, "access token required")
	}
	claims, err := sharedjwt.ParseAndValidate(strings.TrimSpace(raw), cfg.JWTSecret, "access", cfg.JWTIssuer, cfg.JWTAudience)
	if err != nil {
		return sharedjwt.Claims{}, status.Error(codes.Unauthenticated, "invalid access token")
	}
	return claims, nil
}
