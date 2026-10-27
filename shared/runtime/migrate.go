package runtime

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"time"

	"github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"
)

func MigratePostgres(ctx context.Context, migrations fs.FS) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	db, err := sql.Open("postgres", "")
	if err != nil {
		return fmt.Errorf("open migration database: %w", err)
	}
	defer func() { _ = db.Close() }()

	return applyPostgresMigrations(ctx, db, migrations, os.Getenv("PG_RUNTIME_ROLE"))
}

func applyPostgresMigrations(ctx context.Context, db *sql.DB, migrations fs.FS, runtimeRole string) error {
	locker, err := lock.NewPostgresSessionLocker()
	if err != nil {
		return fmt.Errorf("create migration locker: %w", err)
	}
	provider, err := goose.NewProvider(goose.DialectPostgres, db, migrations, goose.WithSessionLocker(locker))
	if err != nil {
		return fmt.Errorf("create migration provider: %w", err)
	}
	if _, err := provider.Up(ctx); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}
	if runtimeRole == "" {
		return nil
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin runtime grants: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, lock.DefaultLockID); err != nil {
		return fmt.Errorf("lock runtime grants: %w", err)
	}

	role := pq.QuoteIdentifier(runtimeRole)
	statements := []string{
		`REVOKE CREATE ON SCHEMA public FROM PUBLIC`,
		`GRANT USAGE ON SCHEMA public TO ` + role,
		`GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO ` + role,
		`GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO ` + role,
		`ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO ` + role,
		`ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT USAGE, SELECT ON SEQUENCES TO ` + role,
		`REVOKE ALL ON TABLE public.goose_db_version FROM ` + role,
	}
	for _, statement := range statements {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("grant runtime database access: %w", err)
		}
	}
	return tx.Commit()
}
