package grpcserver

import (
	"testing"

	ordersv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/orders/v1"
)

func TestDirectCheckoutOrderMutationsNotExposed(t *testing.T) {
	for _, method := range ordersv1.OrdersService_ServiceDesc.Methods {
		if method.MethodName == "CreateOrder" || method.MethodName == "UpdateOrderStatus" {
			t.Fatalf("%s must not be exposed over gRPC", method.MethodName)
		}
	}
}
