package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

type callbackError struct{ status string }

func (e *callbackError) Error() string {
	return "callback failed: " + e.status
}

func postCallback(ctx context.Context, callbackURL string, req callbackRequest) error {
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
	defer resp.Body.Close()
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
	if hostname != "localhost" && hostname != "127.0.0.1" {
		return raw
	}
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
