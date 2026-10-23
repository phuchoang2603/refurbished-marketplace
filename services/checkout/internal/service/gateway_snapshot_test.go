package service

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"testing"
	"testing/synctest"
	"time"

	"github.com/google/uuid"
	checkoutv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/checkout/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type accountsFunc func(context.Context, ...string) ([]time.Time, error)

func (lookup accountsFunc) CreatedAt(ctx context.Context, ids ...string) ([]time.Time, error) {
	return lookup(ctx, ids...)
}

func TestGatewaySnapshotReadsBothProjectedAccounts(t *testing.T) {
	created := []time.Time{
		time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
		time.Date(2021, 3, 4, 0, 0, 0, 0, time.UTC),
	}
	calls := 0
	lookup := accountsFunc(func(_ context.Context, ids ...string) ([]time.Time, error) {
		calls++
		if !slices.Equal(ids, []string{"seller", "buyer"}) {
			t.Fatalf("unexpected accounts: %v", ids)
		}
		return created, nil
	})
	svc := New(nil, Config{CheckoutReturnBaseURL: "https://example.com"}, lookup)
	raw, err := svc.freezeGatewaySnapshot(t.Context(), &checkoutv1.SubmitCheckoutRequest{
		MerchantId: "seller", BuyerUserId: "buyer", ReturnUrl: "https://example.com/return",
	}, uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	var snapshot gatewayRequest
	if err := json.Unmarshal(raw, &snapshot); err != nil {
		t.Fatal(err)
	}
	if calls != 1 || !snapshot.Seller.AccountCreatedAt.Equal(created[0]) || !snapshot.Buyer.AccountCreatedAt.Equal(created[1]) {
		t.Fatalf("account facts mixed up: calls=%d seller=%+v buyer=%+v", calls, snapshot.Seller, snapshot.Buyer)
	}
}

func TestGatewaySnapshotRefusesMissingAccountFacts(t *testing.T) {
	for _, times := range [][]time.Time{nil, {time.Now()}, {time.Now(), {}}, {{}, time.Now()}} {
		svc := New(nil, Config{}, accountsFunc(func(context.Context, ...string) ([]time.Time, error) {
			return times, nil
		}))
		raw, err := svc.freezeGatewaySnapshot(t.Context(), &checkoutv1.SubmitCheckoutRequest{}, uuid.New())
		if status.Code(err) != codes.Unavailable || raw != nil {
			t.Fatalf("expected unavailable for %v, got %s / %v", times, raw, err)
		}
	}
	for _, lookup := range []Accounts{nil, accountsFunc(func(context.Context, ...string) ([]time.Time, error) {
		return nil, errors.New("redis unavailable")
	})} {
		svc := New(nil, Config{}, lookup)
		_, err := svc.freezeGatewaySnapshot(t.Context(), &checkoutv1.SubmitCheckoutRequest{}, uuid.New())
		if status.Code(err) != codes.Unavailable {
			t.Fatalf("expected unavailable, got %v", err)
		}
	}
}

func TestGatewaySnapshotAccountTimeout(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		lookup := accountsFunc(func(ctx context.Context, _ ...string) ([]time.Time, error) {
			<-ctx.Done()
			return nil, ctx.Err()
		})
		svc := New(nil, Config{}, lookup)
		start := time.Now()
		_, err := svc.freezeGatewaySnapshot(t.Context(), &checkoutv1.SubmitCheckoutRequest{}, uuid.New())
		if status.Code(err) != codes.Unavailable || time.Since(start) != 2*time.Second {
			t.Fatalf("expected a two-second wait, got %v after %v", err, time.Since(start))
		}
	})
}
