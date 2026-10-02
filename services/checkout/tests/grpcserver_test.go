package tests

import (
	"context"
	"testing"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/phuchoang2603/refurbished-marketplace/services/checkout/internal/grpcserver"
	authconfig "github.com/phuchoang2603/refurbished-marketplace/shared/auth/config"
	"github.com/phuchoang2603/refurbished-marketplace/shared/auth/grpcauth"
	checkoutv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/checkout/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const checkoutTestJWTSecret = "checkout-test-secret"

func authenticatedContext(t *testing.T, subject string) context.Context {
	t.Helper()
	token, err := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, jwtlib.MapClaims{
		"typ": "access", "iss": authconfig.DefaultJWTIssuer,
		"aud": authconfig.DefaultJWTAudience, "sub": subject,
		"jti": uuid.NewString(), "exp": time.Now().Add(time.Hour).Unix(),
	}).SignedString([]byte(checkoutTestJWTSecret))
	if err != nil {
		t.Fatal(err)
	}
	ctx := metadata.NewIncomingContext(t.Context(), metadata.Pairs("authorization", "Bearer "+token))
	claims, err := grpcauth.Authenticate(ctx, authconfig.DefaultConfig(checkoutTestJWTSecret))
	if err != nil {
		t.Fatal(err)
	}
	return grpcauth.ContextWithClaims(ctx, claims)
}

func TestCheckoutRequiresAuthenticatedBuyer(t *testing.T) {
	server := grpcserver.New(checkoutReader{})
	_, err := server.GetCheckout(t.Context(), &checkoutv1.GetCheckoutRequest{
		CheckoutId: uuid.NewString(), BuyerUserId: uuid.NewString(),
	})
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("missing claims must be rejected: %v", err)
	}
}

type checkoutReader struct {
	view grpcserver.CheckoutView
}

func (reader checkoutReader) GetCheckout(context.Context, uuid.UUID) (grpcserver.CheckoutView, error) {
	return reader.view, nil
}

func TestGetCheckoutBuyerOwnership(t *testing.T) {
	checkoutID := uuid.New()
	buyerID := uuid.New()
	server := grpcserver.New(checkoutReader{view: grpcserver.CheckoutView{
		CheckoutID: checkoutID, BuyerUserID: buyerID,
		State:            checkoutv1.CheckoutState_CHECKOUT_STATE_READY_TO_PAY,
		PaymentSessionID: "hosted-session",
	}})
	buyerContext := authenticatedContext(t, buyerID.String())

	_, err := server.GetCheckout(buyerContext, &checkoutv1.GetCheckoutRequest{
		CheckoutId: checkoutID.String(), BuyerUserId: uuid.NewString(),
	})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("another buyer must not see the checkout: %v", err)
	}

	response, err := server.GetCheckout(buyerContext, &checkoutv1.GetCheckoutRequest{
		CheckoutId: checkoutID.String(), BuyerUserId: buyerID.String(),
	})
	if err != nil || response.GetPaymentSessionId() != "hosted-session" {
		t.Fatalf("buyer should see the ready checkout: response=%v error=%v", response, err)
	}
}
