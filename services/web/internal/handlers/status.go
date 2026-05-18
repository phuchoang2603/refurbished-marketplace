package handlers

import (
	"net/http"

	shared "refurbished-marketplace/services/web/internal/handlers/shared"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) registerStatusRoutes(r chi.Router) {
	r.Get("/healthz", h.handleHealthz)
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("/static"))))
}

func (h *Handler) handleHealthz(w http.ResponseWriter, r *http.Request) {
	shared.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
