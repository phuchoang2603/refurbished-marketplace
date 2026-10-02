package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/phuchoang2603/refurbished-marketplace/services/payment/internal/database"
	"github.com/phuchoang2603/refurbished-marketplace/shared/messaging"
	checkoutv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/checkout/v1"
)

func (service *Service) applyCheckoutGatewayWebhook(ctx context.Context, checkoutSession database.PaymentCheckoutSession, sessionID, outcome, failureReason string) error {
	if outcome != HostedPaymentSessionStatusSucceeded && outcome != HostedPaymentSessionStatusFailed && outcome != HostedPaymentSessionStatusExpired {
		return ErrInvalidSessionFacts
	}
	if outcome == HostedPaymentSessionStatusExpired && failureReason == "" {
		failureReason = "session expired"
	}
	tx, err := service.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	queries := database.New(tx)
	intent, err := queries.GetPaymentIntentByOrderIDForUpdate(ctx, checkoutSession.OrderID)
	if err != nil {
		return err
	}
	if !intent.PaymentSessionID.Valid || intent.PaymentSessionID.String != sessionID {
		return ErrSessionMismatch
	}
	inserted, err := queries.InsertCheckoutCallback(ctx, database.InsertCheckoutCallbackParams{
		OrderID: checkoutSession.OrderID, Outcome: outcome,
	})
	if err != nil {
		return err
	}
	if inserted == 0 {
		return tx.Commit()
	}
	if intent.Status == HostedPaymentSessionStatusPending {
		_, err = queries.UpdateHostedPaymentSessionOutcome(ctx, database.UpdateHostedPaymentSessionOutcomeParams{
			OrderID: checkoutSession.OrderID, PaymentSessionID: sql.NullString{String: sessionID, Valid: true},
			Status: outcome, FailureReason: sql.NullString{String: failureReason, Valid: failureReason != ""},
		})
		if err != nil {
			return err
		}
		transaction, err := queries.GetPaymentTransactionByOrderID(ctx, checkoutSession.OrderID)
		if err != nil {
			return err
		}
		transactionStatus := PaymentTxStatusSucceeded
		if outcome != HostedPaymentSessionStatusSucceeded {
			transactionStatus = PaymentTxStatusFailed
		}
		if _, err := queries.UpdatePaymentTransactionGatewayResult(ctx, database.UpdatePaymentTransactionGatewayResultParams{
			ID: transaction.ID, Status: transactionStatus,
			FailureReason: sql.NullString{String: failureReason, Valid: failureReason != ""},
		}); err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
	}
	if !checkoutSession.CreateOperationID.Valid || !checkoutSession.CreateVersion.Valid {
		return errors.New("checkout session missing original operation")
	}
	topic := messaging.EventTypeCheckoutPaymentSucceeded
	var payload any = &checkoutv1.PaymentSucceeded{PaymentSessionId: sessionID}
	if outcome != HostedPaymentSessionStatusSucceeded {
		topic = messaging.EventTypeCheckoutPaymentFailed
		payload = &checkoutv1.PaymentFailed{PaymentSessionId: sessionID, Reason: failureReason}
	}
	return emitCheckoutPaymentResult(ctx, tx, queries, checkoutSession.CheckoutID, checkoutSession.OrderID,
		&checkoutv1.MessageContext{
			MessageId: uuid.NewString(), OperationId: checkoutSession.CreateOperationID.UUID.String(),
			WorkflowVersion: uint64(checkoutSession.CreateVersion.Int64),
		}, topic, payload)
}
