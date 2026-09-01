package runtime

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"regexp"
	"strings"

	"github.com/XSAM/otelsql"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
)

const maxDBQueryTextLen = 1024

// sqlcNameRe matches `-- name: GetUserByEmail :one` that sqlc embeds in queries.
var sqlcNameRe = regexp.MustCompile(`(?m)^--\s*name:\s*(\S+)`)

// OpenPostgres opens an instrumented *sql.DB for Postgres.
// The caller must import a postgres driver (e.g. github.com/lib/pq).
// Query spans nest under the active request/consumer context; statement
// text is truncated and bound args are never recorded. Span names prefer
// the sqlc `-- name:` comment when present.
func OpenPostgres(dbURL string) (*sql.DB, error) {
	db, err := otelsql.Open(
		"postgres", dbURL,
		otelsql.WithAttributes(semconv.DBSystemNamePostgreSQL),
		otelsql.WithSpanOptions(otelsql.SpanOptions{
			DisableErrSkip:       true,
			DisableQuery:         true, // set truncated db.query.text via AttributesGetter
			OmitConnResetSession: true,
			OmitConnectorConnect: true,
			OmitRows:             true,
		}),
		otelsql.WithSpanNameFormatter(dbSpanName),
		otelsql.WithAttributesGetter(truncatedDBQueryAttrs),
	)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}
	return db, nil
}

func dbSpanName(_ context.Context, method otelsql.Method, query string) string {
	if name := sqlcQueryName(query); name != "" {
		return name
	}
	if q := strings.TrimSpace(query); q != "" {
		// First SQL keyword keeps non-sqlc queries readable without full text.
		if fields := strings.Fields(q); len(fields) > 0 {
			return strings.ToUpper(fields[0])
		}
	}
	return string(method)
}

func sqlcQueryName(query string) string {
	m := sqlcNameRe.FindStringSubmatch(query)
	if len(m) < 2 {
		return ""
	}
	return m[1]
}

func truncatedDBQueryAttrs(_ context.Context, _ otelsql.Method, query string, args []driver.NamedValue) []attribute.KeyValue {
	_ = args // never record bound parameter values
	if query == "" {
		return nil
	}
	q := query
	if len(q) > maxDBQueryTextLen {
		q = q[:maxDBQueryTextLen] + "…"
	}
	attrs := []attribute.KeyValue{attribute.String("db.query.text", q)}
	if name := sqlcQueryName(query); name != "" {
		attrs = append(attrs, attribute.String("db.operation.name", name))
	}
	return attrs
}
