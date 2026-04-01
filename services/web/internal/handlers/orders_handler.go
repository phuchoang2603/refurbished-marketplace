package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	webAuth "refurbished-marketplace/services/web/internal/auth"
)

type createOrderRequest struct {
	ProductID string `json:"product_id"`
	Quantity  int32  `json:"quantity"`
}

type orderResponse struct {
	ID          string `json:"id"`
	BuyerUserID string `json:"buyer_user_id"`
	ProductID   string `json:"product_id"`
	Quantity    int32  `json:"quantity"`
	Status      string `json:"status"`
	TotalCents  int64  `json:"total_cents"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func mapOrder(id, buyerUserID, productID string, quantity int32, status string, totalCents int64, createdAt, updatedAt string) orderResponse {
	return orderResponse{ID: id, BuyerUserID: buyerUserID, ProductID: productID, Quantity: quantity, Status: status, TotalCents: totalCents, CreatedAt: createdAt, UpdatedAt: updatedAt}
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
	if strings.TrimSpace(req.ProductID) == "" || req.Quantity <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	product, err := h.products.GetProductByID(r.Context(), req.ProductID)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	totalCents := product.PriceCents * int64(req.Quantity)
	order, err := h.orders.CreateOrder(r.Context(), buyerUserID, req.ProductID, req.Quantity, totalCents)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, mapOrder(order.Id, order.BuyerUserId, order.ProductId, order.Quantity, order.Status, order.TotalCents, order.CreatedAt.AsTime().UTC().Format("2006-01-02T15:04:05Z07:00"), order.UpdatedAt.AsTime().UTC().Format("2006-01-02T15:04:05Z07:00")))
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

	writeJSON(w, http.StatusOK, mapOrder(order.Id, order.BuyerUserId, order.ProductId, order.Quantity, order.Status, order.TotalCents, order.CreatedAt.AsTime().UTC().Format("2006-01-02T15:04:05Z07:00"), order.UpdatedAt.AsTime().UTC().Format("2006-01-02T15:04:05Z07:00")))
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
		items = append(items, mapOrder(order.Id, order.BuyerUserId, order.ProductId, order.Quantity, order.Status, order.TotalCents, order.CreatedAt.AsTime().UTC().Format("2006-01-02T15:04:05Z07:00"), order.UpdatedAt.AsTime().UTC().Format("2006-01-02T15:04:05Z07:00")))
	}

	writeJSON(w, http.StatusOK, map[string]any{"orders": items})
}

func (h *Handler) handleConfirmOrder(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid order id"})
		return
	}

	order, err := h.orders.UpdateOrderStatus(r.Context(), id, "CONFIRMED")
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, mapOrder(order.Id, order.BuyerUserId, order.ProductId, order.Quantity, order.Status, order.TotalCents, order.CreatedAt.AsTime().UTC().Format("2006-01-02T15:04:05Z07:00"), order.UpdatedAt.AsTime().UTC().Format("2006-01-02T15:04:05Z07:00")))
}

func (h *Handler) handleCancelOrder(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid order id"})
		return
	}

	order, err := h.orders.UpdateOrderStatus(r.Context(), id, "CANCELED")
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, mapOrder(order.Id, order.BuyerUserId, order.ProductId, order.Quantity, order.Status, order.TotalCents, order.CreatedAt.AsTime().UTC().Format("2006-01-02T15:04:05Z07:00"), order.UpdatedAt.AsTime().UTC().Format("2006-01-02T15:04:05Z07:00")))
}
