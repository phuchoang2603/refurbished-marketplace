package shared

import (
	"net/http"
	"net/url"
	"strings"

	paymentv1 "refurbished-marketplace/shared/proto/payment/v1"
)

type HostedPaymentConfig struct {
	GatewayBaseURL string
}

func RequestBaseURL(r *http.Request) string {
	if r == nil {
		return ""
	}
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if proto := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")); proto != "" {
		scheme = proto
	}
	host := strings.TrimSpace(r.Host)
	if host == "" {
		return ""
	}
	if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-Host")); forwarded != "" {
		host = forwarded
	}
	return scheme + "://" + host
}

func OrderPageURL(r *http.Request, orderID string) string {
	base := strings.TrimRight(RequestBaseURL(r), "/")
	if base == "" || strings.TrimSpace(orderID) == "" {
		return ""
	}
	return base + "/orders/" + strings.TrimSpace(orderID)
}

func absolutizeWebURL(r *http.Request, raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		return raw
	}
	base := strings.TrimRight(RequestBaseURL(r), "/")
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
	v.Set("return_url", absolutizeWebURL(r, session.GetReturnUrl()))
	v.Set("cancel_url", absolutizeWebURL(r, session.GetCancelUrl()))
	if webBase := strings.TrimRight(RequestBaseURL(r), "/"); webBase != "" {
		v.Set("callback_url", webBase+"/callbacks/hosted-payment")
	}
	return strings.TrimRight(cfg.GatewayBaseURL, "/") + "/pay?" + v.Encode()
}
