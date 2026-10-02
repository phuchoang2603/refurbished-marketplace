package tests

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/phuchoang2603/refurbished-marketplace/services/checkout/internal/grpcserver"
	checkoutv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/checkout/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

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

	_, err := server.GetCheckout(t.Context(), &checkoutv1.GetCheckoutRequest{
		CheckoutId: checkoutID.String(), BuyerUserId: uuid.NewString(),
	})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("another buyer must not see the checkout: %v", err)
	}

	response, err := server.GetCheckout(t.Context(), &checkoutv1.GetCheckoutRequest{
		CheckoutId: checkoutID.String(), BuyerUserId: buyerID.String(),
	})
	if err != nil || response.GetPaymentSessionId() != "hosted-session" {
		t.Fatalf("buyer should see the ready checkout: response=%v error=%v", response, err)
	}
}
