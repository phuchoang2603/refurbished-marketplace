package handlers

import (
	"net/http"
	"strings"

	paymentv1 "refurbished-marketplace/shared/proto/payment/v1"
)

// Minimal webhook contract for the Stripe simulator.
// In later iterations, add signature verification and stricter schema validation.
type stripeSimWebhookRequest struct {
	PaymentTransactionID string `json:"payment_transaction_id"`
	GatewayTransactionID string `json:"gateway_transaction_id"`
	Status               string `json:"status"`         // "APPROVED" | "DECLINED"
	FailureReason        string `json:"failure_reason"` // optional
}

func (h *Handler) handleStripeSimWebhook(w http.ResponseWriter, r *http.Request) {
	var req stripeSimWebhookRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.PaymentTransactionID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "payment_transaction_id is required"})
		return
	}
	if req.Status == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "status is required"})
		return
	}

	status := strings.TrimSpace(strings.ToUpper(req.Status))
	paymentStatus := paymentv1.PaymentTransactionStatus_PAYMENT_TRANSACTION_STATUS_UNSPECIFIED
	switch status {
	case "APPROVED", "SUCCEEDED", "SUCCESS":
		paymentStatus = paymentv1.PaymentTransactionStatus_PAYMENT_TRANSACTION_STATUS_SUCCEEDED
	case "DECLINED", "FAILED", "FAILURE":
		paymentStatus = paymentv1.PaymentTransactionStatus_PAYMENT_TRANSACTION_STATUS_FAILED
	default:
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid status"})
		return
	}

	_, err := h.payment.HandleGatewayWebhook(r.Context(), &paymentv1.HandleGatewayWebhookRequest{
		PaymentTransactionId: req.PaymentTransactionID,
		GatewayTransactionId: req.GatewayTransactionID,
		Status:               paymentStatus,
		FailureReason:        req.FailureReason,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
