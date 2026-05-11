package handlers

import (
	"net/http"
	"strings"

	"refurbished-marketplace/services/web/internal/views"
	ordersv1 "refurbished-marketplace/shared/proto/orders/v1"
)

func mapOrderView(order *ordersv1.Order) views.OrderView {
	items := make([]views.OrderItemView, 0, len(order.GetItems()))
	for _, item := range order.GetItems() {
		items = append(
			items,
			views.OrderItemView{
				ID:             item.GetId(),
				OrderID:        item.GetOrderId(),
				ProductID:      item.GetProductId(),
				Quantity:       item.GetQuantity(),
				UnitPriceCents: item.GetUnitPriceCents(),
				LineTotalCents: item.GetLineTotalCents(),
				CreatedAt:      formatTimestamp(item.GetCreatedAt()),
			},
		)
	}
	return views.OrderView{
		ID:          order.GetId(),
		BuyerUserID: order.GetBuyerUserId(),
		Status:      order.GetStatus().String(),
		TotalCents:  order.GetTotalCents(),
		Items:       items,
		CreatedAt:   formatTimestamp(order.GetCreatedAt()),
		UpdatedAt:   formatTimestamp(order.GetUpdatedAt()),
	}
}

func createOrderItemFromForm(r *http.Request) (string, int32, error) {
	if !parseForm(r) {
		return "", 0, errInvalidRequestBody
	}
	quantity, err := parseInt32FormValue(r, "quantity")
	if err != nil {
		return "", 0, err
	}
	return r.FormValue("product_id"), quantity, nil
}

func (h *Handler) buildCreateOrderItem(w http.ResponseWriter, r *http.Request, productID string, quantity int32) (*ordersv1.CreateOrderItem, int64, bool) {
	productID = strings.TrimSpace(productID)
	if productID == "" || quantity <= 0 {
		writeBadRequest(w, r, "invalid request body")
		return nil, 0, false
	}

	product, err := h.products.GetProductByID(r.Context(), productID)
	if err != nil {
		writeGRPCError(w, r, err)
		return nil, 0, false
	}

	item := &ordersv1.CreateOrderItem{ProductId: productID, MerchantId: product.GetMerchantId(), Quantity: quantity, UnitPriceCents: product.PriceCents}
	return item, product.PriceCents * int64(quantity), true
}

func (h *Handler) handleCreateOrder(w http.ResponseWriter, r *http.Request) {
	buyerUserID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	productID, quantity, err := createOrderItemFromForm(r)
	if err != nil {
		writeBadRequest(w, r, "invalid request body")
		return
	}

	item, totalCents, ok := h.buildCreateOrderItem(w, r, productID, quantity)
	if !ok {
		return
	}

	order, err := h.orders.CreateOrder(r.Context(), buyerUserID, []*ordersv1.CreateOrderItem{item}, totalCents)
	if err != nil {
		writeGRPCError(w, r, err)
		return
	}

	writeHTML(w, r, http.StatusCreated, views.OrderDetailPage(mapOrderView(order)))
}

func (h *Handler) handleGetOrderByID(w http.ResponseWriter, r *http.Request) {
	buyerUserID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	id, ok := requirePathValue(w, r, "id", "invalid order id")
	if !ok {
		return
	}

	order, err := h.orders.GetOrderByID(r.Context(), id)
	if err != nil {
		writeGRPCError(w, r, err)
		return
	}
	if order.GetBuyerUserId() != buyerUserID {
		writeHTML(w, r, http.StatusForbidden, views.MessagePage("Forbidden", "order does not belong to the current user"))
		return
	}

	writeHTML(w, r, http.StatusOK, views.OrderDetailPage(mapOrderView(order)))
}

func (h *Handler) handleListOrdersByBuyer(w http.ResponseWriter, r *http.Request) {
	buyerUserID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	resp, err := h.orders.ListOrdersByBuyer(r.Context(), buyerUserID, 20, 0)
	if err != nil {
		writeGRPCError(w, r, err)
		return
	}

	items := make([]views.OrderView, 0, len(resp.Orders))
	for _, order := range resp.Orders {
		items = append(items, mapOrderView(order))
	}

	writeHTML(w, r, http.StatusOK, views.OrdersPage(items))
}
