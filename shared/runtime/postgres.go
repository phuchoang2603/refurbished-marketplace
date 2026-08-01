package runtime

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"

	"github.com/XSAM/otelsql"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"
)

const maxDBQueryTextLen = 1024

// OpenPostgres opens an instrumented *sql.DB for Postgres.
// The caller must import a postgres driver (e.g. github.com/lib/pq).
// Query spans nest under the active request/consumer context; statement
// text is truncated and bound args are never recorded.
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

func truncatedDBQueryAttrs(_ context.Context, _ otelsql.Method, query string, args []driver.NamedValue) []attribute.KeyValue {
	_ = args // never record bound parameter values
	if query == "" {
		return nil
	}
	q := query
	if len(q) > maxDBQueryTextLen {
		q = q[:maxDBQueryTextLen] + "…"
	}
	return []attribute.KeyValue{attribute.String("db.query.text", q)}
}
