package shared

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	paymentv1 "refurbished-marketplace/shared/proto/payment/v1"
)

type HostedPaymentConfig struct {
	GatewayBaseURL string
	// PublicBaseURL optionally overrides request-derived absolute web URLs
	// (return/cancel/callback). Use https://shop.example when TLS terminates
	// at the edge and the origin only sees plain HTTP.
	PublicBaseURL string
	// CallbackBaseURL optionally overrides only the simulator→web callback
	// base (for example http://web:8080) so server-side POSTs stay in-cluster.
	CallbackBaseURL string
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

func absolutizeWebURL(cfg HostedPaymentConfig, r *http.Request, raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		return raw
	}
	base := webBaseURL(cfg, r)
	if base == "" {
		return raw
	}
	if !strings.HasPrefix(raw, "/") {
		raw = "/" + raw
	}
	return base + raw
}

func BuildHostedPaymentURL(cfg HostedPaymentConfig, r *http.Request, session *paymentv1.CreateHostedPaymentSessionResponse) string {
	if session == nil || cfg.GatewayBaseURL == "" || session.GetPaymentSessionId() == "" {
		return ""
	}
	v := url.Values{}
	v.Set("order_id", session.GetOrderId())
	v.Set("payment_session_id", session.GetPaymentSessionId())
	v.Set("return_url", absolutizeWebURL(cfg, r, session.GetReturnUrl()))
	v.Set("cancel_url", absolutizeWebURL(cfg, r, session.GetCancelUrl()))
	callbackBase := strings.TrimRight(strings.TrimSpace(cfg.CallbackBaseURL), "/")
	if callbackBase == "" {
		callbackBase = webBaseURL(cfg, r)
	}
	if callbackBase != "" {
		v.Set("callback_url", callbackBase+"/callbacks/hosted-payment")
	}
	return strings.TrimRight(cfg.GatewayBaseURL, "/") + "/pay?" + v.Encode()
}
