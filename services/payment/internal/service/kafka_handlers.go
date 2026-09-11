package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sharedlog "github.com/phuchoang2603/refurbished-marketplace/shared/observe/log"

	"github.com/phuchoang2603/refurbished-marketplace/shared/err/dberr"
	"github.com/phuchoang2603/refurbished-marketplace/shared/messaging"

	productsv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/products/v1"

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

		orderID, _, err := parseOrderUUIDs(&payload)
		if err != nil {
			return err
		}

		if _, err := loadPaymentIntentByOrderID(ctx, s.queries, orderID); err != nil {
			return err
		}
		if _, err := s.queries.GetPaymentTransactionByOrderID(ctx, orderID); err != nil {
			return dberr.MapErrNoRows(err, ErrTransactionNotFound)
		}

		if err := s.ensureTerminalOutcomeForOrder(ctx, orderID); err != nil {
			return err
		}

		if _, err := s.queries.InsertPaymentInboxMessage(ctx, messageID); err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		sharedlog.InfoContext(ctx, "inventory.reserved acknowledged", sharedlog.KeyOrderID, orderID.String())
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
