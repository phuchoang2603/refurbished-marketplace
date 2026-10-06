package tests

import (
	"errors"
	"testing"

	"github.com/phuchoang2603/refurbished-marketplace/services/payment/internal/database"
	"github.com/phuchoang2603/refurbished-marketplace/services/payment/internal/service"
	testpostgres "github.com/phuchoang2603/refurbished-marketplace/shared/testutil/postgres"

	"github.com/google/uuid"
)

func newPaymentFixture(t *testing.T) (*service.Service, *database.Queries) {
	t.Helper()
	db := testpostgres.SetupPostgresWithMigrations(
		t,
		testpostgres.Config{
			Database: "payment_db",
			Username: "payment_app",
			Password: "payment_app_dev_password",
		},
		"../db/migrations",
	)
	queries := database.New(db)
	return service.New(db), queries
}

func TestPaymentService_ApplyGatewayWebhookRequiresCheckoutSession(t *testing.T) {
	svc, _ := newPaymentFixture(t)

	err := svc.ApplyGatewayWebhook(t.Context(), uuid.New(), "sess", service.HostedPaymentSessionStatusSucceeded, "")
	if !errors.Is(err, service.ErrIntentNotFound) {
		t.Fatalf("expected ErrIntentNotFound, got %v", err)
	}
}
