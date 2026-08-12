package orders

import (
	"net/http"
	"strings"

	shared "refurbished-marketplace/services/web/internal/handlers/shared"
	ordersv1 "refurbished-marketplace/shared/proto/orders/v1"
	paymentv1 "refurbished-marketplace/shared/proto/payment/v1"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *Handler) RegisterActions(r chi.Router) {
	r.Post("/orders", h.handleCreateOrder)
	r.Post("/orders/{id}/resume-payment", h.handleResumeHostedPayment)
}

func (h *Handler) buildCreateOrderItem(w http.ResponseWriter, r *http.Request, productID string, quantity int32) (*ordersv1.CreateOrderItem, string, int64, bool) {
	productID = strings.TrimSpace(productID)
	if productID == "" || quantity <= 0 {
		shared.WriteBadRequest(w, r, "invalid request body")
		return nil, "", 0, false
	}

	product, err := h.deps.Products.GetProductByID(r.Context(), productID)
	if err != nil {
		shared.WriteGRPCError(w, r, err)
		return nil, "", 0, false
	}

	item := &ordersv1.CreateOrderItem{ProductId: productID, Quantity: quantity, UnitPriceCents: product.PriceCents}
	return item, product.GetMerchantId(), product.PriceCents * int64(quantity), true
}

func (h *Handler) handleCreateOrder(w http.ResponseWriter, r *http.Request) {
	buyerUserID, ok := shared.RequireUserID(w, r)
	if !ok {
		return
	}

	productID, quantity, err := shared.ProductQuantityFromForm(r)
	if err != nil {
		shared.WriteBadRequest(w, r, "invalid request body")
		return
	}
	item, merchantID, totalCents, ok := h.buildCreateOrderItem(w, r, productID, quantity)
	if !ok {
		return
	}

	order, err := h.deps.Orders.CreateOrder(r.Context(), buyerUserID, merchantID, []*ordersv1.CreateOrderItem{item}, totalCents, uuid.NewString())
	if err != nil {
		shared.WriteGRPCError(w, r, err)
		return
	}

	shared.Redirect(w, r, "/orders/"+order.GetId(), http.StatusSeeOther)
}

func (h *Handler) handleResumeHostedPayment(w http.ResponseWriter, r *http.Request) {
	buyerUserID, ok := shared.RequireUserID(w, r)
	if !ok {
		return
	}
	orderID, ok := shared.RequirePathValue(w, r, "id", "invalid order id")
	if !ok {
		return
	}

	order, err := h.deps.Orders.GetOrderByID(r.Context(), orderID)
	if err != nil {
		shared.WriteGRPCError(w, r, err)
		return
	}
	if order.GetBuyerUserId() != buyerUserID {
		shared.WritePopup(w, r, http.StatusForbidden, "Forbidden", "order does not belong to the current user")
		return
	}
	if order.GetStatus() != ordersv1.OrderStatus_ORDER_STATUS_PENDING {
		shared.WriteBadRequest(w, r, "order is not eligible for payment resume")
		return
	}

	orderPageURL := shared.OrderPageURLWithConfig(h.deps.HostedPayment, r, order.GetId())
	if orderPageURL == "" {
		shared.WriteBadRequest(w, r, "hosted payment unavailable")
		return
	}

	hostedSession, err := h.deps.Payment.CreateHostedPaymentSession(r.Context(), &paymentv1.CreateHostedPaymentSessionRequest{
		OrderId:     order.GetId(),
		BuyerUserId: buyerUserID,
		Currency:    "USD",
		ReturnUrl:   orderPageURL,
		CancelUrl:   orderPageURL,
	})
	if err != nil {
		shared.WriteGRPCError(w, r, err)
		return
	}
	hostedPaymentURL := shared.BuildHostedPaymentURL(h.deps.HostedPayment, r, hostedSession)
	if hostedPaymentURL == "" {
		shared.WriteBadRequest(w, r, "hosted payment unavailable")
		return
	}
	shared.Redirect(w, r, hostedPaymentURL, http.StatusSeeOther)
}
