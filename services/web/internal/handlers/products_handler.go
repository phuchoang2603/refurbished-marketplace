package handlers

import (
	"net/http"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type productResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	PriceCents  int64   `json:"price_cents"`
	Stock       int32   `json:"stock"`
	TerminalID  string  `json:"terminal_id"`
	XPos        float64 `json:"x_pos"`
	YPos        float64 `json:"y_pos"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

func mapProduct(id, name, description string, priceCents int64, stock int32, terminalID string, xPos, yPos float64, createdAt, updatedAt *timestamppb.Timestamp) productResponse {
	return productResponse{
		ID:          id,
		Name:        name,
		Description: description,
		PriceCents:  priceCents,
		Stock:       stock,
		TerminalID:  terminalID,
		XPos:        xPos,
		YPos:        yPos,
		CreatedAt:   formatTimestamp(createdAt),
		UpdatedAt:   formatTimestamp(updatedAt),
	}
}

func (h *Handler) handleGetProductByID(w http.ResponseWriter, r *http.Request) {
	id, ok := requirePathValue(w, r, "id", "invalid product id")
	if !ok {
		return
	}

	p, err := h.products.GetProductByID(r.Context(), id)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, mapProduct(p.Id, p.Name, p.Description, p.PriceCents, 0, p.TerminalId, p.XPos, p.YPos, p.CreatedAt, p.UpdatedAt))
}

func (h *Handler) handleListProducts(w http.ResponseWriter, r *http.Request) {
	limit, ok := queryInt32Param(w, r, "limit", 20, 1, "invalid limit")
	if !ok {
		return
	}
	offset, ok := queryInt32Param(w, r, "offset", 0, 0, "invalid offset")
	if !ok {
		return
	}

	resp, err := h.products.ListProducts(r.Context(), limit, offset)
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	items := make([]productResponse, 0, len(resp.Products))
	for _, p := range resp.Products {
		items = append(items, mapProduct(p.Id, p.Name, p.Description, p.PriceCents, 0, p.TerminalId, p.XPos, p.YPos, p.CreatedAt, p.UpdatedAt))
	}

	writeJSON(w, http.StatusOK, map[string]any{"products": items})
}
