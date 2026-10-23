package service

import (
	"database/sql"
	"testing"

	"github.com/phuchoang2603/refurbished-marketplace/services/checkout/internal/database"
	checkoutv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/checkout/v1"
)

func TestGatewayOutcomeSettlementDecisions(t *testing.T) {
	for _, tc := range []struct {
		name          string
		state         checkoutv1.CheckoutState
		paymentStatus string
		event         string
		contradiction bool
		next          checkoutv1.CheckoutState
	}{
		{
			name: "decline preserves open payment", state: checkoutv1.CheckoutState_CHECKOUT_STATE_READY_TO_PAY,
			paymentStatus: "OPEN", event: "attempt.declined",
		},
		{
			name: "late success resolves uncertain payment", state: checkoutv1.CheckoutState_CHECKOUT_STATE_NEEDS_REVIEW,
			paymentStatus: "OPEN", event: "payment.succeeded", next: checkoutv1.CheckoutState_CHECKOUT_STATE_COMMITTING_STOCK,
		},
		{
			name: "late expiry resolves uncertain payment", state: checkoutv1.CheckoutState_CHECKOUT_STATE_NEEDS_REVIEW,
			paymentStatus: "OPEN", event: "payment.expired", next: checkoutv1.CheckoutState_CHECKOUT_STATE_RELEASING_STOCK,
		},
		{
			name: "review without known open payment stays paused", state: checkoutv1.CheckoutState_CHECKOUT_STATE_NEEDS_REVIEW,
			event: "payment.succeeded",
		},
		{
			name: "expiry cannot undo completed payment", state: checkoutv1.CheckoutState_CHECKOUT_STATE_COMPLETED,
			paymentStatus: "SUCCEEDED", event: "payment.expired", contradiction: true,
		},
		{
			name: "success cannot undo released stock", state: checkoutv1.CheckoutState_CHECKOUT_STATE_RELEASING_STOCK,
			paymentStatus: "OPEN", event: "payment.succeeded", contradiction: true,
		},
		{
			name: "recorded expiry remains contradictory in review", state: checkoutv1.CheckoutState_CHECKOUT_STATE_NEEDS_REVIEW,
			paymentStatus: "EXPIRED", event: "payment.succeeded", contradiction: true,
		},
		{
			name: "repeated success does not recommit stock", state: checkoutv1.CheckoutState_CHECKOUT_STATE_COMMITTING_STOCK,
			paymentStatus: "SUCCEEDED", event: "payment.succeeded",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			row := database.Checkout{State: tc.state.String(), PaymentStatus: sql.NullString{String: tc.paymentStatus, Valid: tc.paymentStatus != ""}}
			if got := gatewayOutcomeContradicts(row, tc.event); got != tc.contradiction {
				t.Fatalf("contradiction=%t, want %t", got, tc.contradiction)
			}
			if tc.contradiction {
				return
			}
			settlement := gatewaySettlement(row, tc.event)
			if tc.next == checkoutv1.CheckoutState_CHECKOUT_STATE_UNSPECIFIED {
				if settlement != nil {
					t.Fatalf("unexpected settlement: %+v", settlement)
				}
				return
			}
			if settlement == nil || settlement.state != tc.next {
				t.Fatalf("settlement=%+v, want %s", settlement, tc.next)
			}
		})
	}
}
