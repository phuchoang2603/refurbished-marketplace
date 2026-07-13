package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
)

type callbackError struct{ status string }

func (e *callbackError) Error() string {
	return "callback failed: " + e.status
}

func postCallback(ctx context.Context, callbackURL string, req callbackRequest) (err error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, normalizeCallbackURL(callbackURL), bytes.NewReader(payload))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()
	if resp.StatusCode != http.StatusNoContent {
		return &callbackError{status: resp.Status}
	}
	return nil
}

func normalizeCallbackURL(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	hostname := parsed.Hostname()
	if hostname == "localhost" || hostname == "127.0.0.1" {
		port := parsed.Port()
		if port == "" {
			port = "80"
			if parsed.Scheme == "https" {
				port = "443"
			}
		}
		parsed.Host = "host.docker.internal:" + port
		return parsed.String()
	}
	// Edge TLS (Cloudflare) redirects http→https with 301; Go's client then
	// retries as GET and the web callback route returns 405. Only rewrite
	// public hostnames — leave cluster DNS (web, *.svc.cluster.local) alone.
	if parsed.Scheme == "http" && isPublicCallbackHost(hostname) {
		parsed.Scheme = "https"
		return parsed.String()
	}
	return raw
}

func isPublicCallbackHost(hostname string) bool {
	if hostname == "" || !strings.Contains(hostname, ".") {
		return false
	}
	if strings.HasSuffix(hostname, ".svc") ||
		strings.HasSuffix(hostname, ".cluster.local") ||
		strings.HasSuffix(hostname, ".local") {
		return false
	}
	return true
}
