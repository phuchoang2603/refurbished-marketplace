package tests

import (
	"database/sql"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/phuchoang2603/refurbished-marketplace/services/checkout/internal/grpcserver"
	authconfig "github.com/phuchoang2603/refurbished-marketplace/shared/auth/config"
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
		Currency: "USD", TotalCents: 2000,
		Items:           []*checkoutv1.CheckoutLine{{ProductId: uuid.NewString(), Name: "Laptop", Category: "laptop", Condition: "good", Quantity: 2, UnitPriceCents: 1000}},
		ShippingAddress: &checkoutv1.CheckoutAddress{Name: "Buyer", Line1: "123 Main St", City: "Seattle", Country: "US"},
		ReturnUrl:       "https://example.com/checkout/return",
	}
}

func TestSubmitIdempotencyAndOwnership(t *testing.T) {
	db := newCheckoutDB(t)
	server := grpcserver.New(newTestService(db), authconfig.DefaultConfig(checkoutTestJWTSecret))
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
	conflict.ShippingAddress.Line1 = "456 Other St"
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
	server := grpcserver.New(newTestService(db), authconfig.DefaultConfig(checkoutTestJWTSecret))
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

func TestCheckoutLookupByOrderIsBuyerOwned(t *testing.T) {
	db := newCheckoutDB(t)
	server := grpcserver.New(newTestService(db), authconfig.DefaultConfig(checkoutTestJWTSecret))
	request := checkoutRequest()
	ctx := authenticatedContext(t, request.GetBuyerUserId())
	submitted, err := server.SubmitCheckout(ctx, request)
	if err != nil {
		t.Fatal(err)
	}
	owned, err := server.GetCheckout(ctx, &checkoutv1.GetCheckoutRequest{CheckoutId: submitted.GetCheckoutId(), BuyerUserId: request.GetBuyerUserId()})
	if err != nil {
		t.Fatal(err)
	}
	byOrder, err := server.GetCheckout(ctx, &checkoutv1.GetCheckoutRequest{OrderId: owned.GetOrderId(), BuyerUserId: request.GetBuyerUserId()})
	if err != nil || !proto.Equal(owned, byOrder) {
		t.Fatalf("order lookup %v %v", byOrder, err)
	}
	other := uuid.NewString()
	_, err = server.GetCheckout(authenticatedContext(t, other), &checkoutv1.GetCheckoutRequest{OrderId: owned.GetOrderId(), BuyerUserId: other})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("order lookup ownership %v", err)
	}
	for _, state := range []checkoutv1.CheckoutState{
		checkoutv1.CheckoutState_CHECKOUT_STATE_READY_TO_PAY,
		checkoutv1.CheckoutState_CHECKOUT_STATE_COMPLETED,
	} {
		t.Run(state.String(), func(t *testing.T) {
			if _, err := db.ExecContext(t.Context(), "UPDATE checkouts SET state=$1, hosted_checkout_url=$2, gateway_payment_id=$3, payment_status=$4 WHERE id=$5", state.String(), "https://pay.example.com/session", "payment-1", "OPEN", submitted.GetCheckoutId()); err != nil {
				t.Fatal(err)
			}
			byID, err := server.GetCheckout(ctx, &checkoutv1.GetCheckoutRequest{CheckoutId: submitted.GetCheckoutId(), BuyerUserId: request.GetBuyerUserId()})
			if err != nil {
				t.Fatal(err)
			}
			byOrder, err := server.GetCheckout(ctx, &checkoutv1.GetCheckoutRequest{OrderId: owned.GetOrderId(), BuyerUserId: request.GetBuyerUserId()})
			if err != nil || !proto.Equal(byID, byOrder) {
				t.Fatalf("lookup views differ: byID=%v byOrder=%v err=%v", byID, byOrder, err)
			}
			if (byOrder.GetHostedCheckoutUrl() != "") != (state == checkoutv1.CheckoutState_CHECKOUT_STATE_READY_TO_PAY) {
				t.Fatalf("unexpected hosted URL for %v: %v", state, byOrder)
			}
		})
	}
}

func TestInvalidGradingRejectedBeforeCheckoutPersistence(t *testing.T) {
	svc := newTestService(nil)
	for _, tc := range []struct{ category, condition string }{
		{"", "good"},
		{"computer", "good"},
		{"Phone", "good"},
		{"phone ", "good"},
		{"phone", ""},
		{"phone", "new"},
		{"phone", "GOOD"},
		{"phone", "good "},
	} {
		t.Run(tc.category+"/"+tc.condition, func(t *testing.T) {
			request := checkoutRequest()
			request.Items[0].Category, request.Items[0].Condition = tc.category, tc.condition
			if _, err := svc.SubmitCheckout(t.Context(), request); status.Code(err) != codes.InvalidArgument {
				t.Fatalf("invalid grading error=%v", err)
			}
		})
	}
}
