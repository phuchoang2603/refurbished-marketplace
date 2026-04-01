package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type createProductRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	PriceCents  int64  `json:"price_cents"`
	Stock       int32  `json:"stock"`
}

type productResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	PriceCents  int64  `json:"price_cents"`
	Stock       int32  `json:"stock"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func mapProduct(id, name, description string, priceCents int64, stock int32, createdAt, updatedAt *timestamppb.Timestamp) productResponse {
	var created string
	var updated string
	if createdAt != nil {
		created = createdAt.AsTime().UTC().Format("2006-01-02T15:04:05Z07:00")
	}
	if updatedAt != nil {
		updated = updatedAt.AsTime().UTC().Format("2006-01-02T15:04:05Z07:00")
	}

	return productResponse{
		ID:          id,
		Name:        name,
		Description: description,
		PriceCents:  priceCents,
		Stock:       stock,
		CreatedAt:   created,
		UpdatedAt:   updated,
	}
}

func (h *Handler) handleCreateProduct(w http.ResponseWriter, r *http.Request) {
	var req createProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	p, err := h.products.CreateProduct(r.Context(), req.Name, req.Description, req.PriceCents, req.Stock)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, mapProduct(p.Id, p.Name, p.Description, p.PriceCents, p.Stock, p.CreatedAt, p.UpdatedAt))
}

func (h *Handler) handleGetProductByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid product id"})
		return
	}

	p, err := h.products.GetProductByID(r.Context(), id)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, mapProduct(p.Id, p.Name, p.Description, p.PriceCents, p.Stock, p.CreatedAt, p.UpdatedAt))
}

func (h *Handler) handleListProducts(w http.ResponseWriter, r *http.Request) {
	limit := int32(20)
	offset := int32(0)

	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		v, err := strconv.ParseInt(raw, 10, 32)
		if err != nil || v <= 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid limit"})
			return
		}
		limit = int32(v)
	}

	if raw := strings.TrimSpace(r.URL.Query().Get("offset")); raw != "" {
		v, err := strconv.ParseInt(raw, 10, 32)
		if err != nil || v < 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid offset"})
			return
		}
		offset = int32(v)
	}

	resp, err := h.products.ListProducts(r.Context(), limit, offset)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	items := make([]productResponse, 0, len(resp.Products))
	for _, p := range resp.Products {
		items = append(items, mapProduct(p.Id, p.Name, p.Description, p.PriceCents, p.Stock, p.CreatedAt, p.UpdatedAt))
	}

	writeJSON(w, http.StatusOK, map[string]any{"products": items})
}
