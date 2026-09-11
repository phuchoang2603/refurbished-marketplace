package cart

import (
	"net/http"
	"strings"

	webAuth "github.com/phuchoang2603/refurbished-marketplace/services/web/internal/auth"
	shared "github.com/phuchoang2603/refurbished-marketplace/services/web/internal/handlers/shared"
	cartv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/cart/v1"
	ordersv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/orders/v1"
	paymentv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/payment/v1"
	productsv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/products/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// maxCheckoutProductLines mirrors services/products GetProductsByIDs max (maxProductsByIDs = 100).
// Enforced at the web edge so buyers get a clear message instead of an opaque InvalidArgument.
const maxCheckoutProductLines = 100

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
	intentKey, err := shared.CheckoutIntentKeyFromForm(r)
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
	shipping, ok := shippingAddressFromForm(r)
	if !ok {
		shared.WritePopup(w, r, http.StatusBadRequest, "Shipping required", "Enter a shipping address with line 1, city, postal code, and country.")
		return
	}

	items, productNames, totalCents, err := h.buildCheckoutOrderItems(r, cart, merchantID)
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

	order, err := h.deps.Orders.CreateOrder(r.Context(), buyerUserID, merchantID, items, totalCents, intentKey)
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.FailedPrecondition {
			rotateCheckoutIntent(w, r, cartID, merchantID)
		}
		shared.WriteGRPCError(w, r, err)
		return
	}
	if err := shared.HoldStockForOrder(r.Context(), h.deps.Products, h.deps.Orders, order); err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.FailedPrecondition {
			rotateCheckoutIntent(w, r, cartID, merchantID)
		}
		shared.WriteGRPCError(w, r, err)
		return
	}

	orderPageURL := shared.OrderPageURLWithConfig(h.deps.HostedPayment, r, order.GetId())
	if orderPageURL == "" {
		shared.WriteBadRequest(w, r, "hosted payment unavailable")
		return
	}
	lineItems := make([]*paymentv1.HostedPaymentLineItem, 0, len(items))
	for _, item := range items {
		lineItems = append(lineItems, &paymentv1.HostedPaymentLineItem{
			ProductId:      item.GetProductId(),
			Name:           productNames[item.GetProductId()],
			Quantity:       item.GetQuantity(),
			UnitPriceCents: item.GetUnitPriceCents(),
		})
	}
	hostedSession, err := h.deps.Payment.CreateHostedPaymentSession(r.Context(), &paymentv1.CreateHostedPaymentSessionRequest{
		OrderId: order.GetId(),
		Buyer: &paymentv1.PartySnapshot{
			Id:    buyerUserID,
			Email: webAuth.EmailFromContext(r.Context()),
		},
		Merchant:        &paymentv1.PartySnapshot{Id: merchantID},
		TotalCents:      totalCents,
		Currency:        "USD",
		ReturnUrl:       orderPageURL,
		ShippingAddress: shipping,
		Items:           lineItems,
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

func (h *Handler) buildCheckoutOrderItems(r *http.Request, cart *cartv1.Cart, merchantID string) ([]*ordersv1.CreateOrderItem, map[string]string, int64, error) {
	selected := make([]*cartv1.CartItem, 0, len(cart.GetItems()))
	selectedProductIDs := make([]string, 0, len(cart.GetItems()))
	for _, item := range cart.GetItems() {
		if item.GetMerchantId() != merchantID {
			continue
		}
		selected = append(selected, item)
		selectedProductIDs = append(selectedProductIDs, item.GetProductId())
	}
	if len(selected) == 0 {
		return nil, nil, 0, nil
	}
	if len(selectedProductIDs) > maxCheckoutProductLines {
		return nil, nil, 0, &checkoutError{
			status:  http.StatusBadRequest,
			title:   "Too many items",
			message: "This merchant group has too many product lines for a single checkout (maximum 100). Remove some items and try again.",
		}
	}

	if h.deps.Products == nil {
		return nil, nil, 0, &checkoutError{
			status:  http.StatusServiceUnavailable,
			title:   "Products unavailable",
			message: "Product details could not be loaded for checkout.",
		}
	}

	batch, err := h.deps.Products.GetProductsByIDs(r.Context(), selectedProductIDs)
	if err != nil {
		return nil, nil, 0, err
	}
	byID := make(map[string]*productsv1.Product, len(batch.GetProducts()))
	for _, product := range batch.GetProducts() {
		byID[product.GetId()] = product
	}

	items := make([]*ordersv1.CreateOrderItem, 0, len(selected))
	names := make(map[string]string, len(selected))
	var totalCents int64
	for _, item := range selected {
		product, ok := byID[item.GetProductId()]
		if !ok {
			return nil, nil, 0, &checkoutError{
				status:  http.StatusConflict,
				title:   "Product unavailable",
				message: "One or more cart items are no longer available.",
			}
		}
		if product.GetMerchantId() != merchantID {
			return nil, nil, 0, &checkoutError{
				status:  http.StatusConflict,
				title:   "Merchant mismatch",
				message: "One or more cart items no longer match the selected merchant.",
			}
		}
		lineTotal := product.GetPriceCents() * int64(item.GetQuantity())
		totalCents += lineTotal
		names[item.GetProductId()] = product.GetName()
		items = append(items, &ordersv1.CreateOrderItem{
			ProductId:      item.GetProductId(),
			Quantity:       item.GetQuantity(),
			UnitPriceCents: product.GetPriceCents(),
		})
	}
	return items, names, totalCents, nil
}

func shippingAddressFromForm(r *http.Request) (*paymentv1.Address, bool) {
	addr := &paymentv1.Address{
		Name:       strings.TrimSpace(r.FormValue("shipping_name")),
		Line1:      strings.TrimSpace(r.FormValue("shipping_line1")),
		Line2:      strings.TrimSpace(r.FormValue("shipping_line2")),
		City:       strings.TrimSpace(r.FormValue("shipping_city")),
		Region:     strings.TrimSpace(r.FormValue("shipping_region")),
		PostalCode: strings.TrimSpace(r.FormValue("shipping_postal_code")),
		Country:    strings.TrimSpace(r.FormValue("shipping_country")),
	}
	if addr.Line1 == "" || addr.City == "" || addr.PostalCode == "" || addr.Country == "" {
		return nil, false
	}
	return addr, true
}
