package tests

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/phuchoang2603/refurbished-marketplace/services/users/internal/service"
	"github.com/phuchoang2603/refurbished-marketplace/shared/messaging"
	testpostgres "github.com/phuchoang2603/refurbished-marketplace/shared/testutil/postgres"
	"github.com/pressly/goose/v3"
)

func TestAccountOutboxBackfillsAndPublishesTransactionally(t *testing.T) {
	db := testpostgres.SetupPostgresWithMigrations(t, testpostgres.Config{Database: "users_db", Username: "users_app", Password: "users_app_dev_password"}, "../db/migrations")
	if err := goose.Down(db, "../db/migrations"); err != nil {
		t.Fatal(err)
	}
	svc := service.New(db, service.DefaultConfig("test-secret"))
	existing, err := svc.CreateUser(t.Context(), "existing@example.com", "password123")
	if err != nil {
		t.Fatal(err)
	}
	if err := goose.Up(db, "../db/migrations"); err != nil {
		t.Fatal(err)
	}
	created, err := svc.CreateUser(t.Context(), "created@example.com", "password123")
	if err != nil {
		t.Fatal(err)
	}
	for _, user := range []service.User{existing, created} {
		var payload []byte
		var eventType string
		if err := db.QueryRowContext(t.Context(), "SELECT event_type, payload FROM users_outbox WHERE aggregate_id = $1", user.ID).Scan(&eventType, &payload); err != nil {
			t.Fatal(err)
		}
		var event messaging.AccountCreated
		if err := json.Unmarshal(payload, &event); err != nil {
			t.Fatal(err)
		}
		if eventType != messaging.EventTypeAccountCreated || event.UserID != user.ID.String() || !event.CreatedAt.Equal(user.CreatedAt) {
			t.Fatalf("incorrect account event: type=%s event=%+v", eventType, event)
		}
	}
	tx, err := db.BeginTx(t.Context(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()
	rolledBack := uuid.New()
	if _, err := tx.ExecContext(t.Context(), "INSERT INTO users (id, email, password_hash) VALUES ($1, $2, $3)", rolledBack, "rolled-back@example.com", "hash"); err != nil {
		t.Fatal(err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRowContext(t.Context(), "SELECT COUNT(*) FROM users_outbox WHERE aggregate_id = $1", rolledBack).Scan(&count); err != nil || count != 0 {
		t.Fatalf("rolled-back account published: count=%d err=%v", count, err)
	}
}
