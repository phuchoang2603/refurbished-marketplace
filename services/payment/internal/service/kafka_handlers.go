package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sharedlog "refurbished-marketplace/shared/observe/log"

	"refurbished-marketplace/services/payment/internal/database"
	"refurbished-marketplace/shared/err/dberr"
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
		if err != nil && !errors.Is(err, sql.ErrNoRows) && !isPostgresUniqueViolation(err) {
			return err
		}
		if err == nil {
			sharedlog.InfoContext(
				ctx, "payment transaction initialized from inventory.reserved",
				sharedlog.KeyOrderID, orderID.String(),
				sharedlog.KeyMerchantID, merchantID.String(),
				"payment_transaction_id", created.ID.String(),
				"amount_cents", payload.GetTotalCents(),
				"currency", intent.Currency,
			)
		}

		// Lock the intent row so we wait for an in-flight expiry sweep, then
		// apply any terminal outcome. Also covers retries where the tx exists.
		if err := s.ensureTerminalOutcomeForOrder(ctx, orderID); err != nil {
			return err
		}

		// Ack only after create + catch-up so Kafka retries remain durable.
		if _, err := s.queries.InsertPaymentInboxMessage(ctx, messageID); err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		return nil
	}
}

func (s *Service) ensureTerminalOutcomeForOrder(ctx context.Context, orderID uuid.UUID) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()
	q := s.queries.WithTx(tx)

	intent, err := q.GetPaymentIntentByOrderIDForUpdate(ctx, orderID)
	if err != nil {
		return dberr.MapErrNoRows(err, ErrIntentNotFound)
	}
	if !hostedPaymentSessionIsTerminal(intent.Status) {
		return tx.Commit()
	}

	txRow, err := q.GetPaymentTransactionByOrderID(ctx, orderID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return tx.Commit()
		}
		return err
	}
	if paymentTransactionIsTerminal(txRow.Status) {
		return tx.Commit()
	}

	if err := s.applyTerminalOutcomeWithQueries(ctx, q, txRow.ID, intent.Status, intent.FailureReason); err != nil {
		return err
	}
	return tx.Commit()
}
