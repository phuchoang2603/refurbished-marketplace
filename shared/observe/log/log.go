package log

import (
	"context"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel/trace"
)

// Common attribute keys for cross-service LogSQL filters.
const (
	KeyService     = "service"
	KeyTraceID     = "trace_id"
	KeySpanID      = "span_id"
	KeyOrderID     = "order_id"
	KeyMerchantID  = "merchant_id"
	KeyBuyerUserID = "buyer_user_id"
	KeyStatus      = "status"
	KeyOutcome     = "outcome"
	KeyEventType   = "event_type"
	KeyErr         = "err"
)

// Init configures the process default logger as JSON to stdout with a service
// attribute and automatic trace_id / span_id injection on *Context calls.
func Init(serviceName string) {
	base := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		ReplaceAttr: replaceAttr,
	})
	handler := &traceHandler{Handler: base.WithAttrs([]slog.Attr{
		slog.String(KeyService, serviceName),
	})}
	slog.SetDefault(slog.New(handler))
}

// Default returns the process logger (after Init). Prefer package helpers
// (InfoContext, etc.) or Default().With(...) for child loggers.
func Default() *slog.Logger {
	return slog.Default()
}

// With returns a child logger with additional attributes (slog best practice
// for stable fields shared across a subsystem).
func With(args ...any) *slog.Logger {
	return slog.Default().With(args...)
}

func Info(msg string, args ...any) {
	slog.Default().Info(msg, args...)
}

func InfoContext(ctx context.Context, msg string, args ...any) {
	slog.Default().InfoContext(ctx, msg, args...)
}

func Warn(msg string, args ...any) {
	slog.Default().Warn(msg, args...)
}

func WarnContext(ctx context.Context, msg string, args ...any) {
	slog.Default().WarnContext(ctx, msg, args...)
}

func Error(msg string, args ...any) {
	slog.Default().Error(msg, args...)
}

func ErrorContext(ctx context.Context, msg string, args ...any) {
	slog.Default().ErrorContext(ctx, msg, args...)
}

func LogAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	slog.Default().LogAttrs(ctx, level, msg, attrs...)
}

// Fatal logs at Error level and exits with status 1.
func Fatal(msg string, args ...any) {
	Error(msg, args...)
	os.Exit(1)
}

// FatalContext logs at Error level with context attrs and exits with status 1.
func FatalContext(ctx context.Context, msg string, args ...any) {
	ErrorContext(ctx, msg, args...)
	os.Exit(1)
}

// Attr helpers keep domain field names consistent across services.
func AttrOrderID(id string) slog.Attr     { return slog.String(KeyOrderID, id) }
func AttrMerchantID(id string) slog.Attr  { return slog.String(KeyMerchantID, id) }
func AttrBuyerUserID(id string) slog.Attr { return slog.String(KeyBuyerUserID, id) }
func AttrStatus(status string) slog.Attr  { return slog.String(KeyStatus, status) }
func AttrOutcome(outcome string) slog.Attr {
	return slog.String(KeyOutcome, outcome)
}
func AttrEventType(t string) slog.Attr { return slog.String(KeyEventType, t) }
func AttrErr(err error) slog.Attr      { return slog.Any(KeyErr, err) }

type traceHandler struct {
	slog.Handler
}

func (h *traceHandler) Handle(ctx context.Context, r slog.Record) error {
	if sc := trace.SpanFromContext(ctx).SpanContext(); sc.IsValid() {
		r.AddAttrs(
			slog.String(KeyTraceID, sc.TraceID().String()),
			slog.String(KeySpanID, sc.SpanID().String()),
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
