package metric

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.opentelemetry.io/otel"
)

func TestExportToMetricsHTTPEndpoint(t *testing.T) {
	requests := make(chan string, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests <- r.Method + " " + r.URL.Path
		w.Header().Set("Content-Type", "application/x-protobuf")
	}))
	defer server.Close()
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://unused.invalid:4318")
	t.Setenv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT", server.URL+"/opentelemetry/v1/metrics")
	previous := otel.GetMeterProvider()
	defer otel.SetMeterProvider(previous)
	stop, err := Init(context.Background(), LoadConfig("test-service"))
	if err != nil {
		t.Fatal(err)
	}
	counter, err := otel.Meter("test").Int64Counter("test.calls")
	if err != nil {
		t.Fatal(err)
	}
	counter.Add(context.Background(), 1)
	if err := stop(context.Background()); err != nil {
		t.Fatal(err)
	}
	select {
	case got := <-requests:
		if got != "POST /opentelemetry/v1/metrics" {
			t.Fatalf("request = %s", got)
		}
	default:
		t.Fatal("no HTTP metric export")
	}
}
