package service

import (
	"context"
	"fmt"
	"refurbished-marketplace/shared/messaging"
	paymentv1 "refurbished-marketplace/shared/proto/payment/v1"

	"github.com/google/uuid"
)

// KafkaPaymentResultHandler consumes payment.item.succeeded / payment.item.failed and updates order status.
func (s *Service) KafkaPaymentResultHandler() messaging.KafkaHandler {
	return func(ctx context.Context, msg messaging.KafkaMessage) error {
		var status string
		switch msg.Topic {
		case messaging.EventTypePaymentItemSucceeded:
			status = OrderStatusPaid
		case messaging.EventTypePaymentItemFailed:
			status = OrderStatusFailed
		default:
			return nil
		}

		var payload paymentv1.PaymentItemOutbox
		if err := messaging.UnmarshalKafkaProtobuf(msg.Value, &payload); err != nil {
			return fmt.Errorf("payment result decode: %w", err)
		}
		orderID, err := uuid.Parse(payload.GetOrderId())
		if err != nil {
			return fmt.Errorf("payment result order_id: %w", err)
		}

		if _, err := s.UpdateOrderStatus(ctx, orderID, status); err != nil {
			return err
		}
		return nil
	}
}
