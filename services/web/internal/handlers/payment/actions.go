package payment

import (
	"net/http"
	"strings"

	shared "refurbished-marketplace/services/web/internal/handlers/shared"
	paymentv1 "refurbished-marketplace/shared/proto/payment/v1"

	"github.com/go-chi/chi/v5"
)

type Handler struct{ deps *shared.Dependencies }

func New(deps *shared.Dependencies) *Handler { return &Handler{deps: deps} }

type stripeSimWebhookRequest struct {
	PaymentTransactionID string `json:"payment_transaction_id"`
	GatewayTransactionID string `json:"gateway_transaction_id"`
	Status               string `json:"status"`
	FailureReason        string `json:"failure_reason"`
}

func (h *Handler) RegisterActions(r chi.Router) {
	r.Post("/webhooks/stripe-simulator", h.handleStripeSimWebhook)
}

func (h *Handler) handleStripeSimWebhook(w http.ResponseWriter, r *http.Request) {
	var req stripeSimWebhookRequest
	if !shared.DecodeJSONResponse(w, r, &req) {
		return
	}
	if req.PaymentTransactionID == "" {
		shared.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "payment_transaction_id is required"})
		return
	}
	if req.Status == "" {
		shared.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "status is required"})
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
		shared.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid status"})
		return
	}

	_, err := h.deps.Payment.HandleGatewayWebhook(r.Context(), &paymentv1.HandleGatewayWebhookRequest{PaymentTransactionId: req.PaymentTransactionID, GatewayTransactionId: req.GatewayTransactionID, Status: paymentStatus, FailureReason: req.FailureReason})
	if err != nil {
		shared.WriteGRPCError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
