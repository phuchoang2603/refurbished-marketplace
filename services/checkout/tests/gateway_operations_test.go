package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/phuchoang2603/refurbished-marketplace/services/checkout/internal/service"
	checkoutv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/checkout/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGatewayRecoveryReplaysFrozenRequest(t *testing.T) {
	db := newCheckoutDB(t)
	var bodies [][]byte
	var keys []string
	handler := func(w http.ResponseWriter, r *http.Request) {
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
		}
		if r.Header.Get("Authorization") != "Bearer test-api-key" || r.URL.Path != "/v1/payments" {
			t.Errorf("gateway contract headers/path mismatch")
		}
		bodies = append(bodies, raw)
		keys = append(keys, r.Header.Get("Idempotency-Key"))
		writePayment(t, w)
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	t.Cleanup(server.Close)
	svc := service.New(db, testConfig(server.URL), testAccounts{})
	request := checkoutRequest()
	result, err := svc.SubmitCheckout(t.Context(), request)
	if err != nil {
		t.Fatal(err)
	}
	var orderID string
	if err := db.QueryRowContext(t.Context(), "SELECT order_id FROM checkouts WHERE id=$1", result.GetCheckoutId()).Scan(&orderID); err != nil {
		t.Fatal(err)
	}
	reserveForPayment(t, svc, result.GetCheckoutId(), orderID)
	if _, err := db.ExecContext(t.Context(), `CREATE FUNCTION fail_ready() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN IF NEW.state='CHECKOUT_STATE_READY_TO_PAY' THEN RAISE EXCEPTION 'simulated lost local commit'; END IF; RETURN NEW; END $$; CREATE TRIGGER fail_ready BEFORE UPDATE ON checkouts FOR EACH ROW EXECUTE FUNCTION fail_ready()`); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ProcessGatewayOperation(t.Context()); err == nil {
		t.Fatal("expected local commit failure")
	}
	if _, err := db.ExecContext(t.Context(), "DROP TRIGGER fail_ready ON checkouts; UPDATE checkout_gateway_operations SET next_attempt_at=NOW()"); err != nil {
		t.Fatal(err)
	}
	recovered := service.New(db, testConfig(server.URL), testAccounts{})
	if _, err := recovered.ProcessGatewayOperation(t.Context()); err != nil {
		t.Fatal(err)
	}
	if len(bodies) != 2 || !bytes.Equal(bodies[0], bodies[1]) || keys[0] != orderID || keys[1] != orderID {
		t.Fatal("recovery must resend identical bytes and order key")
	}
	var snapshot struct {
		OrderReference string `json:"order_reference"`
		Seller         struct {
			Created time.Time `json:"account_created_at"`
		} `json:"seller"`
		Buyer map[string]any `json:"buyer"`
		Items []struct {
			Category  string `json:"category"`
			Condition string `json:"condition"`
			UnitPrice int64  `json:"unit_price"`
		} `json:"items"`
	}
	if err := json.Unmarshal(bodies[0], &snapshot); err != nil {
		t.Fatal(err)
	}
	if snapshot.OrderReference != orderID || !snapshot.Seller.Created.Equal(accountCreatedAt) || snapshot.Items[0].Category != "laptop" || snapshot.Items[0].Condition != "good" || snapshot.Items[0].UnitPrice != 1000 || snapshot.Buyer["email"] != nil {
		t.Fatalf("gateway snapshot mismatch")
	}
	assertCheckoutState(t, db, result.GetCheckoutId(), checkoutv1.CheckoutState_CHECKOUT_STATE_READY_TO_PAY, 2)
	view, err := svc.GetCheckout(t.Context(), uuid.MustParse(result.GetCheckoutId()))
	if err != nil || view.HostedCheckoutURL == "" {
		t.Fatalf("ready hosted URL %v %v", view, err)
	}
}

func TestGatewayDefectAndPreparationDeadlineCompensate(t *testing.T) {
	for _, failure := range []string{"conflict", "deadline", "unavailable"} {
		t.Run(failure, func(t *testing.T) {
			db := newCheckoutDB(t)
			var calls atomic.Int32
			svc := gatewayService(t, db, func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				if failure == "conflict" {
					w.WriteHeader(http.StatusConflict)
				} else {
					w.WriteHeader(http.StatusServiceUnavailable)
				}
			})
			checkoutID, orderID := submitWithService(t, db, svc)
			reserveForPayment(t, svc, checkoutID, orderID)
			if failure == "deadline" {
				if _, err := db.ExecContext(t.Context(), "UPDATE checkouts SET deadline_at=NOW()-INTERVAL '1 second' WHERE id=$1", checkoutID); err != nil {
					t.Fatal(err)
				}
			}
			if processed, err := svc.ProcessGatewayOperation(t.Context()); err != nil || !processed {
				t.Fatalf("operation %v %v", processed, err)
			}
			if failure == "unavailable" {
				assertCheckoutState(t, db, checkoutID, checkoutv1.CheckoutState_CHECKOUT_STATE_CREATING_PAYMENT, 2)
				if _, err := db.ExecContext(t.Context(), "UPDATE checkouts SET deadline_at=NOW()-INTERVAL '1 second' WHERE id=$1", checkoutID); err != nil {
					t.Fatal(err)
				}
				if _, err := svc.ProcessDue(t.Context()); err != nil {
					t.Fatal(err)
				}
			}
			assertCheckoutState(t, db, checkoutID, checkoutv1.CheckoutState_CHECKOUT_STATE_RELEASING_STOCK, 3)
			view, err := svc.GetCheckout(t.Context(), uuid.MustParse(checkoutID))
			if err != nil || view.HostedCheckoutURL != "" {
				t.Fatalf("failed prep exposes destination: %v %v", view, err)
			}
			if failure == "deadline" && calls.Load() != 0 {
				t.Fatal("expired preparation must not call gateway")
			}
		})
	}
}

func TestGatewayWorkersClaimOneOperation(t *testing.T) {
	db := newCheckoutDB(t)
	var calls atomic.Int32
	entered := make(chan struct{})
	release := make(chan struct{})
	svc := gatewayService(t, db, func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		close(entered)
		<-release
		writePayment(t, w)
	})
	checkoutID, orderID := submitWithService(t, db, svc)
	reserveForPayment(t, svc, checkoutID, orderID)
	var worker sync.WaitGroup
	worker.Go(func() {
		if _, err := svc.ProcessGatewayOperation(t.Context()); err != nil {
			t.Error(err)
		}
	})
	<-entered
	if processed, err := svc.ProcessGatewayOperation(t.Context()); err != nil || processed {
		t.Errorf("second worker claimed live operation: %v %v", processed, err)
	}
	close(release)
	worker.Wait()
	if calls.Load() != 1 {
		t.Fatalf("calls %d", calls.Load())
	}
}

func TestDisabledGatewayCreatesNothing(t *testing.T) {
	svc := service.New(nil, service.Config{}, nil)
	if _, err := svc.SubmitCheckout(t.Context(), checkoutRequest()); status.Code(err) != codes.Unavailable {
		t.Fatalf("disabled checkout: %v", err)
	}
}
