package main

import (
	"net/http"
	"time"

	"github.com/phuchoang2603/refurbished-marketplace/services/web/internal/handlers"
	sharedlog "github.com/phuchoang2603/refurbished-marketplace/shared/observe/log"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func newRouter(h *handlers.Handler) http.Handler {
	router := chi.NewRouter()
	router.Use(
		middleware.Recoverer,
		middleware.Timeout(60*time.Second),
		otelHTTPMiddleware(),
		withHTTPRouteAttr, // http.route attr only; span name is From WithSpanNameFormatter
		sharedlog.HTTPAccessWithPath(httpRoutePattern),
	)
	h.Register(router)
	return router
}
