package orders

import (
	"net/http"

	webAuth "github.com/phuchoang2603/refurbished-marketplace/services/web/internal/auth"
	checkoutv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/checkout/v1"

	carthandlers "github.com/phuchoang2603/refurbished-marketplace/services/web/internal/handlers/cart"
	shared "github.com/phuchoang2603/refurbished-marketplace/services/web/internal/handlers/shared"
	orderviews "github.com/phuchoang2603/refurbished-marketplace/services/web/internal/views/orders"
	sharedviews "github.com/phuchoang2603/refurbished-marketplace/services/web/internal/views/shared"
	ordersv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/orders/v1"

	"github.com/go-chi/chi/v5"
)

type Handler struct{ deps *shared.Dependencies }

func New(deps *shared.Dependencies) *Handler { return &Handler{deps: deps} }

func ordersUnavailableView() sharedviews.UnavailableView {
	return shared.NewUnavailableView("Orders", "orders", "Orders unavailable", "Order data is temporarily unavailable. Please try again shortly.")
}

func (h *Handler) RegisterPages(r chi.Router) {
	r.Get("/orders", h.handleListOrdersByBuyer)
	r.Get("/orders/{id}", h.handleGetOrderByID)
}

func OrderToView(order *ordersv1.Order) sharedviews.OrderView {
	items := make([]sharedviews.OrderItemView, 0, len(order.GetItems()))
	for _, item := range order.GetItems() {
		items = append(items, sharedviews.OrderItemView{ID: item.GetId(), OrderID: item.GetOrderId(), ProductID: item.GetProductId(), Quantity: item.GetQuantity(), UnitPriceCents: item.GetUnitPriceCents(), LineTotalCents: item.GetLineTotalCents(), CreatedAt: shared.FormatTimestamp(item.GetCreatedAt())})
	}
	return sharedviews.OrderView{ID: order.GetId(), BuyerUserID: order.GetBuyerUserId(), Status: order.GetStatus().String(), TotalCents: order.GetTotalCents(), Items: items, CreatedAt: shared.FormatTimestamp(order.GetCreatedAt()), UpdatedAt: shared.FormatTimestamp(order.GetUpdatedAt())}
}

func applyHostedPaymentState(view *sharedviews.OrderView, status, failureReason, paymentURL string) {
	view.PaymentStatus = status
	view.PaymentFailureReason = failureReason
	view.PaymentURL = paymentURL
}

func (h *Handler) handleGetOrderByID(w http.ResponseWriter, r *http.Request) {
	buyerUserID, ok := shared.RequireUserID(w, r)
	if !ok {
		return
	}
	id, ok := shared.RequirePathValue(w, r, "id", "invalid order id")
	if !ok {
		return
	}

	order, err := h.deps.Orders.GetOrderByID(r.Context(), id)
	if err != nil {
		if shared.IsUnavailableError(err) {
			shared.WriteUnavailablePage(w, r, http.StatusServiceUnavailable, ordersUnavailableView())
			return
		}
		shared.WriteGRPCError(w, r, err)
		return
	}
	if order.GetBuyerUserId() != buyerUserID {
		shared.WritePopup(w, r, http.StatusForbidden, "Forbidden", "order does not belong to the current user")
		return
	}

	view := OrderToView(order)
	if h.deps.Checkout != nil {
		checkout, err := h.deps.Checkout.GetCheckout(webAuth.OutgoingAccessContext(r), &checkoutv1.GetCheckoutRequest{OrderId: id, BuyerUserId: buyerUserID})
		if err == nil && checkout != nil && checkout.GetOrderId() == id {
			applyHostedPaymentState(&view, checkout.GetPaymentStatus(), checkout.GetFailureReason(), checkout.GetHostedCheckoutUrl())
		} else if err != nil && !shared.IsUnavailableError(err) && !shared.IsNotFoundError(err) {
			shared.WriteGRPCError(w, r, err)
			return
		}
	}
	if view.PaymentStatus == "SUCCEEDED" || view.PaymentStatus == "EXPIRED" {
		ids := make([]string, 0, len(order.GetItems()))
		for _, item := range order.GetItems() {
			ids = append(ids, item.GetProductId())
		}
		carthandlers.DrainPaidProductIDs(w, r, h.deps.Cart, ids, order.GetMerchantId())
	}

	shared.WriteHTML(w, r, http.StatusOK, orderviews.OrderDetailPage(view))
}

func (h *Handler) handleListOrdersByBuyer(w http.ResponseWriter, r *http.Request) {
	buyerUserID, ok := shared.RequireUserID(w, r)
	if !ok {
		return
	}

	resp, err := h.deps.Orders.ListOrdersByBuyer(r.Context(), buyerUserID, 20, 0)
	if err != nil {
		if shared.IsUnavailableError(err) {
			shared.WriteUnavailablePage(w, r, http.StatusServiceUnavailable, ordersUnavailableView())
			return
		}
		shared.WriteGRPCError(w, r, err)
		return
	}

	items := make([]sharedviews.OrderView, 0, len(resp.Orders))
	for _, order := range resp.Orders {
		items = append(items, OrderToView(order))
	}

	shared.WriteHTML(w, r, http.StatusOK, orderviews.OrdersPage(items))
}
