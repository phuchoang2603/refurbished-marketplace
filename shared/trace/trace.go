package trace

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.39.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	defaultOTLPEndpoint = "vtsingle-vmks.monitoring.svc.cluster.local:4317"
	defaultSampleRatio  = 1.0
)

// Config controls the shared tracer provider. Empty Endpoint skips export
// (noop provider) so Tilt/local can run without VictoriaTraces.
type Config struct {
	ServiceName string
	Endpoint    string
	SampleRatio float64
	UseHTTP     bool
}

// LoadConfig reads OTEL_* / SERVICE_NAME style env vars.
func LoadConfig(defaultServiceName string) Config {
	serviceName := strings.TrimSpace(os.Getenv("OTEL_SERVICE_NAME"))
	if serviceName == "" {
		serviceName = strings.TrimSpace(defaultServiceName)
	}
	endpoint := strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))
	if endpoint == "" {
		endpoint = strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT"))
	}
	ratio := defaultSampleRatio
	if raw := strings.TrimSpace(os.Getenv("OTEL_TRACES_SAMPLER_ARG")); raw != "" {
		if parsed, err := strconv.ParseFloat(raw, 64); err == nil {
			ratio = parsed
		}
	}
	useHTTP := strings.EqualFold(strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL")), "http/protobuf") ||
		strings.Contains(endpoint, "/insert/opentelemetry/")
	return Config{
		ServiceName: serviceName,
		Endpoint:    endpoint,
		SampleRatio: ratio,
		UseHTTP:     useHTTP,
	}
}

// Init installs the global TracerProvider and W3C propagators.
// Returns a shutdown func. If Endpoint is empty, installs a noop provider.
func Init(ctx context.Context, cfg Config) (func(context.Context) error, error) {
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	if strings.TrimSpace(cfg.ServiceName) == "" {
		return func(context.Context) error { return nil }, fmt.Errorf("trace: service name is required")
	}
	if strings.TrimSpace(cfg.Endpoint) == "" {
		tp := sdktrace.NewTracerProvider()
		otel.SetTracerProvider(tp)
		return tp.Shutdown, nil
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(cfg.ServiceName),
		),
	)
	if err != nil {
		return nil, err
	}

	ratio := cfg.SampleRatio
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}

	var exp sdktrace.SpanExporter
	if cfg.UseHTTP {
		endpoint := cfg.Endpoint
		if !strings.Contains(endpoint, "/insert/opentelemetry/") {
			endpoint = "http://" + strings.TrimPrefix(strings.TrimPrefix(endpoint, "http://"), "https://")
			endpoint = strings.TrimRight(endpoint, "/") + "/insert/opentelemetry/v1/traces"
		}
		exp, err = otlptracehttp.New(ctx, otlptracehttp.WithEndpointURL(endpoint))
	} else {
		endpoint := strings.TrimPrefix(strings.TrimPrefix(cfg.Endpoint, "http://"), "https://")
		endpoint = strings.TrimSuffix(endpoint, "/insert/opentelemetry/v1/traces")
		exp, err = otlptracegrpc.New(
			ctx,
			otlptracegrpc.WithEndpoint(endpoint),
			otlptracegrpc.WithInsecure(),
		)
	}
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp, sdktrace.WithBatchTimeout(2*time.Second)),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(ratio))),
	)
	otel.SetTracerProvider(tp)
	return tp.Shutdown, nil
}

// DefaultEndpoint is the in-cluster VictoriaTraces OTLP gRPC address.
func DefaultEndpoint() string { return defaultOTLPEndpoint }

// Tracer returns a named tracer from the global provider.
func Tracer(name string) trace.Tracer {
	return otel.Tracer(name)
}

// SerializeContext encodes the active span context as Java Properties text
// (key=value lines) for Debezium EventRouter tracingspancontext.
func SerializeContext(ctx context.Context) string {
	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	if len(carrier) == 0 {
		return ""
	}
	var b strings.Builder
	for k, v := range carrier {
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(v)
		b.WriteByte('\n')
	}
	return b.String()
}

// ContextFromHeaders restores parent context from Kafka/W3C header map.
func ContextFromHeaders(ctx context.Context, headers map[string]string) context.Context {
	if len(headers) == 0 {
		return ctx
	}
	carrier := propagation.MapCarrier{}
	for k, v := range headers {
		carrier[strings.ToLower(k)] = v
	}
	return otel.GetTextMapPropagator().Extract(ctx, carrier)
}
