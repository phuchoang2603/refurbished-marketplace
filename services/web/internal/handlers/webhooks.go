package handlers

import (
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	checkoutv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/checkout/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *Handler) registerWebhookRoutes(r chi.Router) {
	r.Post("/webhooks/payment-gateway", h.handleGatewayWebhook)
}

func (h *Handler) handleGatewayWebhook(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if h.deps == nil || h.deps.Checkout == nil {
		http.Error(w, "checkout unavailable", http.StatusServiceUnavailable)
		return
	}
	result, err := h.deps.Checkout.HandleGatewayWebhook(r.Context(), &checkoutv1.HandleGatewayWebhookRequest{
		RawBody: body, SignatureHeader: r.Header.Get("Gateway-Signature"),
	})
	if err != nil {
		if status.Code(err) == codes.InvalidArgument {
			http.Error(w, "invalid webhook", http.StatusBadRequest)
		} else {
			http.Error(w, "checkout unavailable", http.StatusServiceUnavailable)
		}
		return
	}
	switch result.GetResult() {
	case checkoutv1.GatewayWebhookResult_GATEWAY_WEBHOOK_RESULT_ACCEPTED, checkoutv1.GatewayWebhookResult_GATEWAY_WEBHOOK_RESULT_DUPLICATE:
		w.WriteHeader(http.StatusNoContent)
	case checkoutv1.GatewayWebhookResult_GATEWAY_WEBHOOK_RESULT_REJECTED:
		http.Error(w, "invalid webhook", http.StatusBadRequest)
	default:
		http.Error(w, "checkout unavailable", http.StatusServiceUnavailable)
	}
}
