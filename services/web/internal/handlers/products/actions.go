package products

import (
	"net/http"

	shared "refurbished-marketplace/services/web/internal/handlers/shared"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) RegisterProtectedActions(r chi.Router) {
	r.Post("/seller/products", h.handleCreateProduct)
}

func (h *Handler) handleCreateProduct(w http.ResponseWriter, r *http.Request) {
	userID, ok := shared.RequireUserID(w, r)
	if !ok {
		return
	}
	name, description, priceCents, initialStock, err := shared.ProductCreateFromForm(r)
	if err != nil {
		shared.WriteBadRequest(w, r, "invalid request body")
		return
	}
	name, description, valid := normalizeProductCreateInput(name, description, priceCents, initialStock)
	if !valid {
		shared.WriteBadRequest(w, r, "invalid request body")
		return
	}
	product, err := h.deps.Products.CreateProduct(r.Context(), name, description, priceCents, userID, initialStock)
	if err != nil {
		shared.WriteGRPCError(w, r, err)
		return
	}
	shared.Redirect(w, r, "/products/"+product.GetId(), http.StatusSeeOther)
}
