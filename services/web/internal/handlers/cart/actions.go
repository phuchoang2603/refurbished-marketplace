package cart

import (
	"net/http"
	"strings"

	orderhandlers "refurbished-marketplace/services/web/internal/handlers/orders"
	shared "refurbished-marketplace/services/web/internal/handlers/shared"
	cartviews "refurbished-marketplace/services/web/internal/views/cart"
	orderviews "refurbished-marketplace/services/web/internal/views/orders"
	ordersv1 "refurbished-marketplace/shared/proto/orders/v1"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) RegisterActions(r chi.Router) {
	r.Post("/cart/items", h.handleAddCartItem)
	r.Patch("/cart/items/{product_id}", h.handleSetCartItemQuantity)
	r.Delete("/cart/items/{product_id}", h.handleRemoveCartItem)
}

func (h *Handler) RegisterProtectedActions(r chi.Router) {
	r.Post("/cart/checkout", h.handleCheckoutCart)
}

func (h *Handler) handleAddCartItem(w http.ResponseWriter, r *http.Request) {
	cartID := h.getOrCreateCartID(w, r)
	productID, quantity, err := shared.ProductQuantityFromForm(r)
	if err != nil || strings.TrimSpace(productID) == "" || quantity <= 0 {
		shared.WriteBadRequest(w, r, "invalid request body")
		return
	}
	cart, err := h.deps.Cart.AddCartItem(r.Context(), cartID, strings.TrimSpace(productID), quantity)
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

func (h *Handler) handleSetCartItemQuantity(w http.ResponseWriter, r *http.Request) {
	cartID := h.getOrCreateCartID(w, r)
	productID, ok := shared.RequirePathValue(w, r, "product_id", "invalid product id")
	if !ok {
		return
	}
	_, quantity, err := shared.ProductQuantityFromForm(r)
	if err != nil {
		shared.WriteBadRequest(w, r, "invalid request body")
		return
	}
	cart, err := h.deps.Cart.SetCartItemQuantity(r.Context(), cartID, productID, quantity)
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
	var totalCents int64
	merchantID := ""
	for _, item := range cart.GetItems() {
		product, err := h.deps.Products.GetProductByID(r.Context(), item.GetProductId())
		if err != nil {
			shared.WriteGRPCError(w, r, err)
			return
		}
		lineTotal := product.PriceCents * int64(item.GetQuantity())
		totalCents += lineTotal
		if merchantID == "" {
			merchantID = product.GetMerchantId()
		} else if merchantID != product.GetMerchantId() {
			shared.WritePopup(w, r, http.StatusConflict, "Cart requires split checkout", "This cart contains items from multiple merchants. Please keep checkout to a single merchant for now.")
			return
		}
		items = append(items, &ordersv1.CreateOrderItem{ProductId: item.GetProductId(), Quantity: item.GetQuantity(), UnitPriceCents: product.PriceCents})
	}
	order, err := h.deps.Orders.CreateOrder(r.Context(), buyerUserID, merchantID, items, totalCents)
	if err != nil {
		shared.WriteGRPCError(w, r, err)
		return
	}
	_ = h.deps.Cart.ClearCart(r.Context(), cartID)
	h.clearCartCookie(w)
	shared.WriteHTML(w, r, http.StatusCreated, orderviews.OrderDetailPage(orderhandlers.OrderToView(order)))
}
