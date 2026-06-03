package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	shared "refurbished-marketplace/services/web/internal/handlers/shared"
	paymentv1 "refurbished-marketplace/shared/proto/payment/v1"

	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type hostedPaymentCallbackRequest struct {
	OrderID          string `json:"order_id"`
	PaymentSessionID string `json:"payment_session_id"`
	Status           string `json:"status"`
	FailureReason    string `json:"failure_reason"`
}

func (h *Handler) registerCallbackRoutes(r chi.Router) {
	r.Post("/callbacks/hosted-payment", h.handleHostedPaymentCallback)
}

func (h *Handler) handleHostedPaymentCallback(w http.ResponseWriter, r *http.Request) {
	if h.deps == nil || h.deps.Payment == nil {
		http.Error(w, "payment unavailable", http.StatusServiceUnavailable)
		return
	}

	var req hostedPaymentCallbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.OrderID) == "" {
		http.Error(w, "invalid order_id", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.PaymentSessionID) == "" {
		http.Error(w, "payment_session_id is required", http.StatusBadRequest)
		return
	}
	statusValue, ok := parseHostedPaymentCallbackStatus(strings.TrimSpace(req.Status))
	if !ok {
		http.Error(w, "invalid status", http.StatusBadRequest)
		return
	}

	_, err := h.deps.Payment.HandleGatewayWebhook(r.Context(), &paymentv1.HandleGatewayWebhookRequest{
		OrderId:          strings.TrimSpace(req.OrderID),
		PaymentSessionId: strings.TrimSpace(req.PaymentSessionID),
		Status:           statusValue,
		FailureReason:    strings.TrimSpace(req.FailureReason),
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.InvalidArgument:
				http.Error(w, st.Message(), http.StatusBadRequest)
				return
			case codes.NotFound:
				http.Error(w, "payment session not found", http.StatusNotFound)
				return
			}
		}
		shared.WriteGRPCError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func parseHostedPaymentCallbackStatus(v string) (paymentv1.HostedPaymentSessionStatus, bool) {
	switch strings.ToUpper(v) {
	case "SUCCEEDED":
		return paymentv1.HostedPaymentSessionStatus_HOSTED_PAYMENT_SESSION_STATUS_SUCCEEDED, true
	case "FAILED":
		return paymentv1.HostedPaymentSessionStatus_HOSTED_PAYMENT_SESSION_STATUS_FAILED, true
	case "CANCELLED":
		return paymentv1.HostedPaymentSessionStatus_HOSTED_PAYMENT_SESSION_STATUS_CANCELLED, true
	case "EXPIRED":
		return paymentv1.HostedPaymentSessionStatus_HOSTED_PAYMENT_SESSION_STATUS_EXPIRED, true
	default:
		return paymentv1.HostedPaymentSessionStatus_HOSTED_PAYMENT_SESSION_STATUS_UNSPECIFIED, false
	}
}
