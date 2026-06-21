package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"refurbished-marketplace/shared/messaging"
	paymentv1 "refurbished-marketplace/shared/proto/payment/v1"
	productsv1 "refurbished-marketplace/shared/proto/products/v1"

	"github.com/google/uuid"
)

func (s *Service) KafkaOrderResultHandler() messaging.KafkaHandler {
	return func(ctx context.Context, msg messaging.KafkaMessage) error {
		messageID := messaging.KafkaMessageID(msg)
		if messageID == "" {
			return errors.New("messageID is required")
		}

		var (
			orderID uuid.UUID
			status  string
			err     error
		)

		switch msg.Topic {
		case messaging.EventTypePaymentSucceeded:
			orderID, err = parsePaymentOutcomeOrderID(msg.Value)
			status = OrderStatusPaid
		case messaging.EventTypePaymentFailed:
			orderID, err = parsePaymentOutcomeOrderID(msg.Value)
			status = OrderStatusFailed
		case messaging.EventTypeInventoryReservationFailed:
			orderID, err = parseInventoryReservationFailedOrderID(msg.Value)
			status = OrderStatusFailed
		default:
			return nil
		}

		if err != nil {
			return err
		}

		return s.applyKafkaOrderResult(ctx, messageID, orderID, status)
	}
}

func (s *Service) applyKafkaOrderResult(ctx context.Context, messageID string, orderID uuid.UUID, status string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	q := s.queries.WithTx(tx)
	defer func() {
		_ = tx.Rollback()
	}()

	if _, err := q.InsertOrdersInboxMessage(ctx, messageID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}

	if err := s.updateOrderStatusWithQueries(ctx, q, orderID, status); err != nil {
		return err
	}

	return tx.Commit()
}

func parsePaymentOutcomeOrderID(value []byte) (uuid.UUID, error) {
	var payload paymentv1.PaymentOutcome
	if err := messaging.UnmarshalKafkaProtobuf(value, &payload); err != nil {
		return uuid.Nil, fmt.Errorf("payment result decode: %w", err)
	}
	orderID, err := uuid.Parse(payload.GetOrderId())
	if err != nil {
		return uuid.Nil, fmt.Errorf("payment result order_id: %w", err)
	}
	return orderID, nil
}

func parseInventoryReservationFailedOrderID(value []byte) (uuid.UUID, error) {
	var payload productsv1.InventoryReservationFailed
	if err := messaging.UnmarshalKafkaProtobuf(value, &payload); err != nil {
		return uuid.Nil, fmt.Errorf("inventory reservation failed decode: %w", err)
	}
	orderID, err := uuid.Parse(payload.GetOrderId())
	if err != nil {
		return uuid.Nil, fmt.Errorf("inventory reservation failed order_id: %w", err)
	}
	return orderID, nil
}
