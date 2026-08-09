package cart

import (
	"net/http"
	"strings"

	shared "refurbished-marketplace/services/web/internal/handlers/shared"
	cartviews "refurbished-marketplace/services/web/internal/views/cart"
	productsv1 "refurbished-marketplace/shared/proto/products/v1"

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
	productID = strings.TrimSpace(productID)
	merchantID = strings.TrimSpace(merchantID)
	if err != nil || productID == "" || merchantID == "" || quantity <= 0 {
		shared.WriteBadRequest(w, r, "invalid request body")
		return
	}
	product, ok := h.loadCartProductStamp(w, r, productID, merchantID)
	if !ok {
		return
	}
	_, err = h.deps.Cart.AddCartItem(r.Context(), cartID, productID, merchantID, product.GetName(), quantity, product.GetPriceCents())
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
		product, productOK := h.loadCartProductStamp(w, r, productID, merchantID)
		if !productOK {
			return
		}
		productName = product.GetName()
		unitPriceCents = product.GetPriceCents()
	}
	cart, err := h.deps.Cart.SetCartItemQuantity(r.Context(), cartID, productID, merchantID, productName, quantity, unitPriceCents)
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

func (h *Handler) loadCartProductStamp(w http.ResponseWriter, r *http.Request, productID, merchantID string) (*productsv1.Product, bool) {
	if h.deps.Products == nil {
		shared.WriteBadRequest(w, r, "products unavailable")
		return nil, false
	}
	product, err := h.deps.Products.GetProductByID(r.Context(), productID)
	if err != nil {
		shared.WriteGRPCError(w, r, err)
		return nil, false
	}
	if product == nil || product.GetMerchantId() != merchantID {
		shared.WritePopup(w, r, http.StatusConflict, "Merchant mismatch", "This product no longer matches the selected merchant.")
		return nil, false
	}
	return product, true
}
