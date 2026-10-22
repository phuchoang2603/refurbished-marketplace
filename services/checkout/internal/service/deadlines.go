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
	"go.opentelemetry.io/otel/attribute"
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
	defer func() { _ = tx.Rollback() }()
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
	case checkoutv1.CheckoutState_CHECKOUT_STATE_CREATING_PAYMENT:
		err = queries.FinishGatewayOperation(ctx, database.FinishGatewayOperationParams{CheckoutID: row.ID, Status: "failed"})
		if err == nil {
			err = deadlineCommand(ctx, queries, row, checkoutv1.CheckoutState_CHECKOUT_STATE_RELEASING_STOCK, messaging.EventTypeCheckoutCancelStock, "cancel-stock", &checkoutv1.CancelStock{Reason: "payment preparation deadline exceeded"}, nextDeadline)
		}
	case checkoutv1.CheckoutState_CHECKOUT_STATE_READY_TO_PAY:
		return service.reviewPaymentDeadline(ctx, tx, queries, row)
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
	case *checkoutv1.CommitStock:
		command.Payload = &checkoutv1.CheckoutMessage_CommitStock{CommitStock: value}
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

func (service *Service) reviewPaymentDeadline(ctx context.Context, tx *sql.Tx, queries *database.Queries, row database.Checkout) (bool, error) {
	if err := queries.MarkGatewayStatusChecked(ctx, row.ID); err != nil {
		return false, err
	}
	_, err := queries.AdvanceCheckout(ctx, database.AdvanceCheckoutParams{ID: row.ID, Version: row.Version, State: checkoutv1.CheckoutState_CHECKOUT_STATE_NEEDS_REVIEW.String()})
	if err != nil {
		return false, err
	}
	eventID := "gateway-status:" + row.ID.String()
	if err := queries.InsertException(ctx, database.InsertExceptionParams{ID: uuid.New(), CheckoutID: row.ID, MessageID: sql.NullString{String: eventID, Valid: true}, Kind: "payment_uncertain", Details: "deadline status read has not confirmed a terminal outcome"}); err != nil {
		return false, err
	}
	if err := tx.Commit(); err != nil {
		return false, err
	}
	ctx, span := sharedtrace.Tracer("checkout").Start(ctx, "gateway.payment_status")
	defer span.End()
	span.SetAttributes(attribute.String("order_id", row.OrderID.UUID.String()), attribute.String("gateway_payment_id", row.GatewayPaymentID.String))
	payment, err := service.gateway.get(ctx, row.GatewayPaymentID.String)
	if err != nil || payment.OrderReference != row.OrderID.UUID.String() || (payment.Status != "SUCCEEDED" && payment.Status != "EXPIRED") {
		return true, nil
	}
	settledTx, err := service.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer func() { _ = settledTx.Rollback() }()
	settledQueries := database.New(settledTx)
	row, err = settledQueries.LockCheckout(ctx, row.ID)
	if err != nil {
		return false, err
	}
	eventType := "payment.expired"
	if payment.Status == "SUCCEEDED" {
		eventType = "payment.succeeded"
	}
	if err := service.applyGatewayOutcome(ctx, settledQueries, row, eventType, eventID, time.Now()); err != nil {
		return false, err
	}
	return true, settledTx.Commit()
}
