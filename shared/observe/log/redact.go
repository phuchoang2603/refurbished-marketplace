package log

import (
	"log/slog"
	"strings"
)

// Keys that must never appear in full in log attributes.
var sensitiveKeys = map[string]struct{}{
	"password":      {},
	"passwd":        {},
	"secret":        {},
	"token":         {},
	"authorization": {},
	"card":          {},
	"card_number":   {},
	"cardnumber":    {},
	"cvv":           {},
	"cvc":           {},
	"pan":           {},
}

func redactAttr(_ []string, a slog.Attr) slog.Attr {
	key := strings.ToLower(a.Key)
	if _, ok := sensitiveKeys[key]; ok {
		return slog.String(a.Key, "[redacted]")
	}
	return a
}

// Redacted returns a string attr that always logs as [redacted].
func Redacted(key string) slog.Attr {
	return slog.String(key, "[redacted]")
}
