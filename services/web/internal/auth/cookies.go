package auth

import (
	"net/http"
	"strings"
	"time"
)

const (
	AccessCookieName  = "web_access_token"
	RefreshCookieName = "web_refresh_token"
)

func SetTokenCookies(w http.ResponseWriter, r *http.Request, accessToken, refreshToken string, expiresIn, refreshExpiresIn int64) {
	setAuthCookie(w, r, AccessCookieName, accessToken, int(expiresIn))
	setAuthCookie(w, r, RefreshCookieName, refreshToken, int(refreshExpiresIn))
}

func ClearTokenCookies(w http.ResponseWriter, r *http.Request) {
	setAuthCookie(w, r, AccessCookieName, "", -1)
	setAuthCookie(w, r, RefreshCookieName, "", -1)
}

func RefreshTokenFromRequest(r *http.Request) string {
	return cookieValue(r, RefreshCookieName)
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
		Secure:   r != nil && r.TLS != nil,
	}
	if maxAge > 0 {
		cookie.Expires = time.Now().Add(time.Duration(maxAge) * time.Second)
	}
	http.SetCookie(w, cookie)
}
