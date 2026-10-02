package runtime

import (
	"net/url"
	"testing"
)

func TestPostgresURLFromEnv(t *testing.T) {
	t.Setenv("DB_USER", "checkout@app")
	t.Setenv("DB_PASSWORD", "slash/colon:at@plus+hash#percent% space")
	t.Setenv("DB_HOST", "::1")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_NAME", "checkout_db")

	dbURL := PostgresURLFromEnv()
	parsed, err := url.Parse(dbURL)
	if err != nil {
		t.Fatalf("parse postgres URL: %v", err)
	}
	if parsed.User.Username() != "checkout@app" {
		t.Fatalf("unexpected username: %q", parsed.User.Username())
	}
	password, ok := parsed.User.Password()
	if !ok || password != "slash/colon:at@plus+hash#percent% space" {
		t.Fatal("postgres URL did not preserve the password")
	}
	if parsed.Host != "[::1]:5432" || parsed.Path != "/checkout_db" || parsed.Query().Get("sslmode") != "disable" {
		t.Fatalf("unexpected postgres URL host, database, or options: %s %s %s", parsed.Host, parsed.Path, parsed.RawQuery)
	}
}
