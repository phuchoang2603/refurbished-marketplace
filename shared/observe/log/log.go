package log

import (
	"context"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel/trace"
)

// Init configures the default slog logger as JSON to stdout with a service attribute
// and automatic trace_id / span_id injection from the active OTEL span on *Context calls.
func Init(serviceName string) {
	base := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		ReplaceAttr: redactAttr,
	})
	handler := &traceHandler{Handler: base.WithAttrs([]slog.Attr{
		slog.String("service", serviceName),
	})}
	slog.SetDefault(slog.New(handler))
}

// Fatal logs at Error level and exits with status 1.
func Fatal(msg string, args ...any) {
	slog.Error(msg, args...)
	os.Exit(1)
}

// FatalContext logs at Error level with context attrs and exits with status 1.
func FatalContext(ctx context.Context, msg string, args ...any) {
	slog.ErrorContext(ctx, msg, args...)
	os.Exit(1)
}

type traceHandler struct {
	slog.Handler
}

func (h *traceHandler) Handle(ctx context.Context, r slog.Record) error {
	if sc := trace.SpanFromContext(ctx).SpanContext(); sc.IsValid() {
		r.AddAttrs(
			slog.String("trace_id", sc.TraceID().String()),
			slog.String("span_id", sc.SpanID().String()),
		)
	}
	return h.Handler.Handle(ctx, r)
}

func (h *traceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &traceHandler{Handler: h.Handler.WithAttrs(attrs)}
}

func (h *traceHandler) WithGroup(name string) slog.Handler {
	return &traceHandler{Handler: h.Handler.WithGroup(name)}
}
