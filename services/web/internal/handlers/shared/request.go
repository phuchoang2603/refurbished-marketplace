package shared

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	webAuth "refurbished-marketplace/services/web/internal/auth"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

var ErrInvalidRequestBody = errors.New("invalid request body")

func RefreshTokenFromForm(r *http.Request) (string, error) {
	if !parseForm(r) {
		return "", ErrInvalidRequestBody
	}
	return r.FormValue("refresh_token"), nil
}

func EmailPasswordFromForm(r *http.Request) (string, string, error) {
	if !parseForm(r) {
		return "", "", ErrInvalidRequestBody
	}
	return r.FormValue("email"), r.FormValue("password"), nil
}

func ProductQuantityFromForm(r *http.Request) (string, int32, error) {
	if !parseForm(r) {
		return "", 0, ErrInvalidRequestBody
	}
	quantity, err := parseInt32FormValue(r, "quantity")
	if err != nil {
		return "", 0, err
	}
	return r.FormValue("product_id"), quantity, nil
}

func RequirePathValue(w http.ResponseWriter, r *http.Request, key, errorMessage string) (string, bool) {
	value := strings.TrimSpace(chi.URLParam(r, key))
	if value == "" {
		WriteBadRequest(w, r, errorMessage)
		return "", false
	}
	return value, true
}

func QueryInt32Param(w http.ResponseWriter, r *http.Request, key string, defaultValue, minValue int32, errorMessage string) (int32, bool) {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return defaultValue, true
	}

	v, err := strconv.ParseInt(raw, 10, 32)
	if err != nil || int32(v) < minValue {
		WriteBadRequest(w, r, errorMessage)
		return 0, false
	}

	return int32(v), true
}

func RequireUserID(w http.ResponseWriter, r *http.Request) (string, bool) {
	userID, ok := webAuth.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		WriteUnauthorized(w, r)
		return "", false
	}
	return userID, true
}

func NextTargetFromRequest(r *http.Request, fallback string) string {
	if r == nil {
		return fallback
	}
	return sanitizeRedirectTarget(r.URL.Query().Get("next"), fallback)
}

func Redirect(w http.ResponseWriter, r *http.Request, location string, status int) {
	if r != nil && acceptsDatastar(r) {
		sse := datastar.NewSSE(w, r)
		_ = sse.Redirect(location)
		return
	}
	http.Redirect(w, r, location, status)
}

func parseForm(r *http.Request) bool {
	return r.ParseForm() == nil
}

func parseInt32FormValue(r *http.Request, key string) (int32, error) {
	value, err := strconv.ParseInt(r.FormValue(key), 10, 32)
	if err != nil {
		return 0, ErrInvalidRequestBody
	}
	return int32(value), nil
}

func sanitizeRedirectTarget(raw, fallback string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallback
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.IsAbs() || parsed.Host != "" {
		return fallback
	}
	if !strings.HasPrefix(parsed.Path, "/") {
		return fallback
	}
	if strings.HasPrefix(parsed.Path, "//") {
		return fallback
	}
	return parsed.RequestURI()
}
