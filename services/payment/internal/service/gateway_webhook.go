package service

import (
	"context"
	"database/sql"
	"errors"

	"refurbished-marketplace/services/payment/internal/database"
	shareddb "refurbished-marketplace/shared/db"
	"refurbished-marketplace/shared/messaging"

	paymentv1 "refurbished-marketplace/shared/proto/payment/v1"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

func (s *Service) ApplyGatewayWebhook(ctx context.Context, orderID uuid.UUID, paymentSessionID, status, failureReason string) error {
	intent, err := loadPaymentIntentByOrderID(ctx, s.queries, orderID)
	if err != nil {
		return err
	}
	if !intent.PaymentSessionID.Valid || intent.PaymentSessionID.String != paymentSessionID {
		return ErrSessionMismatch
	}
	if hostedPaymentSessionIsTerminal(intent.Status) {
		return nil
	}

	updatedIntent, err := s.queries.UpdateHostedPaymentSessionOutcome(ctx, database.UpdateHostedPaymentSessionOutcomeParams{
		OrderID:          orderID,
		PaymentSessionID: shareddb.OptionalNullString(paymentSessionID),
		Status:           status,
		FailureReason:    shareddb.OptionalNullString(failureReason),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrSessionMismatch
		}
		return err
	}

	txRow, err := s.queries.GetPaymentTransactionByOrderID(ctx, orderID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}

	return s.applyTerminalOutcome(ctx, txRow.ID, updatedIntent.Status, updatedIntent.FailureReason)
}

func (s *Service) applyTerminalOutcome(ctx context.Context, transactionID uuid.UUID, hostedStatus string, failureReason sql.NullString) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	q := s.queries.WithTx(tx)
	defer func() {
		_ = tx.Rollback()
	}()

	newStatus := PaymentTxStatusFailed
	if hostedPaymentSessionMapsToSuccess(hostedStatus) {
		newStatus = PaymentTxStatusSucceeded
	}

	updated, err := q.UpdatePaymentTransactionGatewayResult(ctx, database.UpdatePaymentTransactionGatewayResultParams{
		ID:                   transactionID,
		Status:               newStatus,
		GatewayTransactionID: sql.NullString{},
		FailureReason:        failureReason,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			row, loadErr := loadPaymentTransaction(ctx, q, transactionID)
			if loadErr != nil {
				return loadErr
			}
			if paymentTransactionIsTerminal(row.Status) {
				return tx.Commit()
			}
		}
		return err
	}

	eventType := messaging.EventTypePaymentFailed
	if newStatus == PaymentTxStatusSucceeded {
		eventType = messaging.EventTypePaymentSucceeded
	}
	payload, err := proto.Marshal(&paymentv1.PaymentOutcome{OrderId: updated.OrderID.String()})
	if err != nil {
		return err
	}

	if _, err := q.CreatePaymentOutbox(ctx, database.CreatePaymentOutboxParams{
		ID:          uuid.New(),
		AggregateID: updated.OrderID,
		EventType:   eventType,
		Payload:     payload,
	}); err != nil {
		return err
	}

	return tx.Commit()
}
