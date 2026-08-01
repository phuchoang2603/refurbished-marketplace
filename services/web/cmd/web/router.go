package main

import (
	"net/http"
	"time"

	"refurbished-marketplace/services/web/internal/handlers"
	sharedlog "refurbished-marketplace/shared/observe/log"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func newRouter(h *handlers.Handler) http.Handler {
	router := chi.NewRouter()
	router.Use(
		middleware.RealIP,
		middleware.Recoverer,
		middleware.Timeout(60*time.Second),
		otelHTTPMiddleware(),
		withHTTPRouteSpanName, // inside otelhttp so name/route update before span ends
		sharedlog.HTTPAccess,  // after otelhttp so request context carries the span
	)
	h.Register(router)
	return router
}
