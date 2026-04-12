package service

import (
	"context"
	"encoding/json"
	"fmt"
	"refurbished-marketplace/shared/messaging"

	"github.com/google/uuid"
)

type paymentItemResult struct {
	OrderID     string `json:"order_id"`
	OrderItemID string `json:"order_item_id"`
	Status      string `json:"status"`
}

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

		var payload paymentItemResult
		if err := json.Unmarshal(msg.Value, &payload); err != nil {
			return fmt.Errorf("payment result json: %w", err)
		}
		orderID, err := uuid.Parse(payload.OrderID)
		if err != nil {
			return fmt.Errorf("payment result order_id: %w", err)
		}

		if _, err := s.UpdateOrderStatus(ctx, orderID, status); err != nil {
			return err
		}
		return nil
	}
}
