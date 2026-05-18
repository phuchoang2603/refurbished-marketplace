package orders

import (
	"net/http"
	"strings"

	shared "refurbished-marketplace/services/web/internal/handlers/shared"
	orderviews "refurbished-marketplace/services/web/internal/views/orders"
	ordersv1 "refurbished-marketplace/shared/proto/orders/v1"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) RegisterActions(r chi.Router) {
	r.Post("/orders", h.handleCreateOrder)
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

	order, err := h.deps.Orders.CreateOrder(r.Context(), buyerUserID, merchantID, []*ordersv1.CreateOrderItem{item}, totalCents)
	if err != nil {
		shared.WriteGRPCError(w, r, err)
		return
	}

	shared.WriteHTML(w, r, http.StatusCreated, orderviews.OrderDetailPage(OrderToView(order)))
}
