package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/phuchoang2603/refurbished-marketplace/services/checkout/internal/database"
	"github.com/phuchoang2603/refurbished-marketplace/shared/messaging"
	sharedtrace "github.com/phuchoang2603/refurbished-marketplace/shared/observe/trace"
	checkoutv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/checkout/v1"
	"google.golang.org/protobuf/proto"
)

const (
	preparationDeadline = 2 * time.Minute
	preparationVersion  = 1

	outcomePending   = "PENDING"
	outcomeSucceeded = "SUCCEEDED"
	outcomeRejected  = "REJECTED"

	closurePrepared = "PREPARED"
	closureRejected = "REJECTED"
	closureExpired  = "EXPIRED"
)

func preparationOperationID(checkoutID uuid.UUID) uuid.UUID {
	return uuid.NewSHA1(checkoutID, []byte("prepare"))
}

type transition struct {
	state          checkoutv1.CheckoutState
	commandTopic   string
	commandPayload any
	operation      string
	deadline       time.Duration
	reason         string
}

func (service *Service) HandleResult(ctx context.Context, kafkaMessage messaging.KafkaMessage) error {
	message := new(checkoutv1.CheckoutMessage)
	if err := proto.Unmarshal(kafkaMessage.Value, message); err != nil {
		return err
	}
	messageContext := message.GetContext()
	if messageContext == nil || messageContext.GetMessageId() == "" || messageContext.GetOperationId() == "" {
		return errors.New("checkout result missing message and operation identifiers")
	}
	checkoutID, err := uuid.Parse(messageContext.GetCheckoutId())
	if err != nil || checkoutID == uuid.Nil {
		return errors.New("checkout result has invalid checkout identifier")
	}

	tx, err := service.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	queries := database.New(tx)
	row, err := queries.LockCheckout(ctx, checkoutID)
	if err != nil {
		return err
	}
	if row.OrderID.UUID.String() != messageContext.GetOrderId() || !row.OrderID.Valid {
		return fmt.Errorf("checkout result has wrong order for checkout %s", checkoutID)
	}
	inserted, err := queries.InsertInbox(ctx, database.InsertInboxParams{
		MessageID: messageContext.GetMessageId(), CheckoutID: checkoutID,
	})
	if err != nil {
		return err
	}
	if inserted == 0 {
		return tx.Commit()
	}

	switch kafkaMessage.Topic {
	case messaging.EventTypeCheckoutOrderCreated, messaging.EventTypeCheckoutOrderRejected,
		messaging.EventTypeCheckoutStockReserved, messaging.EventTypeCheckoutStockRejected:
		if err := handlePreparationResult(ctx, queries, row, kafkaMessage.Topic, message); err != nil {
			return err
		}
		return tx.Commit()
	}

	step, err := nextTransition(row, kafkaMessage.Topic, message)
	if err != nil {
		return err
	}
	if step == nil {
		if messageContext.GetWorkflowVersion() > uint64(row.Version) {
			return fmt.Errorf("checkout result arrived before its command: checkout %s version %d, result version %d", checkoutID, row.Version, messageContext.GetWorkflowVersion())
		}
		return tx.Commit()
	}
	var deadline sql.NullTime
	if step.deadline > 0 {
		deadline = sql.NullTime{Time: time.Now().Add(step.deadline), Valid: true}
	}
	updated, err := queries.AdvanceCheckout(ctx, database.AdvanceCheckoutParams{
		ID: checkoutID, Version: row.Version, State: step.state.String(),
		FailureReason: sql.NullString{String: step.reason, Valid: step.reason != ""},
		DeadlineAt:    deadline,
	})
	if err != nil {
		return err
	}
	if updated != 1 {
		return errors.New("checkout transition lost optimistic concurrency race")
	}
	if step.state == checkoutv1.CheckoutState_CHECKOUT_STATE_CREATING_PAYMENT {
		if len(row.GatewaySnapshot) == 0 {
			return errors.New("missing frozen gateway snapshot")
		}
		if err := queries.InsertGatewayOperation(ctx, database.InsertGatewayOperationParams{CheckoutID: row.ID, Request: row.GatewaySnapshot, Tracingspancontext: sql.NullString{String: sharedtrace.SerializeContext(ctx), Valid: true}}); err != nil {
			return err
		}
	}
	if step.commandTopic != "" {
		if err := enqueueCheckoutCommand(ctx, queries, row, step.commandTopic, step.operation, messageContext.GetMessageId(), step.commandPayload); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func nextTransition(row database.Checkout, topic string, result *checkoutv1.CheckoutMessage) (*transition, error) {
	current := checkoutv1.CheckoutState(checkoutv1.CheckoutState_value[row.State])
	context := result.GetContext()
	expectedStep := ""
	expectedVersion := row.Version
	switch current {
	case checkoutv1.CheckoutState_CHECKOUT_STATE_COMMITTING_STOCK:
		expectedStep = "commit-stock"
	case checkoutv1.CheckoutState_CHECKOUT_STATE_RELEASING_STOCK:
		expectedStep = "cancel-stock"
	case checkoutv1.CheckoutState_CHECKOUT_STATE_FINALIZING_PAID:
		expectedStep = "finalize-paid"
	case checkoutv1.CheckoutState_CHECKOUT_STATE_FINALIZING_FAILED:
		expectedStep = "finalize-failed"

	}
	validExpected := expectedStep != "" && context.GetOperationId() == uuid.NewSHA1(row.ID, []byte(expectedStep)).String() &&
		context.GetWorkflowVersion() == uint64(expectedVersion)
	if !validExpected {
		return nil, nil
	}

	switch current {
	case checkoutv1.CheckoutState_CHECKOUT_STATE_COMMITTING_STOCK:
		if topic == messaging.EventTypeCheckoutStockCommitted && result.GetStockCommitted() != nil {
			return &transition{
				state:          checkoutv1.CheckoutState_CHECKOUT_STATE_FINALIZING_PAID,
				commandTopic:   messaging.EventTypeCheckoutFinalizeOrder,
				commandPayload: &checkoutv1.FinalizeOrder{Status: checkoutv1.FinalOrderStatus_FINAL_ORDER_STATUS_PAID},
				operation:      "finalize-paid", deadline: preparationDeadline,
			}, nil
		}
	case checkoutv1.CheckoutState_CHECKOUT_STATE_RELEASING_STOCK:
		if topic == messaging.EventTypeCheckoutStockReleased && result.GetStockReleased() != nil {
			return &transition{
				state:          checkoutv1.CheckoutState_CHECKOUT_STATE_FINALIZING_FAILED,
				commandTopic:   messaging.EventTypeCheckoutFinalizeOrder,
				commandPayload: &checkoutv1.FinalizeOrder{Status: checkoutv1.FinalOrderStatus_FINAL_ORDER_STATUS_FAILED},
				operation:      "finalize-failed", deadline: preparationDeadline,
			}, nil
		}
	case checkoutv1.CheckoutState_CHECKOUT_STATE_FINALIZING_PAID:
		if topic == messaging.EventTypeCheckoutOrderFinalized && result.GetOrderFinalized().GetStatus() == checkoutv1.FinalOrderStatus_FINAL_ORDER_STATUS_PAID {
			return &transition{state: checkoutv1.CheckoutState_CHECKOUT_STATE_COMPLETED}, nil
		}
	case checkoutv1.CheckoutState_CHECKOUT_STATE_FINALIZING_FAILED:
		if topic == messaging.EventTypeCheckoutOrderFinalized && result.GetOrderFinalized().GetStatus() == checkoutv1.FinalOrderStatus_FINAL_ORDER_STATUS_FAILED {
			return &transition{state: checkoutv1.CheckoutState_CHECKOUT_STATE_FAILED}, nil
		}
	}
	return nil, nil
}

func preparationResult(topic string, result *checkoutv1.CheckoutMessage) (participant, outcome, reason string, ok bool) {
	switch topic {
	case messaging.EventTypeCheckoutOrderCreated:
		return "orders", outcomeSucceeded, "", result.GetOrderCreated() != nil
	case messaging.EventTypeCheckoutOrderRejected:
		return "orders", outcomeRejected, result.GetOrderRejected().GetReason(), result.GetOrderRejected() != nil
	case messaging.EventTypeCheckoutStockReserved:
		return "inventory", outcomeSucceeded, "", result.GetStockReserved() != nil
	case messaging.EventTypeCheckoutStockRejected:
		return "inventory", outcomeRejected, result.GetStockRejected().GetReason(), result.GetStockRejected() != nil
	}
	return "", "", "", false
}

func handlePreparationResult(ctx context.Context, queries *database.Queries, row database.Checkout, topic string, result *checkoutv1.CheckoutMessage) error {
	messageContext := result.GetContext()
	participant, outcome, reason, ok := preparationResult(topic, result)
	if !ok {
		return recordPreparationException(ctx, queries, row, messageContext.GetMessageId(), "preparation_result_mismatch", "result payload does not match its topic")
	}
	if !row.PrepOperationID.Valid || messageContext.GetOperationId() != row.PrepOperationID.UUID.String() ||
		!row.PrepVersion.Valid || messageContext.GetWorkflowVersion() != uint64(row.PrepVersion.Int64) {
		return recordPreparationException(ctx, queries, row, messageContext.GetMessageId(), "preparation_correlation_conflict", "result does not match the preparation operation")
	}

	recorded := row.OrdersOutcome
	if participant == "inventory" {
		recorded = row.InventoryOutcome
	}
	switch {
	case recorded == outcome:
		return nil
	case recorded != outcomePending:
		return recordPreparationException(ctx, queries, row, messageContext.GetMessageId(), "preparation_outcome_conflict",
			fmt.Sprintf("%s reported %s after %s", participant, outcome, recorded))
	}
	if err := queries.SetPreparationOutcome(ctx, database.SetPreparationOutcomeParams{
		ID: row.ID, Participant: participant, Outcome: outcome,
		Reason: sql.NullString{String: reason, Valid: reason != ""},
	}); err != nil {
		return err
	}
	if participant == "orders" {
		row.OrdersOutcome = outcome
	} else {
		row.InventoryOutcome = outcome
	}

	state := checkoutv1.CheckoutState(checkoutv1.CheckoutState_value[row.State])
	if state != checkoutv1.CheckoutState_CHECKOUT_STATE_PREPARING || row.PrepClosedAt.Valid {
		return nil
	}
	if outcome == outcomeRejected {
		return closePreparation(ctx, queries, row, closureRejected, preparationFailureReason(participant, reason), messageContext.GetMessageId())
	}
	if row.OrdersOutcome != outcomeSucceeded || row.InventoryOutcome != outcomeSucceeded {
		return nil
	}
	if len(row.GatewaySnapshot) == 0 {
		return errors.New("missing frozen gateway snapshot")
	}
	if _, err := queries.ClosePreparation(ctx, database.ClosePreparationParams{ID: row.ID, PrepClosure: sql.NullString{String: closurePrepared, Valid: true}}); err != nil {
		return err
	}
	updated, err := queries.AdvanceCheckout(ctx, database.AdvanceCheckoutParams{
		ID: row.ID, Version: row.Version, State: checkoutv1.CheckoutState_CHECKOUT_STATE_CREATING_PAYMENT.String(),
		DeadlineAt: sql.NullTime{Time: time.Now().Add(preparationDeadline), Valid: true},
	})
	if err != nil {
		return err
	}
	if updated != 1 {
		return errors.New("checkout transition lost optimistic concurrency race")
	}
	return queries.InsertGatewayOperation(ctx, database.InsertGatewayOperationParams{
		CheckoutID: row.ID, Request: row.GatewaySnapshot, Tracingspancontext: sql.NullString{String: sharedtrace.SerializeContext(ctx), Valid: true},
	})
}

func preparationFailureReason(participant, reason string) string {
	if reason == "" {
		return participant + " preparation rejected"
	}
	return reason
}

func closePreparation(ctx context.Context, queries *database.Queries, row database.Checkout, closure, reason, causationID string) error {
	closed, err := queries.ClosePreparation(ctx, database.ClosePreparationParams{ID: row.ID, PrepClosure: sql.NullString{String: closure, Valid: true}})
	if err != nil {
		return err
	}
	if closed != 1 {
		return nil
	}
	updated, err := queries.AdvanceCheckout(ctx, database.AdvanceCheckoutParams{
		ID: row.ID, Version: row.Version, State: checkoutv1.CheckoutState_CHECKOUT_STATE_RELEASING_STOCK.String(),
		FailureReason: sql.NullString{String: reason, Valid: reason != ""},
		DeadlineAt:    sql.NullTime{Time: time.Now().Add(preparationDeadline), Valid: true},
	})
	if err != nil {
		return err
	}
	if updated != 1 {
		return errors.New("checkout transition lost optimistic concurrency race")
	}
	return enqueueCheckoutCommand(ctx, queries, row, messaging.EventTypeCheckoutCancelStock, "cancel-stock", causationID, &checkoutv1.CancelStock{Reason: reason})
}

func recordPreparationException(ctx context.Context, queries *database.Queries, row database.Checkout, messageID, kind, details string) error {
	return queries.InsertException(ctx, database.InsertExceptionParams{
		ID: uuid.New(), CheckoutID: row.ID, MessageID: sql.NullString{String: messageID, Valid: messageID != ""}, Kind: kind, Details: details,
	})
}
