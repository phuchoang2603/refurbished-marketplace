package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	webAuth "refurbished-marketplace/services/web/internal/auth"
	cartv1 "refurbished-marketplace/shared/proto/cart/v1"
	ordersv1 "refurbished-marketplace/shared/proto/orders/v1"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const cartCookieName = "cart_id"

type cartResponse struct {
	CartID    string             `json:"cart_id"`
	Items     []cartItemResponse `json:"items"`
	CreatedAt string             `json:"created_at"`
	UpdatedAt string             `json:"updated_at"`
}

type cartItemResponse struct {
	ProductID string `json:"product_id"`
	Quantity  int32  `json:"quantity"`
}

type cartItemRequest struct {
	ProductID string `json:"product_id"`
	Quantity  int32  `json:"quantity"`
}

func timestampString(ts *timestamppb.Timestamp) string {
	if ts == nil {
		return ""
	}
	return ts.AsTime().UTC().Format("2006-01-02T15:04:05Z07:00")
}

func mapCart(c *cartv1.Cart) cartResponse {
	items := make([]cartItemResponse, 0, len(c.GetItems()))
	for _, item := range c.GetItems() {
		items = append(items, cartItemResponse{ProductID: item.GetProductId(), Quantity: item.GetQuantity()})
	}
	return cartResponse{CartID: c.GetCartId(), Items: items, CreatedAt: timestampString(c.GetCreatedAt()), UpdatedAt: timestampString(c.GetUpdatedAt())}
}

func cartIDFromRequest(r *http.Request) string {
	if c, err := r.Cookie(cartCookieName); err == nil {
		return strings.TrimSpace(c.Value)
	}
	return ""
}

func (h *Handler) getOrCreateCartID(w http.ResponseWriter, r *http.Request) string {
	if id := cartIDFromRequest(r); id != "" {
		return id
	}
	id := uuid.NewString()
	http.SetCookie(w, &http.Cookie{Name: cartCookieName, Value: id, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode})
	return id
}

func (h *Handler) clearCartCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: cartCookieName, Value: "", Path: "/", HttpOnly: true, MaxAge: -1, SameSite: http.SameSiteLaxMode})
}

func decodeCartItemRequest(r *http.Request) (cartItemRequest, error) {
	var req cartItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return cartItemRequest{}, err
	}
	return req, nil
}

func (h *Handler) handleGetCart(w http.ResponseWriter, r *http.Request) {
	cartID := h.getOrCreateCartID(w, r)
	cart, err := h.cart.GetCart(r.Context(), cartID)
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, mapCart(cart))
}

func (h *Handler) handleAddCartItem(w http.ResponseWriter, r *http.Request) {
	cartID := h.getOrCreateCartID(w, r)
	req, err := decodeCartItemRequest(r)
	if err != nil || strings.TrimSpace(req.ProductID) == "" || req.Quantity <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	cart, err := h.cart.AddCartItem(r.Context(), cartID, strings.TrimSpace(req.ProductID), req.Quantity)
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, mapCart(cart))
}

func (h *Handler) handleSetCartItemQuantity(w http.ResponseWriter, r *http.Request) {
	cartID := h.getOrCreateCartID(w, r)
	productID := strings.TrimSpace(r.PathValue("product_id"))
	if productID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid product id"})
		return
	}
	req, err := decodeCartItemRequest(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	cart, err := h.cart.SetCartItemQuantity(r.Context(), cartID, productID, req.Quantity)
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, mapCart(cart))
}

func (h *Handler) handleRemoveCartItem(w http.ResponseWriter, r *http.Request) {
	cartID := h.getOrCreateCartID(w, r)
	productID := strings.TrimSpace(r.PathValue("product_id"))
	if productID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid product id"})
		return
	}
	cart, err := h.cart.RemoveCartItem(r.Context(), cartID, productID)
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, mapCart(cart))
}

func (h *Handler) handleCheckoutCart(w http.ResponseWriter, r *http.Request) {
	buyerUserID, ok := webAuth.UserIDFromContext(r.Context())
	if !ok || buyerUserID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	cartID := cartIDFromRequest(r)
	if cartID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "empty cart"})
		return
	}
	cart, err := h.cart.GetCart(r.Context(), cartID)
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	if len(cart.GetItems()) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "empty cart"})
		return
	}
	items := make([]*ordersv1.CreateOrderItem, 0, len(cart.GetItems()))
	var totalCents int64
	for _, item := range cart.GetItems() {
		product, err := h.products.GetProductByID(r.Context(), item.GetProductId())
		if err != nil {
			writeGRPCError(w, err)
			return
		}
		lineTotal := product.PriceCents * int64(item.GetQuantity())
		totalCents += lineTotal
		items = append(items, &ordersv1.CreateOrderItem{ProductId: item.GetProductId(), Quantity: item.GetQuantity(), UnitPriceCents: product.PriceCents})
	}
	order, err := h.orders.CreateOrder(r.Context(), buyerUserID, items, totalCents)
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	_ = h.cart.ClearCart(r.Context(), cartID)
	h.clearCartCookie(w)
	orderItems := make([]orderItemResponse, 0, len(order.Items))
	for _, item := range order.Items {
		orderItems = append(orderItems, mapOrderItem(item.Id, item.OrderId, item.ProductId, item.Quantity, item.UnitPriceCents, item.LineTotalCents, item.CreatedAt.AsTime().UTC().Format("2006-01-02T15:04:05Z07:00")))
	}
	writeJSON(w, http.StatusCreated, mapOrder(order.Id, order.BuyerUserId, order.Status.String(), order.TotalCents, orderItems, order.CreatedAt.AsTime().UTC().Format("2006-01-02T15:04:05Z07:00"), order.UpdatedAt.AsTime().UTC().Format("2006-01-02T15:04:05Z07:00")))
}
