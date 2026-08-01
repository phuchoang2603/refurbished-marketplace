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
	// otelhttp starts the server span; WithSpanNameFormatter covers cases where
	// r.Pattern is set, but chi often matches later — withHTTPRouteSpanName
	// re-applies METHOD + route pattern and http.route after the handler runs.
	// HTTPAccess runs inside that stack so logs can use the same pattern.
	router.Use(
		middleware.RealIP,
		middleware.Recoverer,
		middleware.Timeout(60*time.Second),
		otelHTTPMiddleware(),
		withHTTPRouteSpanName,
		sharedlog.HTTPAccessWithPath(httpRoutePattern),
	)
	h.Register(router)
	return router
}
