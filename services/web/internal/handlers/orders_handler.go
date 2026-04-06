package handlers

import (
	"net/http"
	"strings"

	ordersv1 "refurbished-marketplace/shared/proto/orders/v1"
)

type createOrderRequest struct {
	Items []createOrderItemRequest `json:"items"`
}

type createOrderItemRequest struct {
	ProductID string `json:"product_id"`
	Quantity  int32  `json:"quantity"`
}

type orderResponse struct {
	ID          string              `json:"id"`
	BuyerUserID string              `json:"buyer_user_id"`
	Status      string              `json:"status"`
	TotalCents  int64               `json:"total_cents"`
	Items       []orderItemResponse `json:"items"`
	CreatedAt   string              `json:"created_at"`
	UpdatedAt   string              `json:"updated_at"`
}

type orderItemResponse struct {
	ID             string `json:"id"`
	OrderID        string `json:"order_id"`
	ProductID      string `json:"product_id"`
	Quantity       int32  `json:"quantity"`
	UnitPriceCents int64  `json:"unit_price_cents"`
	LineTotalCents int64  `json:"line_total_cents"`
	CreatedAt      string `json:"created_at"`
}

func mapProtoOrderItem(item *ordersv1.OrderItem) orderItemResponse {
	return orderItemResponse{
		ID:             item.GetId(),
		OrderID:        item.GetOrderId(),
		ProductID:      item.GetProductId(),
		Quantity:       item.GetQuantity(),
		UnitPriceCents: item.GetUnitPriceCents(),
		LineTotalCents: item.GetLineTotalCents(),
		CreatedAt:      formatTimestamp(item.GetCreatedAt()),
	}
}

func mapProtoOrder(order *ordersv1.Order) orderResponse {
	items := make([]orderItemResponse, 0, len(order.GetItems()))
	for _, item := range order.GetItems() {
		items = append(items, mapProtoOrderItem(item))
	}
	return orderResponse{
		ID:          order.GetId(),
		BuyerUserID: order.GetBuyerUserId(),
		Status:      order.GetStatus().String(),
		TotalCents:  order.GetTotalCents(),
		Items:       items,
		CreatedAt:   formatTimestamp(order.GetCreatedAt()),
		UpdatedAt:   formatTimestamp(order.GetUpdatedAt()),
	}
}

func (h *Handler) buildCreateOrderItems(w http.ResponseWriter, r *http.Request, reqItems []createOrderItemRequest) ([]*ordersv1.CreateOrderItem, int64, bool) {
	items := make([]*ordersv1.CreateOrderItem, 0, len(reqItems))
	var totalCents int64
	for _, item := range reqItems {
		productID := strings.TrimSpace(item.ProductID)
		if productID == "" || item.Quantity <= 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
			return nil, 0, false
		}

		product, err := h.products.GetProductByID(r.Context(), productID)
		if err != nil {
			writeGRPCError(w, err)
			return nil, 0, false
		}

		lineTotal := product.PriceCents * int64(item.Quantity)
		totalCents += lineTotal
		items = append(items, &ordersv1.CreateOrderItem{ProductId: productID, Quantity: item.Quantity, UnitPriceCents: product.PriceCents})
	}

	return items, totalCents, true
}

func (h *Handler) handleCreateOrder(w http.ResponseWriter, r *http.Request) {
	buyerUserID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	var req createOrderRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if len(req.Items) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	items, totalCents, ok := h.buildCreateOrderItems(w, r, req.Items)
	if !ok {
		return
	}

	order, err := h.orders.CreateOrder(r.Context(), buyerUserID, items, totalCents)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, mapProtoOrder(order))
}

func (h *Handler) handleGetOrderByID(w http.ResponseWriter, r *http.Request) {
	id, ok := requirePathValue(w, r, "id", "invalid order id")
	if !ok {
		return
	}

	order, err := h.orders.GetOrderByID(r.Context(), id)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, mapProtoOrder(order))
}

func (h *Handler) handleListOrdersByBuyer(w http.ResponseWriter, r *http.Request) {
	buyerUserID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	resp, err := h.orders.ListOrdersByBuyer(r.Context(), buyerUserID, 20, 0)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	items := make([]orderResponse, 0, len(resp.Orders))
	for _, order := range resp.Orders {
		items = append(items, mapProtoOrder(order))
	}

	writeJSON(w, http.StatusOK, map[string]any{"orders": items})
}
