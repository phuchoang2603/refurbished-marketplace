package cart

import (
	"net/http"

	shared "refurbished-marketplace/services/web/internal/handlers/shared"
	cartv1 "refurbished-marketplace/shared/proto/cart/v1"
	ordersv1 "refurbished-marketplace/shared/proto/orders/v1"
	paymentv1 "refurbished-marketplace/shared/proto/payment/v1"
)

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

	items, selectedProductIDs, totalCents, err := h.buildCheckoutOrderItems(r, cart, merchantID)
	if err != nil {
		if checkoutErr, ok := err.(*checkoutError); ok {
			checkoutErr.Write(w, r)
			return
		}
		shared.WriteGRPCError(w, r, err)
		return
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
	if err := h.removeCheckedOutItems(w, r, cartID, cart, selectedProductIDs); err != nil {
		shared.WriteGRPCError(w, r, err)
		return
	}

	orderPageURL := shared.OrderPageURLWithConfig(h.deps.HostedPayment, r, order.GetId())
	if orderPageURL == "" {
		shared.WriteBadRequest(w, r, "hosted payment unavailable")
		return
	}
	hostedSession, err := h.deps.Payment.CreateHostedPaymentSession(r.Context(), &paymentv1.CreateHostedPaymentSessionRequest{
		OrderId:     order.GetId(),
		BuyerUserId: buyerUserID,
		Currency:    "USD",
		ReturnUrl:   orderPageURL,
		CancelUrl:   orderPageURL,
	})
	if err != nil {
		shared.WriteGRPCError(w, r, err)
		return
	}
	hostedPaymentURL := shared.BuildHostedPaymentURL(h.deps.HostedPayment, r, hostedSession)
	if hostedPaymentURL == "" {
		shared.WriteBadRequest(w, r, "hosted payment unavailable")
		return
	}
	shared.Redirect(w, r, hostedPaymentURL, http.StatusSeeOther)
}

type checkoutError struct {
	status  int
	title   string
	message string
}

func (e *checkoutError) Error() string {
	return e.message
}

func (e *checkoutError) Write(w http.ResponseWriter, r *http.Request) {
	shared.WritePopup(w, r, e.status, e.title, e.message)
}

func (h *Handler) buildCheckoutOrderItems(r *http.Request, cart *cartv1.Cart, merchantID string) ([]*ordersv1.CreateOrderItem, []string, int64, error) {
	items := make([]*ordersv1.CreateOrderItem, 0, len(cart.GetItems()))
	selectedProductIDs := make([]string, 0, len(cart.GetItems()))
	var totalCents int64
	for _, item := range cart.GetItems() {
		if item.GetMerchantId() != merchantID {
			continue
		}
		product, err := h.deps.Products.GetProductByID(r.Context(), item.GetProductId())
		if err != nil {
			return nil, nil, 0, err
		}
		if product.GetMerchantId() != merchantID {
			return nil, nil, 0, &checkoutError{
				status:  http.StatusConflict,
				title:   "Merchant mismatch",
				message: "One or more cart items no longer match the selected merchant.",
			}
		}
		lineTotal := product.PriceCents * int64(item.GetQuantity())
		totalCents += lineTotal
		items = append(items, &ordersv1.CreateOrderItem{
			ProductId:      item.GetProductId(),
			Quantity:       item.GetQuantity(),
			UnitPriceCents: product.PriceCents,
		})
		selectedProductIDs = append(selectedProductIDs, item.GetProductId())
	}
	return items, selectedProductIDs, totalCents, nil
}

func (h *Handler) removeCheckedOutItems(w http.ResponseWriter, r *http.Request, cartID string, cart *cartv1.Cart, selectedProductIDs []string) error {
	remainingItems := len(cart.GetItems())
	for _, productID := range selectedProductIDs {
		updatedCart, err := h.deps.Cart.RemoveCartItem(r.Context(), cartID, productID)
		if err != nil {
			return err
		}
		remainingItems = len(updatedCart.GetItems())
	}
	if remainingItems == 0 {
		h.clearCartCookie(w)
	}
	return nil
}
