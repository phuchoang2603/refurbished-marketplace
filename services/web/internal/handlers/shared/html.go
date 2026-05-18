package shared

import (
	"context"
	"net/http"
	"strings"

	sharedviews "refurbished-marketplace/services/web/internal/views/shared"

	"github.com/a-h/templ"
	"github.com/starfederation/datastar-go/datastar"
)

func WriteHTML(w http.ResponseWriter, r *http.Request, status int, component templ.Component) {
	renderComponent(w, r, status, nil, component)
}

func WriteFragment(w http.ResponseWriter, r *http.Request, status int, selector string, component templ.Component) {
	if r != nil && acceptsDatastar(r) {
		sse := datastar.NewSSE(w, r)
		opts := []datastar.PatchElementOption{datastar.WithModeOuter()}
		if selector != "" {
			opts = append(opts, datastar.WithSelector(selector))
		}
		_ = sse.PatchElementTempl(component, opts...)
		return
	}

	headers := http.Header{}
	if selector != "" {
		headers.Set("datastar-selector", selector)
		headers.Set("datastar-mode", "outer")
	}
	renderComponent(w, r, status, headers, component)
}

func WriteBadRequest(w http.ResponseWriter, r *http.Request, message string) {
	WriteHTML(w, r, http.StatusBadRequest, sharedviews.MessagePage("Bad request", message))
}

func WriteUnauthorized(w http.ResponseWriter, r *http.Request) {
	WriteHTML(w, r, http.StatusUnauthorized, sharedviews.MessagePage("Unauthorized", "unauthorized"))
}

func renderComponent(w http.ResponseWriter, r *http.Request, status int, headers http.Header, component templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	for key, values := range headers {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(status)
	ctx := context.Background()
	if r != nil {
		ctx = r.Context()
	}
	_ = component.Render(ctx, w)
}

func acceptsDatastar(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Accept"), "text/event-stream")
}
