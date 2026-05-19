package products

import (
	"net/http"

	shared "refurbished-marketplace/services/web/internal/handlers/shared"
	utils "refurbished-marketplace/services/web/internal/utils"
	productviews "refurbished-marketplace/services/web/internal/views/products"
	sharedviews "refurbished-marketplace/services/web/internal/views/shared"

	"github.com/go-chi/chi/v5"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Handler struct{ deps *shared.Dependencies }

func New(deps *shared.Dependencies) *Handler { return &Handler{deps: deps} }

func (h *Handler) RegisterPages(r chi.Router) {
	r.Get("/", h.handleListProducts)
	r.Get("/products", h.handleListProducts)
	r.Get("/products/{id}", h.handleGetProductByID)
}

func mapProductView(id, name, description string, priceCents int64, stock int32, createdAt, updatedAt *timestamppb.Timestamp) sharedviews.ProductView {
	return sharedviews.ProductView{ID: id, Name: name, Description: description, PriceCents: priceCents, Stock: stock, CreatedAt: utils.FormatTimestamp(createdAt), UpdatedAt: utils.FormatTimestamp(updatedAt)}
}

func (h *Handler) handleGetProductByID(w http.ResponseWriter, r *http.Request) {
	id, ok := shared.RequirePathValue(w, r, "id", "invalid product id")
	if !ok {
		return
	}

	p, err := h.deps.Products.GetProductByID(r.Context(), id)
	if err != nil {
		shared.WriteGRPCError(w, r, err)
		return
	}

	shared.WriteHTML(w, r, http.StatusOK, productviews.ProductDetailPage(mapProductView(p.Id, p.Name, p.Description, p.PriceCents, 0, p.CreatedAt, p.UpdatedAt)))
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

	resp, err := h.deps.Products.ListProducts(r.Context(), limit, offset)
	if err != nil {
		shared.WriteGRPCError(w, r, err)
		return
	}

	items := make([]sharedviews.ProductView, 0, len(resp.Products))
	for _, p := range resp.Products {
		items = append(items, mapProductView(p.Id, p.Name, p.Description, p.PriceCents, 0, p.CreatedAt, p.UpdatedAt))
	}

	shared.WriteHTML(w, r, http.StatusOK, productviews.ProductsPage(items))
}
