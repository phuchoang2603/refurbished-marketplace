package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	shared "refurbished-marketplace/services/web/internal/handlers/shared"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) registerStatusRoutes(r chi.Router) {
	r.Get("/healthz", h.handleHealthz)
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir()))))
}

func (h *Handler) handleHealthz(w http.ResponseWriter, r *http.Request) {
	shared.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func staticDir() string {
	candidates := []string{"services/web/static", "static", "/static"}
	if _, filename, _, ok := runtime.Caller(0); ok {
		candidates = append([]string{filepath.Clean(filepath.Join(filepath.Dir(filename), "..", "..", "static"))}, candidates...)
	}
	for _, candidate := range candidates {
		info, err := os.Stat(candidate)
		if err == nil && info.IsDir() {
			return candidate
		}
	}
	return "static"
}
