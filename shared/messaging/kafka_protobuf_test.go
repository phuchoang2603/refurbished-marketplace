package messaging

import (
	"testing"

	ordersv1 "refurbished-marketplace/shared/proto/orders/v1"

	"google.golang.org/protobuf/proto"
)

func TestUnmarshalKafkaProtobuf_Raw(t *testing.T) {
	t.Helper()
	in := &ordersv1.OrderItemCreated{OrderId: "a", OrderItemId: "b", MerchantId: "c"}
	raw, err := proto.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	var out ordersv1.OrderItemCreated
	if err := UnmarshalKafkaProtobuf(raw, &out); err != nil {
		t.Fatal(err)
	}
	if out.GetOrderId() != "a" || out.GetOrderItemId() != "b" {
		t.Fatalf("got %+v", &out)
	}
}

func TestUnmarshalKafkaProtobuf_ConfluentWire(t *testing.T) {
	t.Helper()
	in := &ordersv1.OrderItemCreated{OrderId: "x", OrderItemId: "y", MerchantId: "z"}
	body, err := proto.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	wire := append([]byte{0, 0, 0, 0, 1}, body...)
	var out ordersv1.OrderItemCreated
	if err := UnmarshalKafkaProtobuf(wire, &out); err != nil {
		t.Fatal(err)
	}
	if out.GetOrderId() != "x" {
		t.Fatalf("got %q", out.GetOrderId())
	}
}
