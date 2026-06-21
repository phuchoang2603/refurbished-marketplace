package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"refurbished-marketplace/services/products/internal/database"
	"refurbished-marketplace/shared/messaging"
	ordersv1 "refurbished-marketplace/shared/proto/orders/v1"
	paymentv1 "refurbished-marketplace/shared/proto/payment/v1"

	"github.com/google/uuid"
)

func (s *Service) KafkaReservationHandler() messaging.KafkaHandler {
	return func(ctx context.Context, msg messaging.KafkaMessage) error {
		switch msg.Topic {
		case messaging.EventTypeOrderCreated:
			return s.HandleOrdersCreated(ctx, messaging.KafkaMessageID(msg), msg.Value)
		case messaging.EventTypePaymentSucceeded, messaging.EventTypePaymentFailed:
			return s.HandlePaymentOutcome(ctx, messaging.KafkaMessageID(msg), msg.Topic, msg.Value)
		default:
			return nil
		}
	}
}

func (s *Service) HandleOrdersCreated(ctx context.Context, messageID string, value []byte) error {
	if messageID == "" {
		return fmt.Errorf("messageID is required")
	}

	var msg ordersv1.OrderCreated
	if err := messaging.UnmarshalKafkaProtobuf(value, &msg); err != nil {
		return fmt.Errorf("decode orders.created payload: %w", err)
	}

	orderID, merchantID, totalCents, items, err := parseOrderCreatedReservation(&msg)
	if err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	q := s.queries.WithTx(tx)
	defer func() {
		_ = tx.Rollback()
	}()

	if _, err := q.InsertInventoryInboxMessage(ctx, messageID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return tx.Commit()
		}
		return err
	}

	if err := reserveOrderItems(ctx, q, orderID, items); err != nil {
		if errors.Is(err, ErrInventoryNotFound) || errors.Is(err, ErrInsufficientStock) {
			if outboxErr := createInventoryReservationFailedOutbox(ctx, q, orderID); outboxErr != nil {
				return outboxErr
			}
			return tx.Commit()
		}
		return err
	}

	if err := createInventoryReservedOutbox(ctx, q, orderID, merchantID, totalCents); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Service) HandlePaymentOutcome(ctx context.Context, messageID, topic string, value []byte) error {
	if messageID == "" {
		return fmt.Errorf("messageID is required")
	}

	var msg paymentv1.PaymentOutcome
	if err := messaging.UnmarshalKafkaProtobuf(value, &msg); err != nil {
		return fmt.Errorf("decode payment outcome payload: %w", err)
	}

	orderID, err := uuid.Parse(msg.GetOrderId())
	if err != nil {
		return fmt.Errorf("payment outcome order_id: %w", err)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	q := s.queries.WithTx(tx)
	defer func() {
		_ = tx.Rollback()
	}()

	if _, err := q.InsertInventoryInboxMessage(ctx, messageID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return tx.Commit()
		}
		return err
	}

	reservations, err := q.ListActiveInventoryReservationsByOrderID(ctx, orderID)
	if err != nil {
		return err
	}

	for _, reservation := range reservations {
		switch topic {
		case messaging.EventTypePaymentSucceeded:
			if _, err := q.CommitInventoryReservedStock(ctx, database.CommitInventoryReservedStockParams{ProductID: reservation.ProductID, Quantity: reservation.Quantity}); err != nil {
				return err
			}
			if _, err := q.MarkInventoryReservationCommitted(ctx, database.MarkInventoryReservationCommittedParams{OrderID: reservation.OrderID, ProductID: reservation.ProductID}); err != nil {
				return err
			}
		case messaging.EventTypePaymentFailed:
			if _, err := q.ReleaseInventoryReservedStock(ctx, database.ReleaseInventoryReservedStockParams{ProductID: reservation.ProductID, Quantity: reservation.Quantity}); err != nil {
				return err
			}
			if _, err := q.MarkInventoryReservationReleased(ctx, database.MarkInventoryReservationReleasedParams{OrderID: reservation.OrderID, ProductID: reservation.ProductID}); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported payment outcome topic: %s", topic)
		}
	}

	return tx.Commit()
}
