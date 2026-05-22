package tests

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestVendoredDatastarRuntimeIsServed(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/static/vendor/datastar.js", nil)

	newTestRouter(t, routerDeps{}).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Content-Type"); !strings.Contains(got, "javascript") {
		t.Fatalf("content-type = %q, want javascript", got)
	}
	body := rec.Body.String()
	for _, want := range []string{"datastar", "fetch"} {
		if !strings.Contains(body, want) {
			t.Fatalf("body missing %q", want)
		}
	}
}
