package grpcauth_test

import (
	"context"
	"testing"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	authconfig "github.com/phuchoang2603/refurbished-marketplace/shared/auth/config"
	"github.com/phuchoang2603/refurbished-marketplace/shared/auth/grpcauth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const secret = "grpcauth-test-secret"

func accessToken(t *testing.T, subject string) string {
	t.Helper()
	token, err := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, jwtlib.MapClaims{
		"typ": "access", "iss": authconfig.DefaultJWTIssuer, "aud": authconfig.DefaultJWTAudience,
		"sub": subject, "jti": "jti-1", "exp": time.Now().Add(time.Hour).Unix(),
	}).SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}
	return token
}

// forward runs the client interceptor and hands its outgoing metadata to the server interceptor.
func forward(t *testing.T, ctx context.Context) (string, error) {
	t.Helper()
	var outgoing metadata.MD
	invoker := func(ctx context.Context, _ string, _, _ any, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
		outgoing, _ = metadata.FromOutgoingContext(ctx)
		return nil
	}
	if err := grpcauth.UnaryClientInterceptor()(ctx, "/svc/Method", nil, nil, nil, invoker); err != nil {
		t.Fatal(err)
	}
	var subject string
	handler := func(ctx context.Context, _ any) (any, error) {
		claims, ok := grpcauth.ClaimsFromContext(ctx)
		if !ok {
			t.Fatal("handler ran without claims")
		}
		subject = claims.Subject
		return nil, nil
	}
	server := grpcauth.UnaryServerInterceptor(authconfig.DefaultConfig(secret))
	_, err := server(metadata.NewIncomingContext(t.Context(), outgoing), nil, &grpc.UnaryServerInfo{}, handler)
	return subject, err
}

func TestAccessTokenRoundTrip(t *testing.T) {
	subject, err := forward(t, grpcauth.WithAccessToken(t.Context(), accessToken(t, "buyer-1")))
	if err != nil || subject != "buyer-1" {
		t.Fatalf("expected buyer-1, got subject=%q err=%v", subject, err)
	}
}

func TestMissingAccessTokenIsUnauthenticated(t *testing.T) {
	if _, err := forward(t, t.Context()); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated, got %v", err)
	}
}

func TestInvalidAccessTokenIsUnauthenticated(t *testing.T) {
	_, err := forward(t, grpcauth.WithAccessToken(t.Context(), "not-a-jwt"))
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated, got %v", err)
	}
}
