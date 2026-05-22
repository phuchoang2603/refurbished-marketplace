package shared

import (
	"context"
	"net/http"
	"strings"

	"github.com/a-h/templ"
	"github.com/starfederation/datastar-go/datastar"
	dialog "refurbished-marketplace/services/web/internal/views/components/dialog"
	sharedviews "refurbished-marketplace/services/web/internal/views/shared"
)

func WriteHTML(w http.ResponseWriter, r *http.Request, status int, component templ.Component) {
	renderComponent(w, r, status, nil, component)
}

func WriteFragment(w http.ResponseWriter, r *http.Request, status int, selector string, component templ.Component) {
	if r != nil && acceptsDatastar(r) {
		sse := datastar.NewSSE(w, r)
		_ = sse.PatchElementTempl(dialog.DialogRoot("", ""), datastar.WithSelector("#dialog-root"), datastar.WithModeOuter())
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

func NewUnavailableView(pageTitle, sectionID, title, subtitle string) sharedviews.UnavailableView {
	return sharedviews.UnavailableView{
		PageTitle: pageTitle,
		SectionID: sectionID,
		Title:     title,
		Subtitle:  subtitle,
	}
}

func WriteUnavailablePage(w http.ResponseWriter, r *http.Request, status int, unavailable sharedviews.UnavailableView) {
	WriteHTML(w, r, status, sharedviews.UnavailablePage(unavailable))
}

func WritePopup(w http.ResponseWriter, r *http.Request, status int, title, message string) {
	if r != nil && acceptsDatastar(r) {
		sse := datastar.NewSSE(w, r)
		_ = sse.PatchElementTempl(dialog.DialogRoot(title, message), datastar.WithSelector("#dialog-root"), datastar.WithModeOuter())
		return
	}
	WriteHTML(w, r, status, sharedviews.AppShell(title, sharedviews.DefaultNav, dialog.DialogPageBody(title, message)))
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
