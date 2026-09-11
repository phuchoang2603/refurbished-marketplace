package cart

import (
	"net/http"
	"strings"

	shared "github.com/phuchoang2603/refurbished-marketplace/services/web/internal/handlers/shared"
	sharedlog "github.com/phuchoang2603/refurbished-marketplace/shared/observe/log"
	paymentv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/payment/v1"
)

func DrainPaidProductIDs(w http.ResponseWriter, r *http.Request, cartSvc shared.CartService, productIDs []string, merchantID string) {
	if cartSvc == nil || len(productIDs) == 0 {
		return
	}
	cartID := ""
	if c, err := r.Cookie(cartCookieName); err == nil {
		cartID = strings.TrimSpace(c.Value)
	}
	if cartID == "" {
		return
	}
	updated, err := cartSvc.RemoveCartItems(r.Context(), cartID, productIDs)
	if err != nil {
		sharedlog.ErrorContext(r.Context(), "cart drain after payment failed", sharedlog.KeyErr, err, "cart_id", cartID)
		return
	}
	clearCheckoutIntent(w, r, cartID, merchantID)
	if updated != nil && len(updated.GetItems()) == 0 {
		http.SetCookie(w, &http.Cookie{Name: cartCookieName, Value: "", Path: "/", HttpOnly: true, MaxAge: -1, SameSite: http.SameSiteLaxMode})
		writeCheckoutIntentStore(w, checkoutIntentStore{})
	}
}

func HostedSessionDrainsCart(status string) bool {
	switch status {
	case "SUCCEEDED", "FAILED", "EXPIRED":
		return true
	default:
		return false
	}
}

func HostedSessionDrainsCartStatus(status paymentv1.HostedPaymentSessionStatus) bool {
	switch status {
	case paymentv1.HostedPaymentSessionStatus_HOSTED_PAYMENT_SESSION_STATUS_SUCCEEDED,
		paymentv1.HostedPaymentSessionStatus_HOSTED_PAYMENT_SESSION_STATUS_FAILED,
		paymentv1.HostedPaymentSessionStatus_HOSTED_PAYMENT_SESSION_STATUS_EXPIRED:
		return true
	default:
		return false
	}
}
