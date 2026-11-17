package log

import (
	"log/slog"
	"strings"
)

var sensitiveKeys = map[string]struct{}{
	"password":      {},
	"passwd":        {},
	"secret":        {},
	"token":         {},
	"authorization": {},
	"api_key":       {},
	"apikey":        {},
	"bearer":        {},
	"access_token":  {},
	"refresh_token": {},
	"client_secret": {},
	"private_key":   {},
	"card":          {},
	"card_number":   {},
	"cardnumber":    {},
	"cvv":           {},
	"cvc":           {},
	"pan":           {},
}

func replaceAttr(_ []string, a slog.Attr) slog.Attr {
	if a.Key == slog.LevelKey {
		return slog.String(a.Key, strings.ToLower(a.Value.String()))
	}
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
