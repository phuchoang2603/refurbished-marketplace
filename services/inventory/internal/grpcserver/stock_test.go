package grpcserver

import (
	"testing"

	inventoryv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/inventory/v1"
)

func TestDirectCheckoutReservationNotExposed(t *testing.T) {
	for _, method := range inventoryv1.InventoryService_ServiceDesc.Methods {
		if method.MethodName == "ReserveStock" {
			t.Fatal("ReserveStock must not be exposed over gRPC")
		}
	}
}
