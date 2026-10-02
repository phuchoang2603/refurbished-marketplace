package cart

import (
	"net/http"

	shared "github.com/phuchoang2603/refurbished-marketplace/services/web/internal/handlers/shared"
	checkoutviews "github.com/phuchoang2603/refurbished-marketplace/services/web/internal/views/checkout"
	checkoutv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/checkout/v1"
	paymentv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/payment/v1"

	"github.com/starfederation/datastar-go/datastar"
)

func (h *Handler) getOwnedCheckout(w http.ResponseWriter, r *http.Request) (*checkoutv1.CheckoutStatus, bool) {
	buyerUserID, ok := shared.RequireUserID(w, r)
	if !ok {
		return nil, false
	}
	checkoutID, ok := shared.RequirePathValue(w, r, "id", "invalid checkout id")
	if !ok {
		return nil, false
	}
	if h.deps.Checkout == nil {
		shared.WriteBadRequest(w, r, "checkout unavailable")
		return nil, false
	}
	status, err := h.deps.Checkout.GetCheckout(r.Context(), &checkoutv1.GetCheckoutRequest{
		CheckoutId: checkoutID, BuyerUserId: buyerUserID,
	})
	if err != nil {
		shared.WriteGRPCError(w, r, err)
		return nil, false
	}
	return status, true
}

func (h *Handler) paymentURL(r *http.Request, status *checkoutv1.CheckoutStatus) string {
	if status.GetState() != checkoutv1.CheckoutState_CHECKOUT_STATE_READY_TO_PAY ||
		status.GetPaymentSessionId() == "" || status.GetReturnUrl() == "" {
		return ""
	}
	return shared.BuildHostedPaymentURL(h.deps.HostedPayment, r, &paymentv1.CreateHostedPaymentSessionResponse{
		OrderId: status.GetOrderId(), PaymentSessionId: status.GetPaymentSessionId(), ReturnUrl: status.GetReturnUrl(),
	})
}

func (h *Handler) handleCheckoutProgressPage(w http.ResponseWriter, r *http.Request) {
	status, ok := h.getOwnedCheckout(w, r)
	if !ok {
		return
	}
	if url := h.paymentURL(r, status); url != "" {
		shared.Redirect(w, r, url, http.StatusSeeOther)
		return
	}
	shared.WriteHTML(w, r, http.StatusOK, checkoutviews.ProgressPage(status))
}

func (h *Handler) handleCheckoutProgress(w http.ResponseWriter, r *http.Request) {
	status, ok := h.getOwnedCheckout(w, r)
	if !ok {
		return
	}
	sse := datastar.NewSSE(w, r)
	if url := h.paymentURL(r, status); url != "" {
		_ = sse.Redirect(url)
		return
	}
	_ = sse.PatchElementTempl(checkoutviews.Progress(status))
}
