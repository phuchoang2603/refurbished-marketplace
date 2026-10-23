package shared

import (
	"encoding/json"
	"net/http"
	"strings"
)

type HostedPaymentConfig struct {
	PublicBaseURL string
}

func RequestBaseURL(r *http.Request) string {
	if r == nil {
		return ""
	}
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if proto := firstForwardedValue(r.Header.Get("X-Forwarded-Proto")); proto != "" {
		scheme = proto
	} else if visitorScheme := cfVisitorScheme(r.Header.Get("Cf-Visitor")); visitorScheme != "" {
		scheme = visitorScheme
	}
	host := strings.TrimSpace(r.Host)
	if host == "" {
		return ""
	}
	if forwarded := firstForwardedValue(r.Header.Get("X-Forwarded-Host")); forwarded != "" {
		host = forwarded
	}
	return scheme + "://" + host
}

func firstForwardedValue(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if i := strings.IndexByte(raw, ','); i >= 0 {
		raw = raw[:i]
	}
	return strings.ToLower(strings.TrimSpace(raw))
}

func cfVisitorScheme(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	var visitor struct {
		Scheme string `json:"scheme"`
	}
	if err := json.Unmarshal([]byte(raw), &visitor); err != nil {
		return ""
	}
	switch scheme := strings.ToLower(strings.TrimSpace(visitor.Scheme)); scheme {
	case "http", "https":
		return scheme
	default:
		return ""
	}
}

func webBaseURL(cfg HostedPaymentConfig, r *http.Request) string {
	if base := strings.TrimRight(strings.TrimSpace(cfg.PublicBaseURL), "/"); base != "" {
		return base
	}
	return strings.TrimRight(RequestBaseURL(r), "/")
}

func OrderPageURL(r *http.Request, orderID string) string {
	return OrderPageURLWithConfig(HostedPaymentConfig{}, r, orderID)
}

func OrderPageURLWithConfig(cfg HostedPaymentConfig, r *http.Request, orderID string) string {
	base := webBaseURL(cfg, r)
	if base == "" || strings.TrimSpace(orderID) == "" {
		return ""
	}
	return base + "/orders/" + strings.TrimSpace(orderID)
}
