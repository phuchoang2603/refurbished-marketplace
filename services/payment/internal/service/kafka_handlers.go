package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sharedlog "refurbished-marketplace/shared/observe/log"

	"refurbished-marketplace/services/payment/internal/database"
	"refurbished-marketplace/shared/messaging"

	productsv1 "refurbished-marketplace/shared/proto/products/v1"

	"github.com/google/uuid"
)

func (s *Service) KafkaInventoryReservedHandler() messaging.KafkaHandler {
	return func(ctx context.Context, msg messaging.KafkaMessage) error {
		messageID := messaging.KafkaMessageID(msg)
		if messageID == "" {
			return errors.New("messageID is required")
		}

		var payload productsv1.InventoryReserved
		if err := messaging.UnmarshalKafkaProtobuf(msg.Value, &payload); err != nil {
			return fmt.Errorf("decode inventory.reserved payload: %w", err)
		}
		if payload.GetOrderId() == "" {
			return errors.New("invalid inventory.reserved payload: missing order_id")
		}

		orderID, merchantID, err := parseOrderUUIDs(&payload)
		if err != nil {
			return err
		}

		intent, err := loadPaymentIntentByOrderID(ctx, s.queries, orderID)
		if err != nil {
			return err
		}

		created, err := s.queries.CreatePaymentTransaction(ctx, database.CreatePaymentTransactionParams{
			ID:             uuid.New(),
			OrderID:        orderID,
			MerchantID:     merchantID,
			AmountCents:    payload.GetTotalCents(),
			Currency:       intent.Currency,
			Status:         PaymentTxStatusInitialized,
			IdempotencyKey: "order:" + orderID.String(),
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) || isPostgresUniqueViolation(err) {
				if err := s.ensureTerminalOutcomeForOrder(ctx, orderID); err != nil {
					return err
				}
				return s.ackPaymentInbox(ctx, messageID)
			}
			return err
		}

		// Re-read after create so a concurrent expiry sweep is observed.
		if err := s.ensureTerminalOutcomeForOrder(ctx, orderID); err != nil {
			return err
		}

		sharedlog.InfoContext(
			ctx, "payment transaction initialized from inventory.reserved",
			sharedlog.KeyOrderID, orderID.String(),
			sharedlog.KeyMerchantID, merchantID.String(),
			"payment_transaction_id", created.ID.String(),
			"amount_cents", payload.GetTotalCents(),
			"currency", intent.Currency,
		)
		return s.ackPaymentInbox(ctx, messageID)
	}
}

func (s *Service) ensureTerminalOutcomeForOrder(ctx context.Context, orderID uuid.UUID) error {
	intent, err := loadPaymentIntentByOrderID(ctx, s.queries, orderID)
	if err != nil {
		return err
	}
	if !hostedPaymentSessionIsTerminal(intent.Status) {
		return nil
	}

	txRow, err := s.queries.GetPaymentTransactionByOrderID(ctx, orderID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}
	if paymentTransactionIsTerminal(txRow.Status) {
		return nil
	}

	return s.applyTerminalOutcome(ctx, txRow.ID, intent.Status, intent.FailureReason)
}

func (s *Service) ackPaymentInbox(ctx context.Context, messageID string) error {
	if _, err := s.queries.InsertPaymentInboxMessage(ctx, messageID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}
	return nil
}
