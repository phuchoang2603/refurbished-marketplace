package auth

import (
	"net/http"
	"strings"
	"time"
)

const (
	AccessCookieName  = "__Host-web_access_token"
	RefreshCookieName = "__Secure-web_refresh_token"
)

func SetTokenCookies(w http.ResponseWriter, r *http.Request, accessToken, refreshToken string, expiresIn, refreshExpiresIn int64) {
	w.Header().Set("Cache-Control", "no-store")
	setAuthCookie(w, r, AccessCookieName, accessToken, int(expiresIn))
	setAuthCookie(w, r, RefreshCookieName, refreshToken, int(refreshExpiresIn))
}

func ClearTokenCookies(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	setAuthCookie(w, r, AccessCookieName, "", -1)
	setAuthCookie(w, r, RefreshCookieName, "", -1)
}

func RefreshTokenFromRequest(r *http.Request) string {
	return cookieValue(r, RefreshCookieName)
}

func AccessTokenFromRequest(r *http.Request) string {
	if token := accessTokenFromContext(r.Context()); token != "" {
		return token
	}
	return cookieValue(r, AccessCookieName)
}

func cookieValue(r *http.Request, name string) string {
	if c, err := r.Cookie(name); err == nil {
		return strings.TrimSpace(c.Value)
	}
	return ""
}

func setAuthCookie(w http.ResponseWriter, r *http.Request, name, value string, maxAge int) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   maxAge,
		Secure:   true,
	}
	if maxAge > 0 {
		cookie.Expires = time.Now().Add(time.Duration(maxAge) * time.Second)
	}
	http.SetCookie(w, cookie)
}
