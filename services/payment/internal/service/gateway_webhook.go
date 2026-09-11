package service

import (
	"context"
	"database/sql"
	"errors"

	sharedlog "github.com/phuchoang2603/refurbished-marketplace/shared/observe/log"

	"github.com/phuchoang2603/refurbished-marketplace/services/payment/internal/database"
	"github.com/phuchoang2603/refurbished-marketplace/shared/err/dberr"
	"github.com/phuchoang2603/refurbished-marketplace/shared/messaging"
	sharedtrace "github.com/phuchoang2603/refurbished-marketplace/shared/observe/trace"

	paymentv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/payment/v1"

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
		return s.publishTerminalOutcomeIfNeeded(ctx, orderID, intent.Status, intent.FailureReason)
	}

	updatedIntent, err := s.queries.UpdateHostedPaymentSessionOutcome(ctx, database.UpdateHostedPaymentSessionOutcomeParams{
		OrderID:          orderID,
		PaymentSessionID: dberr.OptionalNullString(paymentSessionID),
		Status:           status,
		FailureReason:    dberr.OptionalNullString(failureReason),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			current, loadErr := loadPaymentIntentByOrderID(ctx, s.queries, orderID)
			if loadErr != nil {
				return loadErr
			}
			if hostedPaymentSessionIsTerminal(current.Status) {
				return s.publishTerminalOutcomeIfNeeded(ctx, orderID, current.Status, current.FailureReason)
			}
			return ErrSessionMismatch
		}
		return err
	}

	return s.publishTerminalOutcomeIfNeeded(ctx, orderID, updatedIntent.Status, updatedIntent.FailureReason)
}

func (s *Service) publishTerminalOutcomeIfNeeded(ctx context.Context, orderID uuid.UUID, hostedStatus string, failureReason sql.NullString) error {
	txRow, err := s.queries.GetPaymentTransactionByOrderID(ctx, orderID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if hostedStatus == HostedPaymentSessionStatusFailed {
				return s.emitPaymentOutcome(ctx, s.queries, orderID, messaging.EventTypePaymentFailed)
			}
			return nil
		}
		return err
	}
	return s.applyTerminalOutcome(ctx, txRow.ID, hostedStatus, failureReason)
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
				return nil
			}
		}
		return err
	}

	eventType := messaging.EventTypePaymentFailed
	if newStatus == PaymentTxStatusSucceeded {
		eventType = messaging.EventTypePaymentSucceeded
	}
	if err := s.emitPaymentOutcome(ctx, q, updated.OrderID, eventType); err != nil {
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

func (s *Service) emitPaymentOutcome(ctx context.Context, q *database.Queries, orderID uuid.UUID, eventType string) error {
	existing, err := q.ListPaymentOutboxByAggregateID(ctx, orderID)
	if err != nil {
		return err
	}
	for _, row := range existing {
		if row.EventType == messaging.EventTypePaymentFailed || row.EventType == messaging.EventTypePaymentSucceeded {
			return nil
		}
	}

	payload, err := proto.Marshal(&paymentv1.PaymentOutcome{OrderId: orderID.String()})
	if err != nil {
		return err
	}
	if _, err := q.CreatePaymentOutbox(ctx, database.CreatePaymentOutboxParams{
		ID:                 uuid.New(),
		AggregateID:        orderID,
		EventType:          eventType,
		Payload:            payload,
		Tracingspancontext: sharedtrace.SerializeContext(ctx),
	}); err != nil {
		return err
	}
	return nil
}
