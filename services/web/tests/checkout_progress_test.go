package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/phuchoang2603/refurbished-marketplace/services/web/internal/auth"
	"github.com/phuchoang2603/refurbished-marketplace/services/web/tests/fakes"
	checkoutv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/checkout/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCheckoutProgressWaitsForSessionAndRestrictsBuyer(t *testing.T) {
	ready := false
	checkoutSvc := &fakes.CheckoutService{
		GetFn: func(_ context.Context, request *checkoutv1.GetCheckoutRequest) (*checkoutv1.CheckoutStatus, error) {
			if request.GetBuyerUserId() != "user-1" || request.GetCheckoutId() != "checkout-1" {
				return nil, status.Error(codes.NotFound, "checkout not found")
			}
			state := checkoutv1.CheckoutState_CHECKOUT_STATE_RESERVING_STOCK
			if ready {
				state = checkoutv1.CheckoutState_CHECKOUT_STATE_READY_TO_PAY
			}
			return &checkoutv1.CheckoutStatus{
				CheckoutId: "checkout-1", BuyerUserId: "user-1", OrderId: "order-1", State: state,
				PaymentSessionId: "session-1", ReturnUrl: "http://localhost:8080/orders/order-1",
			}, nil
		},
	}
	router := newTestRouter(t, routerDeps{checkout: checkoutSvc})
	newRequest := func(subject, path string) *http.Request {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		request.AddCookie(&http.Cookie{Name: auth.AccessCookieName, Value: signedAccessToken(t, subject)})
		return request
	}
	progress := httptest.NewRecorder()
	router.ServeHTTP(progress, newRequest("user-1", "/checkouts/checkout-1"))
	if progress.Code != http.StatusOK || !strings.Contains(progress.Body.String(), "reserving stock") || strings.Contains(progress.Body.String(), "session-1") {
		t.Fatalf("pending page leaked payment destination: %d %s", progress.Code, progress.Body.String())
	}
	otherBuyer := httptest.NewRecorder()
	router.ServeHTTP(otherBuyer, newRequest("user-2", "/checkouts/checkout-1"))
	if otherBuyer.Code != http.StatusNotFound || strings.Contains(otherBuyer.Body.String(), "session-1") {
		t.Fatalf("other buyer read checkout: %d", otherBuyer.Code)
	}
	ready = true
	payment := httptest.NewRecorder()
	router.ServeHTTP(payment, newRequest("user-1", "/checkouts/checkout-1"))
	if payment.Code != http.StatusSeeOther || !strings.Contains(payment.Header().Get("Location"), "/pay?") {
		t.Fatalf("ready session did not redirect to gateway: %d %s", payment.Code, payment.Header().Get("Location"))
	}
	poll := httptest.NewRecorder()
	router.ServeHTTP(poll, newRequest("user-1", "/checkouts/checkout-1/progress"))
	if poll.Code != http.StatusOK || !strings.Contains(poll.Header().Get("Content-Type"), "text/event-stream") || !strings.Contains(poll.Body.String(), "/pay?") {
		t.Fatalf("poll did not redirect only after readiness: %d %s", poll.Code, poll.Body.String())
	}
}
