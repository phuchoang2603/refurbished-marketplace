package runtime

import (
	"context"
	"database/sql"
	"os"
	"sync"
	"testing"
	"testing/fstest"
	"time"
)

func TestPostgresMigrationsConcurrentStartup(t *testing.T) {
	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("TEST_POSTGRES_DSN must point to a disposable PostgreSQL database")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	if _, err := db.ExecContext(ctx, `CREATE ROLE migration_test_runtime NOLOGIN`); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = db.Exec(`DROP TABLE IF EXISTS migration_test, migration_future, goose_db_version; DROP OWNED BY migration_test_runtime; DROP ROLE migration_test_runtime`)
	}()

	migrations := fstest.MapFS{
		"001_test.sql": {Data: []byte("-- +goose Up\nCREATE TABLE migration_test (id bigserial PRIMARY KEY, value text);\nINSERT INTO migration_test (value) VALUES ('once');\nSELECT pg_sleep(0.1);\n-- +goose Down\nDROP TABLE migration_test;\n")},
	}
	var wg sync.WaitGroup
	for range 3 {
		wg.Go(func() {
			if err := applyPostgresMigrations(ctx, db, migrations, "migration_test_runtime"); err != nil {
				t.Errorf("concurrent migration: %v", err)
			}
		})
	}
	wg.Wait()
	if t.Failed() {
		return
	}
	var count int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM migration_test`).Scan(&count); err != nil || count != 1 {
		t.Fatalf("migration must run exactly once: count=%d, err=%v", count, err)
	}
	if _, err := db.ExecContext(ctx, `CREATE TABLE migration_future (id bigserial PRIMARY KEY)`); err != nil {
		t.Fatal(err)
	}
	checks := map[string]bool{
		`has_table_privilege('migration_test_runtime', 'migration_test', 'SELECT')`:            true,
		`has_table_privilege('migration_test_runtime', 'migration_test', 'INSERT')`:            true,
		`has_table_privilege('migration_test_runtime', 'migration_test', 'UPDATE')`:            true,
		`has_table_privilege('migration_test_runtime', 'migration_test', 'DELETE')`:            true,
		`has_sequence_privilege('migration_test_runtime', 'migration_test_id_seq', 'USAGE')`:   true,
		`has_table_privilege('migration_test_runtime', 'migration_future', 'INSERT')`:          true,
		`has_sequence_privilege('migration_test_runtime', 'migration_future_id_seq', 'USAGE')`: true,
		`has_table_privilege('migration_test_runtime', 'goose_db_version', 'SELECT')`:          false,
		`has_table_privilege('migration_test_runtime', 'goose_db_version', 'UPDATE')`:          false,
		`has_schema_privilege('migration_test_runtime', 'public', 'CREATE')`:                   false,
	}
	for expression, want := range checks {
		var got bool
		if err := db.QueryRowContext(ctx, "SELECT "+expression).Scan(&got); err != nil || got != want {
			t.Errorf("%s: got=%v want=%v err=%v", expression, got, want, err)
		}
	}
}
