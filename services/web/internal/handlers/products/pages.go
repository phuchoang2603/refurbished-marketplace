package products

import (
	"net/http"
	"strings"

	webAuth "refurbished-marketplace/services/web/internal/auth"
	shared "refurbished-marketplace/services/web/internal/handlers/shared"
	productviews "refurbished-marketplace/services/web/internal/views/products"
	sharedviews "refurbished-marketplace/services/web/internal/views/shared"

	"github.com/go-chi/chi/v5"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Handler struct{ deps *shared.Dependencies }

func New(deps *shared.Dependencies) *Handler { return &Handler{deps: deps} }

func productsUnavailableView() sharedviews.UnavailableView {
	return shared.NewUnavailableView("Products", "products", "Products unavailable", "The catalog is temporarily unavailable. Please try again shortly.")
}

func productManagementUnavailableView() sharedviews.UnavailableView {
	return shared.NewUnavailableView("Create product", "create-product", "Product management unavailable", "Seller product management is temporarily unavailable. Please try again shortly.")
}

func productDetailUnavailableView() sharedviews.UnavailableView {
	return shared.NewUnavailableView("Product unavailable", "product-detail-unavailable", "Product unavailable", "Product availability is temporarily unavailable. Please try again shortly.")
}

func (h *Handler) RegisterPages(r chi.Router) {
	r.Get("/", h.handleListProducts)
	r.Get("/products", h.handleListProducts)
	r.Get("/products/{id}", h.handleGetProductByID)
}

func (h *Handler) RegisterProtectedPages(r chi.Router) {
	r.Get("/seller/products", h.handleListSellerProducts)
	r.Get("/seller/products/new", h.handleNewProductPage)
}

func mapProductView(id, merchantID, name, description string, priceCents int64, stock int32, isOwner bool, createdAt, updatedAt *timestamppb.Timestamp) sharedviews.ProductView {
	return sharedviews.ProductView{ID: id, MerchantID: merchantID, IsOwner: isOwner, Name: name, Description: description, PriceCents: priceCents, Stock: stock, CreatedAt: shared.FormatTimestamp(createdAt), UpdatedAt: shared.FormatTimestamp(updatedAt)}
}

func (h *Handler) handleGetProductByID(w http.ResponseWriter, r *http.Request) {
	id, ok := shared.RequirePathValue(w, r, "id", "invalid product id")
	if !ok {
		return
	}
	viewerUserID, _ := webAuth.UserIDFromContext(r.Context())

	p, err := h.deps.Products.GetProductByID(r.Context(), id)
	if err != nil {
		if shared.IsUnavailableError(err) {
			shared.WriteUnavailablePage(w, r, http.StatusServiceUnavailable, productsUnavailableView())
			return
		}
		shared.WriteGRPCError(w, r, err)
		return
	}
	if p.AvailableQty == nil {
		shared.WriteUnavailablePage(w, r, http.StatusServiceUnavailable, productDetailUnavailableView())
		return
	}

	shared.WriteHTML(w, r, http.StatusOK, productviews.ProductDetailPage(mapProductView(p.Id, p.MerchantId, p.Name, p.Description, p.PriceCents, p.GetAvailableQty(), viewerUserID != "" && viewerUserID == p.GetMerchantId(), p.CreatedAt, p.UpdatedAt)))
}

func (h *Handler) handleListProducts(w http.ResponseWriter, r *http.Request) {
	limit, ok := shared.QueryInt32Param(w, r, "limit", 20, 1, "invalid limit")
	if !ok {
		return
	}
	offset, ok := shared.QueryInt32Param(w, r, "offset", 0, 0, "invalid offset")
	if !ok {
		return
	}
	viewerUserID, _ := webAuth.UserIDFromContext(r.Context())

	resp, err := h.deps.Products.ListProducts(r.Context(), limit, offset)
	if err != nil {
		if shared.IsUnavailableError(err) {
			shared.WriteUnavailablePage(w, r, http.StatusServiceUnavailable, productsUnavailableView())
			return
		}
		shared.WriteGRPCError(w, r, err)
		return
	}

	items := make([]sharedviews.ProductView, 0, len(resp.Products))
	for _, p := range resp.Products {
		if viewerUserID != "" && p.GetMerchantId() == viewerUserID {
			continue
		}
		items = append(items, mapProductView(p.Id, p.MerchantId, p.Name, p.Description, p.PriceCents, 0, false, p.CreatedAt, p.UpdatedAt))
	}

	shared.WriteHTML(w, r, http.StatusOK, productviews.ProductsPage(items))
}

func (h *Handler) handleNewProductPage(w http.ResponseWriter, r *http.Request) {
	shared.WriteHTML(w, r, http.StatusOK, productviews.CreateProductPage())
}

func (h *Handler) handleListSellerProducts(w http.ResponseWriter, r *http.Request) {
	userID, ok := shared.RequireUserID(w, r)
	if !ok {
		return
	}
	resp, err := h.deps.Products.ListProducts(r.Context(), 100, 0)
	if err != nil {
		if shared.IsUnavailableError(err) {
			shared.WriteUnavailablePage(w, r, http.StatusServiceUnavailable, productManagementUnavailableView())
			return
		}
		shared.WriteGRPCError(w, r, err)
		return
	}
	items := make([]sharedviews.ProductView, 0, len(resp.Products))
	for _, p := range resp.Products {
		if p.GetMerchantId() != userID {
			continue
		}
		items = append(items, mapProductView(p.Id, p.MerchantId, p.Name, p.Description, p.PriceCents, 0, true, p.CreatedAt, p.UpdatedAt))
	}
	shared.WriteHTML(w, r, http.StatusOK, productviews.SellerProductsPage(items))
}

func normalizeProductCreateInput(name, description string, priceCents int64, initialStock int32) (string, string, bool) {
	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)
	if name == "" || priceCents <= 0 || initialStock <= 0 {
		return "", "", false
	}
	return name, description, true
}
