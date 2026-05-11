package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"refurbished-marketplace/services/web/internal/views"

	webAuth "refurbished-marketplace/services/web/internal/auth"

	"github.com/a-h/templ"
	"github.com/starfederation/datastar-go/datastar"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const timestampFormat = "2006-01-02T15:04:05Z07:00"

var errInvalidRequestBody = errors.New("invalid request body")

func formatTimestamp(ts *timestamppb.Timestamp) string {
	if ts == nil {
		return ""
	}
	return ts.AsTime().UTC().Format(timestampFormat)
}

func writeGRPCError(w http.ResponseWriter, r *http.Request, err error) {
	st, ok := status.FromError(err)
	if !ok {
		writeHTML(w, r, http.StatusInternalServerError, views.MessagePage("Error", "internal server error"))
		return
	}

	switch st.Code() {
	case codes.InvalidArgument:
		writeHTML(w, r, http.StatusBadRequest, views.MessagePage("Bad request", st.Message()))
	case codes.NotFound:
		writeHTML(w, r, http.StatusNotFound, views.MessagePage("Not found", st.Message()))
	case codes.PermissionDenied:
		writeHTML(w, r, http.StatusForbidden, views.MessagePage("Forbidden", st.Message()))
	case codes.AlreadyExists:
		writeHTML(w, r, http.StatusConflict, views.MessagePage("Conflict", st.Message()))
	case codes.Unauthenticated:
		writeHTML(w, r, http.StatusUnauthorized, views.MessagePage("Unauthorized", st.Message()))
	default:
		writeHTML(w, r, http.StatusInternalServerError, views.MessagePage("Error", "internal server error"))
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func decodeJSONResponse(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return false
	}
	return true
}

func writeHTML(w http.ResponseWriter, r *http.Request, status int, component templ.Component) {
	renderComponent(w, r, status, nil, component)
}

func writeFragment(w http.ResponseWriter, r *http.Request, status int, selector string, component templ.Component) {
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

func acceptsDatastar(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Accept"), "text/event-stream")
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

func parseForm(r *http.Request) bool {
	return r.ParseForm() == nil
}

func parseInt32FormValue(r *http.Request, key string) (int32, error) {
	value, err := strconv.ParseInt(r.FormValue(key), 10, 32)
	if err != nil {
		return 0, errInvalidRequestBody
	}
	return int32(value), nil
}

func writeBadRequest(w http.ResponseWriter, r *http.Request, message string) {
	writeHTML(w, r, http.StatusBadRequest, views.MessagePage("Bad request", message))
}

func writeUnauthorized(w http.ResponseWriter, r *http.Request) {
	writeHTML(w, r, http.StatusUnauthorized, views.MessagePage("Unauthorized", "unauthorized"))
}

func requireUserID(w http.ResponseWriter, r *http.Request) (string, bool) {
	userID, ok := webAuth.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		writeUnauthorized(w, r)
		return "", false
	}
	return userID, true
}

func requirePathValue(w http.ResponseWriter, r *http.Request, key, errorMessage string) (string, bool) {
	value := strings.TrimSpace(r.PathValue(key))
	if value == "" {
		writeHTML(w, r, http.StatusBadRequest, views.MessagePage("Bad request", errorMessage))
		return "", false
	}
	return value, true
}

func queryInt32Param(w http.ResponseWriter, r *http.Request, key string, defaultValue int32, minValue int32, errorMessage string) (int32, bool) {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return defaultValue, true
	}

	v, err := strconv.ParseInt(raw, 10, 32)
	if err != nil || int32(v) < minValue {
		writeHTML(w, r, http.StatusBadRequest, views.MessagePage("Bad request", errorMessage))
		return 0, false
	}

	return int32(v), true
}
