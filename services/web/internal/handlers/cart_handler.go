package handlers

import (
	"net/http"
	"strings"

	cartv1 "refurbished-marketplace/shared/proto/cart/v1"
	ordersv1 "refurbished-marketplace/shared/proto/orders/v1"

	"refurbished-marketplace/services/web/internal/views"

	"github.com/google/uuid"
)

const cartCookieName = "cart_id"

func mapCartView(c *cartv1.Cart) views.CartView {
	items := make([]views.CartItemView, 0, len(c.GetItems()))
	for _, item := range c.GetItems() {
		items = append(items, views.CartItemView{ProductID: item.GetProductId(), Quantity: item.GetQuantity()})
	}
	return views.CartView{CartID: c.GetCartId(), Items: items, CreatedAt: formatTimestamp(c.GetCreatedAt()), UpdatedAt: formatTimestamp(c.GetUpdatedAt())}
}

func cartIDFromRequest(r *http.Request) string {
	if c, err := r.Cookie(cartCookieName); err == nil {
		return strings.TrimSpace(c.Value)
	}
	return ""
}

func cartItemFromForm(r *http.Request) (string, int32, error) {
	if !parseForm(r) {
		return "", 0, errInvalidRequestBody
	}
	quantity, err := parseInt32FormValue(r, "quantity")
	if err != nil {
		return "", 0, err
	}
	return r.FormValue("product_id"), quantity, nil
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

func (h *Handler) handleGetCart(w http.ResponseWriter, r *http.Request) {
	cartID := h.getOrCreateCartID(w, r)
	cart, err := h.cart.GetCart(r.Context(), cartID)
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeHTML(w, r, http.StatusOK, views.CartPage(mapCartView(cart)))
}

func (h *Handler) handleAddCartItem(w http.ResponseWriter, r *http.Request) {
	cartID := h.getOrCreateCartID(w, r)
	productID, quantity, err := cartItemFromForm(r)
	if err != nil || strings.TrimSpace(productID) == "" || quantity <= 0 {
		writeBadRequest(w, r, "invalid request body")
		return
	}
	cart, err := h.cart.AddCartItem(r.Context(), cartID, strings.TrimSpace(productID), quantity)
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeFragment(w, r, http.StatusOK, "#cart", views.CartSection(mapCartView(cart)))
}

func (h *Handler) handleSetCartItemQuantity(w http.ResponseWriter, r *http.Request) {
	cartID := h.getOrCreateCartID(w, r)
	productID, ok := requirePathValue(w, r, "product_id", "invalid product id")
	if !ok {
		return
	}
	_, quantity, err := cartItemFromForm(r)
	if err != nil {
		writeBadRequest(w, r, "invalid request body")
		return
	}
	cart, err := h.cart.SetCartItemQuantity(r.Context(), cartID, productID, quantity)
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeFragment(w, r, http.StatusOK, "#cart", views.CartSection(mapCartView(cart)))
}

func (h *Handler) handleRemoveCartItem(w http.ResponseWriter, r *http.Request) {
	cartID := h.getOrCreateCartID(w, r)
	productID, ok := requirePathValue(w, r, "product_id", "invalid product id")
	if !ok {
		return
	}
	cart, err := h.cart.RemoveCartItem(r.Context(), cartID, productID)
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeFragment(w, r, http.StatusOK, "#cart", views.CartSection(mapCartView(cart)))
}

func (h *Handler) handleCheckoutCart(w http.ResponseWriter, r *http.Request) {
	buyerUserID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	cartID := cartIDFromRequest(r)
	if cartID == "" {
		writeBadRequest(w, r, "empty cart")
		return
	}
	cart, err := h.cart.GetCart(r.Context(), cartID)
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	if len(cart.GetItems()) == 0 {
		writeBadRequest(w, r, "empty cart")
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
		items = append(items, &ordersv1.CreateOrderItem{ProductId: item.GetProductId(), MerchantId: product.GetMerchantId(), Quantity: item.GetQuantity(), UnitPriceCents: product.PriceCents})
	}
	order, err := h.orders.CreateOrder(r.Context(), buyerUserID, items, totalCents)
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	_ = h.cart.ClearCart(r.Context(), cartID)
	h.clearCartCookie(w)
	writeHTML(w, r, http.StatusCreated, views.OrderDetailPage(mapOrderView(order)))
}
