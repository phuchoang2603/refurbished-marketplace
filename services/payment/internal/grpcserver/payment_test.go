package grpcserver

import (
	"testing"

	paymentv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/payment/v1"
)

func TestDirectHostedSessionCreateNotExposed(t *testing.T) {
	for _, method := range paymentv1.PaymentService_ServiceDesc.Methods {
		if method.MethodName == "CreateHostedPaymentSession" {
			t.Fatal("CreateHostedPaymentSession must not be exposed over gRPC")
		}
	}
}
