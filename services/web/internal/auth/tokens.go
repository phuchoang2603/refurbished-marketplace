package auth

import (
	"net/http"
	"strings"
)

func accessTokenFromRequest(r *http.Request) string {
	if raw := bearerToken(r.Header.Get("Authorization")); raw != "" {
		return raw
	}
	return cookieValue(r, AccessCookieName)
}

func bearerToken(header string) string {
	header = strings.TrimSpace(header)
	if header == "" {
		return ""
	}

	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return ""
	}

	token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	if token == "" {
		return ""
	}

	return token
}
