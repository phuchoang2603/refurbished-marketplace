package cart

import (
	"net/http"
	"strings"

	shared "refurbished-marketplace/services/web/internal/handlers/shared"
	cartviews "refurbished-marketplace/services/web/internal/views/cart"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) RegisterActions(r chi.Router) {
	r.Post("/cart/items", h.handleAddCartItem)
	r.Post("/cart/items/{product_id}/quantity", h.handleSetCartItemQuantity)
	r.Post("/cart/items/{product_id}/remove", h.handleRemoveCartItem)
}

func (h *Handler) RegisterProtectedActions(r chi.Router) {
	r.Post("/cart/checkout", h.handleCheckoutCart)
}

func (h *Handler) handleAddCartItem(w http.ResponseWriter, r *http.Request) {
	cartID := h.getOrCreateCartID(w, r)
	productID, merchantID, quantity, err := shared.ProductQuantityMerchantFromForm(r)
	if err != nil || strings.TrimSpace(productID) == "" || strings.TrimSpace(merchantID) == "" || quantity <= 0 {
		shared.WriteBadRequest(w, r, "invalid request body")
		return
	}
	_, err = h.deps.Cart.AddCartItem(r.Context(), cartID, strings.TrimSpace(productID), strings.TrimSpace(merchantID), quantity)
	if err != nil {
		shared.WriteGRPCError(w, r, err)
		return
	}
	shared.Redirect(w, r, "/cart", http.StatusSeeOther)
}

func (h *Handler) handleSetCartItemQuantity(w http.ResponseWriter, r *http.Request) {
	cartID := h.getOrCreateCartID(w, r)
	productID, ok := shared.RequirePathValue(w, r, "product_id", "invalid product id")
	if !ok {
		return
	}
	_, merchantID, quantity, err := shared.ProductQuantityMerchantFromForm(r)
	if err != nil || strings.TrimSpace(merchantID) == "" {
		shared.WriteBadRequest(w, r, "invalid request body")
		return
	}
	cart, err := h.deps.Cart.SetCartItemQuantity(r.Context(), cartID, productID, strings.TrimSpace(merchantID), quantity)
	if err != nil {
		shared.WriteGRPCError(w, r, err)
		return
	}
	view, err := h.mapCartView(r.Context(), cart)
	if err != nil {
		shared.WriteGRPCError(w, r, err)
		return
	}
	shared.WriteFragment(w, r, http.StatusOK, "#cart", cartviews.CartSection(view))
}

func (h *Handler) handleRemoveCartItem(w http.ResponseWriter, r *http.Request) {
	cartID := h.getOrCreateCartID(w, r)
	productID, ok := shared.RequirePathValue(w, r, "product_id", "invalid product id")
	if !ok {
		return
	}
	cart, err := h.deps.Cart.RemoveCartItem(r.Context(), cartID, productID)
	if err != nil {
		shared.WriteGRPCError(w, r, err)
		return
	}
	view, err := h.mapCartView(r.Context(), cart)
	if err != nil {
		shared.WriteGRPCError(w, r, err)
		return
	}
	shared.WriteFragment(w, r, http.StatusOK, "#cart", cartviews.CartSection(view))
}
