package cart

import (
	"context"
	"net/http"
	"strings"

	shared "refurbished-marketplace/services/web/internal/handlers/shared"
	cartviews "refurbished-marketplace/services/web/internal/views/cart"
	sharedviews "refurbished-marketplace/services/web/internal/views/shared"
	cartv1 "refurbished-marketplace/shared/proto/cart/v1"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const cartCookieName = "cart_id"

type Handler struct{ deps *shared.Dependencies }

func New(deps *shared.Dependencies) *Handler { return &Handler{deps: deps} }

func (h *Handler) RegisterPages(r chi.Router) {
	r.Get("/cart", h.handleGetCart)
}

func (h *Handler) mapCartView(ctx context.Context, c *cartv1.Cart) (sharedviews.CartView, error) {
	items := make([]sharedviews.CartItemView, 0, len(c.GetItems()))
	var estimatedTotalCents int64
	for _, item := range c.GetItems() {
		view := sharedviews.CartItemView{ProductID: item.GetProductId(), Quantity: item.GetQuantity()}
		if h.deps.Products != nil {
			product, err := h.deps.Products.GetProductByID(ctx, item.GetProductId())
			if err != nil {
				if st, ok := status.FromError(err); !ok || st.Code() != codes.NotFound {
					return sharedviews.CartView{}, err
				}
			} else {
				view.ProductName = product.GetName()
				view.ProductPrice = product.GetPriceCents()
				view.LineTotalCents = product.GetPriceCents() * int64(item.GetQuantity())
				view.Available = true
				estimatedTotalCents += view.LineTotalCents
			}
		}
		items = append(items, view)
	}
	return sharedviews.CartView{CartID: c.GetCartId(), Items: items, EstimatedTotalCents: estimatedTotalCents, CreatedAt: shared.FormatTimestamp(c.GetCreatedAt()), UpdatedAt: shared.FormatTimestamp(c.GetUpdatedAt())}, nil
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

func (h *Handler) handleGetCart(w http.ResponseWriter, r *http.Request) {
	cartID := h.getOrCreateCartID(w, r)
	cart, err := h.deps.Cart.GetCart(r.Context(), cartID)
	if err != nil {
		shared.WriteGRPCError(w, r, err)
		return
	}
	view, err := h.mapCartView(r.Context(), cart)
	if err != nil {
		shared.WriteGRPCError(w, r, err)
		return
	}
	shared.WriteHTML(w, r, http.StatusOK, cartviews.CartPage(view))
}
