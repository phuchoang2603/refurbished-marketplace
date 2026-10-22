package service

import (
	"context"
	"encoding/json"
	"testing"
	"testing/synctest"
	"time"

	"github.com/google/uuid"
	checkoutv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/checkout/v1"
	usersv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/users/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type usersFunc func(context.Context, *usersv1.GetUserByIDRequest) (*usersv1.User, error)

func (lookup usersFunc) GetUserByID(ctx context.Context, request *usersv1.GetUserByIDRequest, _ ...grpc.CallOption) (*usersv1.User, error) {
	return lookup(ctx, request)
}

func TestGatewaySnapshotLooksUpAccountsConcurrently(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		created := map[string]time.Time{
			"seller": time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
			"buyer":  time.Date(2021, 3, 4, 0, 0, 0, 0, time.UTC),
		}
		started := make(chan string, 2)
		release := make(chan struct{})
		lookup := usersFunc(func(ctx context.Context, request *usersv1.GetUserByIDRequest) (*usersv1.User, error) {
			started <- request.GetId()
			select {
			case <-release:
				return &usersv1.User{Id: request.GetId(), CreatedAt: timestamppb.New(created[request.GetId()])}, nil
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		})
		svc := New(nil, Config{CheckoutReturnBaseURL: "https://example.com"}, lookup)
		request := &checkoutv1.SubmitCheckoutRequest{MerchantId: "seller", BuyerUserId: "buyer", ReturnUrl: "https://example.com/return"}
		var raw []byte
		var err error
		done := make(chan struct{})
		go func() {
			raw, err = svc.freezeGatewaySnapshot(t.Context(), request, uuid.New())
			close(done)
		}()
		first, second := <-started, <-started
		if first == second {
			t.Fatalf("expected both accounts, got %q twice", first)
		}
		close(release)
		<-done
		if err != nil {
			t.Fatal(err)
		}
		var snapshot gatewayRequest
		if err := json.Unmarshal(raw, &snapshot); err != nil {
			t.Fatal(err)
		}
		if snapshot.Seller.ExternalID != "seller" || !snapshot.Seller.AccountCreatedAt.Equal(created["seller"]) || snapshot.Buyer.ExternalID != "buyer" || !snapshot.Buyer.AccountCreatedAt.Equal(created["buyer"]) {
			t.Fatalf("account facts mixed up: seller=%+v buyer=%+v", snapshot.Seller, snapshot.Buyer)
		}
	})
}

func TestGatewaySnapshotCancelsSiblingOnInvalidAccount(t *testing.T) {
	for _, failingID := range []string{"seller", "buyer"} {
		t.Run(failingID, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				started := make(chan struct{})
				canceled := false
				lookup := usersFunc(func(ctx context.Context, request *usersv1.GetUserByIDRequest) (*usersv1.User, error) {
					if request.GetId() == failingID {
						<-started
						return nil, nil
					}
					close(started)
					<-ctx.Done()
					canceled = ctx.Err() == context.Canceled
					return nil, ctx.Err()
				})
				svc := New(nil, Config{}, lookup)
				raw, err := svc.freezeGatewaySnapshot(t.Context(), &checkoutv1.SubmitCheckoutRequest{MerchantId: "seller", BuyerUserId: "buyer"}, uuid.New())
				if status.Code(err) != codes.Unavailable || raw != nil || !canceled {
					t.Fatalf("expected unavailable and sibling cancellation: raw=%s err=%v canceled=%v", raw, err, canceled)
				}
			})
		})
	}
}

func TestGatewaySnapshotAccountTimeout(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		lookup := usersFunc(func(ctx context.Context, _ *usersv1.GetUserByIDRequest) (*usersv1.User, error) {
			<-ctx.Done()
			return nil, ctx.Err()
		})
		svc := New(nil, Config{}, lookup)
		start := time.Now()
		_, err := svc.freezeGatewaySnapshot(t.Context(), &checkoutv1.SubmitCheckoutRequest{MerchantId: "seller", BuyerUserId: "buyer"}, uuid.New())
		if status.Code(err) != codes.Unavailable || time.Since(start) != 5*time.Second {
			t.Fatalf("expected a shared five-second wait, got %v after %v", err, time.Since(start))
		}
	})
}
