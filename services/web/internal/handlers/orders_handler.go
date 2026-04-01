package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	webAuth "refurbished-marketplace/services/web/internal/auth"
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

func mapOrderItem(id, orderID, productID string, quantity int32, unitPriceCents, lineTotalCents int64, createdAt string) orderItemResponse {
	return orderItemResponse{ID: id, OrderID: orderID, ProductID: productID, Quantity: quantity, UnitPriceCents: unitPriceCents, LineTotalCents: lineTotalCents, CreatedAt: createdAt}
}

func mapOrder(id, buyerUserID, status string, totalCents int64, items []orderItemResponse, createdAt, updatedAt string) orderResponse {
	return orderResponse{ID: id, BuyerUserID: buyerUserID, Status: status, TotalCents: totalCents, Items: items, CreatedAt: createdAt, UpdatedAt: updatedAt}
}

func (h *Handler) handleCreateOrder(w http.ResponseWriter, r *http.Request) {
	buyerUserID, ok := webAuth.UserIDFromContext(r.Context())
	if !ok || buyerUserID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var req createOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if len(req.Items) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	items := make([]*ordersv1.CreateOrderItem, 0, len(req.Items))
	itemResponses := make([]orderItemResponse, 0, len(req.Items))
	var totalCents int64
	for _, item := range req.Items {
		productID := strings.TrimSpace(item.ProductID)
		if productID == "" || item.Quantity <= 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
			return
		}

		product, err := h.products.GetProductByID(r.Context(), productID)
		if err != nil {
			writeGRPCError(w, err)
			return
		}

		lineTotal := product.PriceCents * int64(item.Quantity)
		totalCents += lineTotal
		items = append(items, &ordersv1.CreateOrderItem{ProductId: productID, Quantity: item.Quantity, UnitPriceCents: product.PriceCents})
		itemResponses = append(itemResponses, mapOrderItem("", "", productID, item.Quantity, product.PriceCents, lineTotal, ""))
	}

	order, err := h.orders.CreateOrder(r.Context(), buyerUserID, items, totalCents)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	itemResponses = itemResponses[:0]
	for _, item := range order.Items {
		itemResponses = append(itemResponses, mapOrderItem(item.Id, item.OrderId, item.ProductId, item.Quantity, item.UnitPriceCents, item.LineTotalCents, item.CreatedAt.AsTime().UTC().Format("2006-01-02T15:04:05Z07:00")))
	}

	writeJSON(w, http.StatusCreated, mapOrder(order.Id, order.BuyerUserId, order.Status.String(), order.TotalCents, itemResponses, order.CreatedAt.AsTime().UTC().Format("2006-01-02T15:04:05Z07:00"), order.UpdatedAt.AsTime().UTC().Format("2006-01-02T15:04:05Z07:00")))
}

func (h *Handler) handleGetOrderByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid order id"})
		return
	}

	order, err := h.orders.GetOrderByID(r.Context(), id)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	items := make([]orderItemResponse, 0, len(order.Items))
	for _, item := range order.Items {
		items = append(items, mapOrderItem(item.Id, item.OrderId, item.ProductId, item.Quantity, item.UnitPriceCents, item.LineTotalCents, item.CreatedAt.AsTime().UTC().Format("2006-01-02T15:04:05Z07:00")))
	}

	writeJSON(w, http.StatusOK, mapOrder(order.Id, order.BuyerUserId, order.Status.String(), order.TotalCents, items, order.CreatedAt.AsTime().UTC().Format("2006-01-02T15:04:05Z07:00"), order.UpdatedAt.AsTime().UTC().Format("2006-01-02T15:04:05Z07:00")))
}

func (h *Handler) handleListOrdersByBuyer(w http.ResponseWriter, r *http.Request) {
	buyerUserID, ok := webAuth.UserIDFromContext(r.Context())
	if !ok || buyerUserID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	resp, err := h.orders.ListOrdersByBuyer(r.Context(), buyerUserID, 20, 0)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	items := make([]orderResponse, 0, len(resp.Orders))
	for _, order := range resp.Orders {
		orderItems := make([]orderItemResponse, 0, len(order.Items))
		for _, item := range order.Items {
			orderItems = append(orderItems, mapOrderItem(item.Id, item.OrderId, item.ProductId, item.Quantity, item.UnitPriceCents, item.LineTotalCents, item.CreatedAt.AsTime().UTC().Format("2006-01-02T15:04:05Z07:00")))
		}
		items = append(items, mapOrder(order.Id, order.BuyerUserId, order.Status.String(), order.TotalCents, orderItems, order.CreatedAt.AsTime().UTC().Format("2006-01-02T15:04:05Z07:00"), order.UpdatedAt.AsTime().UTC().Format("2006-01-02T15:04:05Z07:00")))
	}

	writeJSON(w, http.StatusOK, map[string]any{"orders": items})
}
