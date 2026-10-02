package tests

import (
	"database/sql"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/phuchoang2603/refurbished-marketplace/services/checkout/internal/grpcserver"
	"github.com/phuchoang2603/refurbished-marketplace/services/checkout/internal/service"
	checkoutv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/checkout/v1"
	testpostgres "github.com/phuchoang2603/refurbished-marketplace/shared/testutil/postgres"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func newCheckoutDB(t *testing.T) *sql.DB {
	t.Helper()
	return testpostgres.SetupPostgresWithMigrations(t, testpostgres.Config{
		Database: "checkout_test", Username: "checkout_test", Password: "checkout_test",
	}, "../db/migrations")
}

func checkoutRequest() *checkoutv1.SubmitCheckoutRequest {
	return &checkoutv1.SubmitCheckoutRequest{
		BuyerUserId: uuid.NewString(), MerchantId: uuid.NewString(), IntentKey: uuid.NewString(),
		BuyerEmail: "buyer@example.com", Currency: "USD", TotalCents: 2000,
		Items:           []*checkoutv1.CheckoutLine{{ProductId: uuid.NewString(), Name: "Laptop", Quantity: 2, UnitPriceCents: 1000}},
		ShippingAddress: &checkoutv1.CheckoutAddress{Name: "Buyer", Country: "US"},
		ReturnUrl:       "https://example.com/checkout/return",
	}
}

func TestSubmitIdempotencyAndOwnership(t *testing.T) {
	db := newCheckoutDB(t)
	server := grpcserver.New(service.New(db))
	request := checkoutRequest()
	buyerContext := authenticatedContext(t, request.GetBuyerUserId())
	first, err := server.SubmitCheckout(buyerContext, request)
	if err != nil || first.GetCheckoutId() == "" {
		t.Fatalf("submit: %v, %v", first, err)
	}
	again, err := server.SubmitCheckout(buyerContext, request)
	if err != nil || first.GetCheckoutId() != again.GetCheckoutId() {
		t.Fatalf("idempotent retry: %v, %v", again, err)
	}
	var checkoutCount, outboxCount int
	if err := db.QueryRowContext(t.Context(), "SELECT count(*) FROM checkouts").Scan(&checkoutCount); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(t.Context(), "SELECT count(*) FROM checkout_outbox").Scan(&outboxCount); err != nil {
		t.Fatal(err)
	}
	if checkoutCount != 1 || outboxCount != 1 {
		t.Fatalf("expected one checkout and command, got %d and %d", checkoutCount, outboxCount)
	}
	conflict := proto.Clone(request).(*checkoutv1.SubmitCheckoutRequest)
	conflict.BuyerEmail = "other@example.com"
	_, err = server.SubmitCheckout(buyerContext, conflict)
	if status.Code(err) != codes.AlreadyExists {
		t.Fatalf("conflicting key should be rejected: %v", err)
	}
	_, err = server.GetCheckout(buyerContext, &checkoutv1.GetCheckoutRequest{
		CheckoutId: first.GetCheckoutId(), BuyerUserId: uuid.NewString(),
	})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("other buyer should not access status: %v", err)
	}
	owned, err := server.GetCheckout(buyerContext, &checkoutv1.GetCheckoutRequest{
		CheckoutId: first.GetCheckoutId(), BuyerUserId: request.GetBuyerUserId(),
	})
	if err != nil || owned.GetState() != checkoutv1.CheckoutState_CHECKOUT_STATE_CREATING_ORDER || owned.GetOrderId() == "" {
		t.Fatalf("owner should see pending order: %v, %v", owned, err)
	}
}

func TestConcurrentSubmitEmitsOneOrder(t *testing.T) {
	db := newCheckoutDB(t)
	server := grpcserver.New(service.New(db))
	request := checkoutRequest()
	buyerContext := authenticatedContext(t, request.GetBuyerUserId())
	var wait sync.WaitGroup
	results := make(chan string, 8)
	for range 8 {
		wait.Go(func() {
			result, err := server.SubmitCheckout(buyerContext, request)
			if err != nil {
				results <- err.Error()
				return
			}
			results <- result.GetCheckoutId()
		})
	}
	wait.Wait()
	close(results)
	var checkoutID string
	for result := range results {
		if checkoutID == "" {
			checkoutID = result
		} else if result != checkoutID {
			t.Fatalf("concurrent requests produced multiple outcomes: %s vs %s", checkoutID, result)
		}
	}
	if _, err := uuid.Parse(checkoutID); err != nil {
		t.Fatalf("invalid checkout id: %s", checkoutID)
	}
	var outboxCount int
	if err := db.QueryRowContext(t.Context(), "SELECT count(*) FROM checkout_outbox").Scan(&outboxCount); err != nil || outboxCount != 1 {
		t.Fatalf("expected one command, got %d: %v", outboxCount, err)
	}
}
