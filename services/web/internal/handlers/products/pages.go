package products

import (
	"net/http"
	"strings"

	webAuth "github.com/phuchoang2603/refurbished-marketplace/services/web/internal/auth"
	shared "github.com/phuchoang2603/refurbished-marketplace/services/web/internal/handlers/shared"
	productviews "github.com/phuchoang2603/refurbished-marketplace/services/web/internal/views/products"
	sharedviews "github.com/phuchoang2603/refurbished-marketplace/services/web/internal/views/shared"
	searchv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/search/v1"

	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Handler struct{ deps *shared.Dependencies }

func New(deps *shared.Dependencies) *Handler { return &Handler{deps: deps} }

func catalogUnavailableView() sharedviews.UnavailableView {
	return shared.NewUnavailableView("Products", "products", "Catalog unavailable", "The catalog is temporarily unavailable. Please try again shortly.")
}

func productsUnavailableView() sharedviews.UnavailableView {
	return shared.NewUnavailableView("Products", "products", "Products unavailable", "The catalog is temporarily unavailable. Please try again shortly.")
}

func productManagementUnavailableView() sharedviews.UnavailableView {
	return shared.NewUnavailableView("Create product", "create-product", "Product management unavailable", "Seller product management is temporarily unavailable. Please try again shortly.")
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

func mapListingHit(hit *searchv1.ListingHit, isOwner bool) sharedviews.ProductView {
	return mapProductView(hit.GetId(), hit.GetMerchantId(), hit.GetName(), hit.GetDescription(), hit.GetPriceCents(), 0, isOwner, hit.GetCreatedAt(), hit.GetCreatedAt())
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
	view := mapProductView(p.Id, p.MerchantId, p.Name, p.Description, p.PriceCents, 0, viewerUserID != "" && viewerUserID == p.GetMerchantId(), p.CreatedAt, p.UpdatedAt)
	view.JustCreated = view.IsOwner && r.URL.Query().Get("created") == "1"
	view.StockState = "unavailable"
	if h.deps.Inventory != nil {
		row, stockErr := h.deps.Inventory.GetStock(r.Context(), id)
		switch {
		case stockErr == nil && row != nil:
			view.StockState = "ready"
			view.Stock = row.GetAvailableQty()
		case status.Code(stockErr) == codes.NotFound:
			view.StockState = "pending"
		}
	}
	shared.WriteHTML(w, r, http.StatusOK, productviews.ProductDetailPage(view))
}

func (h *Handler) handleListProducts(w http.ResponseWriter, r *http.Request) {
	if h.deps.Search == nil {
		shared.WriteUnavailablePage(w, r, http.StatusServiceUnavailable, catalogUnavailableView())
		return
	}
	resp, err := h.deps.Search.SearchProducts(r.Context(), "", "", 100, 0)
	if err != nil {
		if shared.IsUnavailableError(err) {
			shared.WriteUnavailablePage(w, r, http.StatusServiceUnavailable, catalogUnavailableView())
			return
		}
		shared.WriteGRPCError(w, r, err)
		return
	}
	items := make([]sharedviews.ProductView, 0, len(resp.Listings))
	for _, hit := range resp.Listings {
		items = append(items, mapListingHit(hit, false))
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
	resp, err := h.deps.Search.SearchProducts(r.Context(), "", userID, 100, 0)
	if err != nil {
		if shared.IsUnavailableError(err) {
			shared.WriteUnavailablePage(w, r, http.StatusServiceUnavailable, productManagementUnavailableView())
			return
		}
		shared.WriteGRPCError(w, r, err)
		return
	}
	items := make([]sharedviews.ProductView, 0, len(resp.Listings))
	for _, hit := range resp.Listings {
		items = append(items, mapListingHit(hit, true))
	}
	shared.WriteHTML(w, r, http.StatusOK, productviews.SellerProductsPage(items))
}

func normalizeProductCreateInput(name, description string, priceCents int64, initialStock int32) (string, string, bool) {
	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)
	if name == "" || priceCents <= 0 || initialStock < 0 {
		return "", "", false
	}
	return name, description, true
}
