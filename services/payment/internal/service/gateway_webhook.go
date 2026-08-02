package service

import (
	"context"
	"database/sql"
	"errors"

	sharedlog "refurbished-marketplace/shared/observe/log"

	"refurbished-marketplace/services/payment/internal/database"
	"refurbished-marketplace/shared/err/dberr"
	"refurbished-marketplace/shared/messaging"
	sharedtrace "refurbished-marketplace/shared/observe/trace"

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
		sharedlog.InfoContext(
			ctx, "gateway webhook ignored; session already terminal",
			sharedlog.KeyOrderID, orderID.String(),
			"payment_session_id", paymentSessionID,
			sharedlog.KeyStatus, intent.Status,
		)
		return nil
	}

	updatedIntent, err := s.queries.UpdateHostedPaymentSessionOutcome(ctx, database.UpdateHostedPaymentSessionOutcomeParams{
		OrderID:          orderID,
		PaymentSessionID: dberr.OptionalNullString(paymentSessionID),
		Status:           status,
		FailureReason:    dberr.OptionalNullString(failureReason),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Lost a race to another terminal writer (e.g. session expiry sweep).
			current, loadErr := loadPaymentIntentByOrderID(ctx, s.queries, orderID)
			if loadErr != nil {
				return loadErr
			}
			if hostedPaymentSessionIsTerminal(current.Status) {
				sharedlog.InfoContext(
					ctx, "gateway webhook ignored; session became terminal",
					sharedlog.KeyOrderID, orderID.String(),
					"payment_session_id", paymentSessionID,
					sharedlog.KeyStatus, current.Status,
				)
				return nil
			}
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
	defer func() {
		_ = tx.Rollback()
	}()

	if err := s.applyTerminalOutcomeWithQueries(ctx, s.queries.WithTx(tx), transactionID, hostedStatus, failureReason); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Service) applyTerminalOutcomeWithQueries(
	ctx context.Context,
	q *database.Queries,
	transactionID uuid.UUID,
	hostedStatus string,
	failureReason sql.NullString,
) error {
	newStatus := terminalPaymentTxStatus(hostedStatus)

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
				return nil
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
		ID:                 uuid.New(),
		AggregateID:        updated.OrderID,
		EventType:          eventType,
		Payload:            payload,
		Tracingspancontext: sharedtrace.SerializeContext(ctx),
	}); err != nil {
		return err
	}

	sharedlog.InfoContext(
		ctx, "payment terminal outcome applied",
		sharedlog.KeyOrderID, updated.OrderID.String(),
		"payment_transaction_id", transactionID.String(),
		sharedlog.KeyStatus, newStatus,
		sharedlog.KeyEventType, eventType,
	)
	return nil
}

func terminalPaymentTxStatus(hostedStatus string) string {
	if hostedPaymentSessionMapsToSuccess(hostedStatus) {
		return PaymentTxStatusSucceeded
	}
	return PaymentTxStatusFailed
}
