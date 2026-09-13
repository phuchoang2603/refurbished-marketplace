package cart

import (
	"net/http"
	"strings"

	shared "github.com/phuchoang2603/refurbished-marketplace/services/web/internal/handlers/shared"
	cartviews "github.com/phuchoang2603/refurbished-marketplace/services/web/internal/views/cart"

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
	productID, merchantID, productName, quantity, unitPriceCents, err := shared.CartItemStampFromForm(r)
	productID = strings.TrimSpace(productID)
	merchantID = strings.TrimSpace(merchantID)
	productName = strings.TrimSpace(productName)
	if err != nil || productID == "" || merchantID == "" || productName == "" || quantity <= 0 {
		shared.WriteBadRequest(w, r, "invalid request body")
		return
	}
	_, err = h.deps.Cart.AddCartItem(r.Context(), cartID, productID, merchantID, productName, quantity, unitPriceCents)
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
	merchantID = strings.TrimSpace(merchantID)
	if err != nil || merchantID == "" {
		shared.WriteBadRequest(w, r, "invalid request body")
		return
	}
	var productName string
	var unitPriceCents int64
	if quantity > 0 {
		cart, cartErr := h.deps.Cart.GetCart(r.Context(), cartID)
		if cartErr != nil {
			shared.WriteGRPCError(w, r, cartErr)
			return
		}
		for _, item := range cart.GetItems() {
			if item.GetProductId() == productID && item.GetMerchantId() == merchantID {
				productName = item.GetProductName()
				unitPriceCents = item.GetUnitPriceCents()
				break
			}
		}
		if productName == "" {
			shared.WriteBadRequest(w, r, "invalid request body")
			return
		}
	}
	cart, err := h.deps.Cart.SetCartItemQuantity(r.Context(), cartID, productID, merchantID, productName, quantity, unitPriceCents)
	if err != nil {
		shared.WriteGRPCError(w, r, err)
		return
	}
	h.clearCartCookieIfEmpty(w, cart)
	view, err := h.mapCartView(w, r, cart)
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
	cart, err := h.deps.Cart.RemoveCartItems(r.Context(), cartID, []string{productID})
	if err != nil {
		shared.WriteGRPCError(w, r, err)
		return
	}
	h.clearCartCookieIfEmpty(w, cart)
	view, err := h.mapCartView(w, r, cart)
	if err != nil {
		shared.WriteGRPCError(w, r, err)
		return
	}
	shared.WriteFragment(w, r, http.StatusOK, "#cart", cartviews.CartSection(view))
}
