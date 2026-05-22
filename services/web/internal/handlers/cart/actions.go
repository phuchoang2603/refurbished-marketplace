package cart

import (
	"net/http"
	"strings"

	shared "refurbished-marketplace/services/web/internal/handlers/shared"
	cartviews "refurbished-marketplace/services/web/internal/views/cart"
	ordersv1 "refurbished-marketplace/shared/proto/orders/v1"

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

func (h *Handler) handleCheckoutCart(w http.ResponseWriter, r *http.Request) {
	buyerUserID, ok := shared.RequireUserID(w, r)
	if !ok {
		return
	}
	merchantID, err := shared.MerchantIDFromForm(r)
	if err != nil {
		shared.WriteBadRequest(w, r, "invalid request body")
		return
	}
	cartID := cartIDFromRequest(r)
	if cartID == "" {
		shared.WriteBadRequest(w, r, "empty cart")
		return
	}
	cart, err := h.deps.Cart.GetCart(r.Context(), cartID)
	if err != nil {
		shared.WriteGRPCError(w, r, err)
		return
	}
	if len(cart.GetItems()) == 0 {
		shared.WriteBadRequest(w, r, "empty cart")
		return
	}
	items := make([]*ordersv1.CreateOrderItem, 0, len(cart.GetItems()))
	selectedProductIDs := make([]string, 0, len(cart.GetItems()))
	var totalCents int64
	for _, item := range cart.GetItems() {
		if item.GetMerchantId() != merchantID {
			continue
		}
		product, err := h.deps.Products.GetProductByID(r.Context(), item.GetProductId())
		if err != nil {
			shared.WriteGRPCError(w, r, err)
			return
		}
		lineTotal := product.PriceCents * int64(item.GetQuantity())
		totalCents += lineTotal
		if product.GetMerchantId() != merchantID {
			shared.WritePopup(w, r, http.StatusConflict, "Merchant mismatch", "One or more cart items no longer match the selected merchant.")
			return
		}
		items = append(items, &ordersv1.CreateOrderItem{ProductId: item.GetProductId(), Quantity: item.GetQuantity(), UnitPriceCents: product.PriceCents})
		selectedProductIDs = append(selectedProductIDs, item.GetProductId())
	}
	if len(items) == 0 {
		shared.WriteBadRequest(w, r, "no items for selected merchant")
		return
	}
	order, err := h.deps.Orders.CreateOrder(r.Context(), buyerUserID, merchantID, items, totalCents)
	if err != nil {
		shared.WriteGRPCError(w, r, err)
		return
	}
	remainingItems := len(cart.GetItems())
	for _, productID := range selectedProductIDs {
		updatedCart, err := h.deps.Cart.RemoveCartItem(r.Context(), cartID, productID)
		if err != nil {
			shared.WriteGRPCError(w, r, err)
			return
		}
		remainingItems = len(updatedCart.GetItems())
	}
	if remainingItems == 0 {
		h.clearCartCookie(w)
	}
	shared.Redirect(w, r, "/orders/"+order.GetId(), http.StatusSeeOther)
}
