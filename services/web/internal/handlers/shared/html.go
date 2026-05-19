package shared

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"strings"

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
	WritePopup(w, r, http.StatusBadRequest, "Bad request", message)
}

func WriteUnauthorized(w http.ResponseWriter, r *http.Request) {
	WritePopup(w, r, http.StatusUnauthorized, "Unauthorized", "unauthorized")
}

func WritePopup(w http.ResponseWriter, r *http.Request, status int, title, message string) {
	alertText := title + ": " + message
	encodedAlert, err := json.Marshal(alertText)
	if err != nil {
		encodedAlert = []byte(`"internal server error"`)
	}
	alertJS := strings.ReplaceAll(string(encodedAlert), "</", "<\\/")
	if r != nil && acceptsDatastar(r) {
		sse := datastar.NewSSE(w, r)
		_ = sse.ExecuteScript("window.alert(" + alertJS + ")")
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = fmt.Fprintf(w, "<!doctype html><html><head><meta charset=\"utf-8\"><title>%s</title><style>html,body{margin:0;background:transparent;opacity:0}</style></head><body><script>window.alert(%s);window.history.back();</script></body></html>", html.EscapeString(title), alertJS)
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
