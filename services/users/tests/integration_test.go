package tests

import (
	"database/sql"
	"errors"
	"testing"

	"refurbished-marketplace/services/users/internal/database"
	"refurbished-marketplace/shared/testutil"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

func TestUsersQueries(t *testing.T) {
	db := testutil.SetupPostgresWithMigrations(
		t,
		testutil.PostgresConfig{
			Database: "users_db",
			Username: "users_app",
			Password: "users_app_dev_password",
		},
		"../db/migrations",
	)

	ctx := t.Context()
	queries := database.New(db)

	t.Run("create and read user", func(t *testing.T) {
		email := "readback@test.com"
		id := uuid.New()

		created, err := queries.CreateUser(ctx, database.CreateUserParams{
			ID:           id,
			Email:        email,
			PasswordHash: "hash",
		})
		if err != nil {
			t.Fatalf("create user failed: %v", err)
		}

		if created.ID != id {
			t.Fatalf("expected id %s, got %s", id, created.ID)
		}

		byID, err := queries.GetUserByID(ctx, id)
		if err != nil {
			t.Fatalf("get by id failed: %v", err)
		}
		if byID.Email != email {
			t.Fatalf("expected email %s, got %s", email, byID.Email)
		}

		byEmail, err := queries.GetUserByEmail(ctx, email)
		if err != nil {
			t.Fatalf("get by email failed: %v", err)
		}
		if byEmail.ID != id {
			t.Fatalf("expected id %s, got %s", id, byEmail.ID)
		}
	})

	t.Run("get missing user returns no rows", func(t *testing.T) {
		_, err := queries.GetUserByID(ctx, uuid.New())
		if !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("expected sql.ErrNoRows, got %v", err)
		}
	})

	t.Run("uniqueness constraint", func(t *testing.T) {
		id := uuid.New()
		_, err := queries.CreateUser(ctx, database.CreateUserParams{
			ID:           id,
			Email:        "duplicate@test.com",
			PasswordHash: "hash",
		})
		if err != nil {
			t.Fatal(err)
		}

		_, err = queries.CreateUser(ctx, database.CreateUserParams{
			ID:           uuid.New(),
			Email:        "duplicate@test.com",
			PasswordHash: "hash2",
		})

		var pqErr *pq.Error
		if !errors.As(err, &pqErr) || pqErr.Code != "23505" {
			t.Errorf("expected unique violation, got %v", err)
		}
	})
}
