package service

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/phuchoang2603/refurbished-marketplace/services/checkout/internal/database"
	"github.com/phuchoang2603/refurbished-marketplace/shared/messaging"
	sharedlog "github.com/phuchoang2603/refurbished-marketplace/shared/observe/log"
	sharedtrace "github.com/phuchoang2603/refurbished-marketplace/shared/observe/trace"
	checkoutv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/checkout/v1"
	"google.golang.org/protobuf/proto"
)

const deadlineRetryInterval = 30 * time.Second

func (service *Service) RunDeadlines(ctx context.Context) error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		processed, err := service.ProcessDue(ctx)
		if err != nil && ctx.Err() == nil {
			sharedlog.Error("checkout deadline processing failed", "err", err)
		}
		if processed {
			continue
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

func (service *Service) ProcessDue(ctx context.Context) (bool, error) {
	tx, err := service.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	queries := database.New(tx)
	rows, err := queries.ClaimDueCheckouts(ctx, 1)
	if err != nil || len(rows) == 0 {
		return false, err
	}
	row := rows[0]
	state := checkoutv1.CheckoutState(checkoutv1.CheckoutState_value[row.State])
	nextDeadline := time.Now().Add(deadlineRetryInterval)

	switch state {
	case checkoutv1.CheckoutState_CHECKOUT_STATE_RESERVING_STOCK:
		err = deadlineCommand(ctx, queries, row, checkoutv1.CheckoutState_CHECKOUT_STATE_RELEASING_STOCK,
			messaging.EventTypeCheckoutCancelStock, "cancel-stock", &checkoutv1.CancelStock{Reason: "reservation deadline exceeded"}, nextDeadline)
	case checkoutv1.CheckoutState_CHECKOUT_STATE_CREATING_SESSION,
		checkoutv1.CheckoutState_CHECKOUT_STATE_READY_TO_PAY:
		err = deadlineCommand(ctx, queries, row, checkoutv1.CheckoutState_CHECKOUT_STATE_RECONCILING,
			messaging.EventTypeCheckoutCancelSession, "cancel-session", &checkoutv1.CancelPaymentSession{
				PaymentSessionId: row.PaymentSessionID.String, Reason: "payment outcome verification required",
			}, nextDeadline)
	case checkoutv1.CheckoutState_CHECKOUT_STATE_RECONCILING:
		_, err = queries.AdvanceCheckout(ctx, database.AdvanceCheckoutParams{
			ID: row.ID, Version: row.Version, State: checkoutv1.CheckoutState_CHECKOUT_STATE_NEEDS_REVIEW.String(),
			DeadlineAt: sql.NullTime{},
		})
		if err == nil {
			err = queries.InsertException(ctx, database.InsertExceptionParams{
				ID: uuid.New(), CheckoutID: row.ID, Kind: "payment_uncertain", Details: "payment cancellation or verification did not complete before deadline",
			})
		}
	case checkoutv1.CheckoutState_CHECKOUT_STATE_CREATING_ORDER,
		checkoutv1.CheckoutState_CHECKOUT_STATE_COMMITTING_STOCK,
		checkoutv1.CheckoutState_CHECKOUT_STATE_RELEASING_STOCK,
		checkoutv1.CheckoutState_CHECKOUT_STATE_FINALIZING_PAID,
		checkoutv1.CheckoutState_CHECKOUT_STATE_FINALIZING_FAILED:
		err = retryLastCommand(ctx, queries, row)
		if err == nil {
			_, err = queries.DelayCheckoutDeadline(ctx, database.DelayCheckoutDeadlineParams{
				ID: row.ID, DeadlineAt: sql.NullTime{Time: nextDeadline, Valid: true}, Version: row.Version,
			})
		}
	default:
		_, err = queries.DelayCheckoutDeadline(ctx, database.DelayCheckoutDeadlineParams{ID: row.ID, Version: row.Version})
	}
	if err != nil {
		return false, err
	}
	return true, tx.Commit()
}

func deadlineCommand(ctx context.Context, queries *database.Queries, row database.Checkout, nextState checkoutv1.CheckoutState, topic, operation string, payload any, deadline time.Time) error {
	updated, err := queries.AdvanceCheckout(ctx, database.AdvanceCheckoutParams{
		ID: row.ID, Version: row.Version, State: nextState.String(),
		DeadlineAt: sql.NullTime{Time: deadline, Valid: true},
	})
	if err != nil {
		return err
	}
	if updated != 1 {
		return errors.New("deadline transition lost checkout lock")
	}
	operationID := uuid.NewSHA1(row.ID, []byte(operation))
	command := &checkoutv1.CheckoutMessage{Context: &checkoutv1.MessageContext{
		MessageId: operationID.String(), CheckoutId: row.ID.String(), OrderId: row.OrderID.UUID.String(),
		OperationId: operationID.String(), WorkflowVersion: uint64(row.Version + 1),
	}}
	switch value := payload.(type) {
	case *checkoutv1.CancelStock:
		command.Payload = &checkoutv1.CheckoutMessage_CancelStock{CancelStock: value}
	case *checkoutv1.CancelPaymentSession:
		command.Payload = &checkoutv1.CheckoutMessage_CancelPaymentSession{CancelPaymentSession: value}
	default:
		return errors.New("unexpected deadline command")
	}
	encoded, err := proto.Marshal(command)
	if err != nil {
		return err
	}
	return queries.InsertOutbox(ctx, database.InsertOutboxParams{
		ID: operationID, AggregateID: row.ID, EventType: topic,
		Payload: encoded, Tracingspancontext: sql.NullString{String: sharedtrace.SerializeContext(ctx), Valid: true},
	})
}

func retryLastCommand(ctx context.Context, queries *database.Queries, row database.Checkout) error {
	last, err := queries.GetLatestOutbox(ctx, row.ID)
	if err != nil {
		return err
	}
	command := new(checkoutv1.CheckoutMessage)
	if err := proto.Unmarshal(last.Payload, command); err != nil {
		return err
	}
	if command.GetContext().GetWorkflowVersion() != uint64(row.Version) {
		return errors.New("latest command does not match due checkout version")
	}
	command.GetContext().MessageId = uuid.NewString()
	encoded, err := proto.Marshal(command)
	if err != nil {
		return err
	}
	return queries.InsertOutbox(ctx, database.InsertOutboxParams{
		ID: uuid.New(), AggregateID: row.ID, EventType: last.EventType, Payload: encoded,
		Tracingspancontext: sql.NullString{String: sharedtrace.SerializeContext(ctx), Valid: true},
	})
}
