package tests

import (
	"errors"
	"testing"

	"refurbished-marketplace/services/users/internal/database"
	"refurbished-marketplace/services/users/internal/service"
	"refurbished-marketplace/shared/testutil"

	"github.com/google/uuid"
)

func newUserService(t *testing.T) *service.Service {
	t.Helper()
	db := testutil.SetupPostgresWithMigrations(
		t,
		testutil.PostgresConfig{
			Database: "users_db",
			Username: "users_app",
			Password: "users_app_dev_password",
		},
		"../db/migrations",
	)

	return service.New(database.New(db), service.DefaultConfig("test-secret"))
}

func TestAuthLoginAndRefresh(t *testing.T) {
	svc := newUserService(t)
	ctx := t.Context()

	created, err := svc.CreateUser(ctx, "auth@test.com", "password123", 12.5, -4.25)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	t.Run("login success", func(t *testing.T) {
		tokens, err := svc.Login(ctx, "auth@test.com", "password123")
		if err != nil {
			t.Fatalf("login failed: %v", err)
		}
		if tokens.AccessToken == "" || tokens.RefreshToken == "" {
			t.Fatalf("expected non-empty tokens")
		}
	})

	t.Run("login invalid credentials", func(t *testing.T) {
		_, err := svc.Login(ctx, "auth@test.com", "wrong-password")
		if !errors.Is(err, service.ErrInvalidCredentials) {
			t.Fatalf("expected ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("refresh rotates token", func(t *testing.T) {
		tokens, err := svc.Login(ctx, "auth@test.com", "password123")
		if err != nil {
			t.Fatalf("login failed: %v", err)
		}

		refreshed, err := svc.Refresh(ctx, tokens.RefreshToken)
		if err != nil {
			t.Fatalf("refresh failed: %v", err)
		}

		if refreshed.AccessToken == "" || refreshed.RefreshToken == "" {
			t.Fatalf("expected refreshed tokens")
		}

		if refreshed.RefreshToken == tokens.RefreshToken {
			t.Fatalf("expected rotated refresh token")
		}
	})

	t.Run("refresh old token revoked", func(t *testing.T) {
		tokens, err := svc.Login(ctx, "auth@test.com", "password123")
		if err != nil {
			t.Fatalf("login failed: %v", err)
		}

		_, err = svc.Refresh(ctx, tokens.RefreshToken)
		if err != nil {
			t.Fatalf("first refresh failed: %v", err)
		}

		_, err = svc.Refresh(ctx, tokens.RefreshToken)
		if !errors.Is(err, service.ErrTokenRevoked) {
			t.Fatalf("expected ErrTokenRevoked, got %v", err)
		}
	})

	t.Run("logout revokes refresh token", func(t *testing.T) {
		tokens, err := svc.Login(ctx, "auth@test.com", "password123")
		if err != nil {
			t.Fatalf("login failed: %v", err)
		}

		if err := svc.Logout(ctx, tokens.RefreshToken); err != nil {
			t.Fatalf("logout failed: %v", err)
		}

		_, err = svc.Refresh(ctx, tokens.RefreshToken)
		if !errors.Is(err, service.ErrTokenRevoked) {
			t.Fatalf("expected ErrTokenRevoked, got %v", err)
		}
	})

	t.Run("get user by id", func(t *testing.T) {
		u, err := svc.GetUserByID(ctx, created.ID)
		if err != nil {
			t.Fatalf("get user by id: %v", err)
		}
		if u.ID != created.ID {
			t.Fatalf("expected id %s, got %s", created.ID, u.ID)
		}
	})

	t.Run("get missing user", func(t *testing.T) {
		_, err := svc.GetUserByID(ctx, uuid.New())
		if !errors.Is(err, service.ErrUserNotFound) {
			t.Fatalf("expected ErrUserNotFound, got %v", err)
		}
	})
}

func TestServiceCreateUserValidation(t *testing.T) {
	t.Run("invalid email", func(t *testing.T) {
		svc := newUserService(t)
		ctx := t.Context()

		_, err := svc.CreateUser(ctx, "bad-email", "password123", 0, 0)
		if !errors.Is(err, service.ErrInvalidEmail) {
			t.Fatalf("expected ErrInvalidEmail, got %v", err)
		}
	})

	t.Run("invalid password", func(t *testing.T) {
		svc := newUserService(t)
		ctx := t.Context()

		_, err := svc.CreateUser(ctx, "user@test.com", "short", 0, 0)
		if !errors.Is(err, service.ErrInvalidPassword) {
			t.Fatalf("expected ErrInvalidPassword, got %v", err)
		}
	})
}
