package trace

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.opentelemetry.io/otel"
)

func TestInitMergesDefaultResourceSchema(t *testing.T) {
	shutdown, err := Init(context.Background(), Config{
		ServiceName: "web",
		Endpoint:    "localhost:4317",
	})
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	t.Cleanup(func() {
		_ = shutdown(context.Background())
	})
}

func TestExportToTracesHTTPEndpoint(t *testing.T) {
	requests := make(chan string, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests <- r.Method + " " + r.URL.Path
		w.Header().Set("Content-Type", "application/x-protobuf")
	}))
	defer server.Close()
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://unused.invalid:4318")
	t.Setenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", server.URL+"/insert/opentelemetry/v1/traces")
	previous := otel.GetTracerProvider()
	defer otel.SetTracerProvider(previous)
	stop, err := Init(context.Background(), LoadConfig("test-service"))
	if err != nil {
		t.Fatal(err)
	}
	_, span := Tracer("test").Start(context.Background(), "request")
	span.End()
	if err := stop(context.Background()); err != nil {
		t.Fatal(err)
	}
	select {
	case got := <-requests:
		if got != "POST /insert/opentelemetry/v1/traces" {
			t.Fatalf("request = %s", got)
		}
	default:
		t.Fatal("no HTTP trace export")
	}
}
