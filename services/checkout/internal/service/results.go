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
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const preparationDeadline = 2 * time.Minute

type transition struct {
	state          checkoutv1.CheckoutState
	commandTopic   string
	commandPayload any
	operation      string
	deadline       time.Duration
	sessionID      string
	returnURL      string
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
	defer tx.Rollback()
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

	if resultIsFinancialException(row, kafkaMessage.Topic, message) {
		if err := queries.InsertException(ctx, database.InsertExceptionParams{
			ID: uuid.New(), CheckoutID: checkoutID,
			MessageID: sql.NullString{String: messageContext.GetMessageId(), Valid: true},
			Kind:      "contradictory_payment", Details: "payment result contradicts checkout settlement",
		}); err != nil {
			return err
		}
		if row.State != checkoutv1.CheckoutState_CHECKOUT_STATE_FAILED.String() &&
			row.State != checkoutv1.CheckoutState_CHECKOUT_STATE_COMPLETED.String() {
			if _, err := queries.AdvanceCheckout(ctx, database.AdvanceCheckoutParams{
				ID: checkoutID, Version: row.Version, State: checkoutv1.CheckoutState_CHECKOUT_STATE_NEEDS_REVIEW.String(),
			}); err != nil {
				return err
			}
		}
		return tx.Commit()
	}
	step, err := nextTransition(row, kafkaMessage.Topic, message)
	if err != nil {
		return err
	}
	if step == nil {
		if row.State == checkoutv1.CheckoutState_CHECKOUT_STATE_CREATING_SESSION.String() &&
			(kafkaMessage.Topic == messaging.EventTypeCheckoutPaymentSucceeded ||
				kafkaMessage.Topic == messaging.EventTypeCheckoutPaymentFailed ||
				kafkaMessage.Topic == messaging.EventTypeCheckoutPaymentUncertain) {
			return fmt.Errorf("payment outcome arrived before session readiness for checkout %s", checkoutID)
		}
		if messageContext.GetWorkflowVersion() > uint64(row.Version) {
			return fmt.Errorf("checkout result arrived before its command: checkout %s version %d, result version %d", checkoutID, row.Version, messageContext.GetWorkflowVersion())
		}
		return tx.Commit()
	}
	if step.state == checkoutv1.CheckoutState_CHECKOUT_STATE_NEEDS_REVIEW {
		if err := queries.InsertException(ctx, database.InsertExceptionParams{
			ID: uuid.New(), CheckoutID: checkoutID,
			MessageID: sql.NullString{String: messageContext.GetMessageId(), Valid: true},
			Kind:      "payment_uncertain", Details: "payment processor could not confirm capture outcome",
		}); err != nil {
			return err
		}
	}
	var deadline sql.NullTime
	if step.deadline > 0 {
		deadline = sql.NullTime{Time: time.Now().Add(step.deadline), Valid: true}
	}
	updated, err := queries.AdvanceCheckout(ctx, database.AdvanceCheckoutParams{
		ID: checkoutID, Version: row.Version, State: step.state.String(),
		PaymentSessionID: sql.NullString{String: step.sessionID, Valid: step.sessionID != ""},
		PaymentReturnUrl: sql.NullString{String: step.returnURL, Valid: step.returnURL != ""},
		FailureReason:    sql.NullString{String: step.reason, Valid: step.reason != ""},
		DeadlineAt:       deadline,
	})
	if err != nil {
		return err
	}
	if updated != 1 {
		return errors.New("checkout transition lost optimistic concurrency race")
	}
	if step.commandTopic != "" {
		operationID := uuid.NewSHA1(checkoutID, []byte(step.operation))
		command := &checkoutv1.CheckoutMessage{Context: &checkoutv1.MessageContext{
			CheckoutId: checkoutID.String(), OrderId: row.OrderID.UUID.String(),
			MessageId: operationID.String(), OperationId: operationID.String(),
			CausationId: messageContext.GetMessageId(), WorkflowVersion: uint64(row.Version + 1),
		}}
		if err := setCommandPayload(command, step.commandPayload); err != nil {
			return err
		}
		payload, err := proto.Marshal(command)
		if err != nil {
			return err
		}
		if err := queries.InsertOutbox(ctx, database.InsertOutboxParams{
			ID: operationID, AggregateID: checkoutID, EventType: step.commandTopic,
			Payload: payload, Tracingspancontext: sql.NullString{String: sharedtrace.SerializeContext(ctx), Valid: true},
		}); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func resultIsFinancialException(row database.Checkout, topic string, result *checkoutv1.CheckoutMessage) bool {
	if row.PaymentSessionID.String == "" {
		state := checkoutv1.CheckoutState(checkoutv1.CheckoutState_value[row.State])
		return topic == messaging.EventTypeCheckoutPaymentSucceeded && result.GetPaymentSucceeded() != nil &&
			result.GetPaymentSucceeded().GetPaymentSessionId() != "" &&
			(state == checkoutv1.CheckoutState_CHECKOUT_STATE_RECONCILING || state == checkoutv1.CheckoutState_CHECKOUT_STATE_NEEDS_REVIEW)
	}
	state := checkoutv1.CheckoutState(checkoutv1.CheckoutState_value[row.State])
	if topic == messaging.EventTypeCheckoutPaymentSucceeded &&
		result.GetPaymentSucceeded() != nil &&
		result.GetPaymentSucceeded().GetPaymentSessionId() == row.PaymentSessionID.String {
		return state == checkoutv1.CheckoutState_CHECKOUT_STATE_RELEASING_STOCK ||
			state == checkoutv1.CheckoutState_CHECKOUT_STATE_FINALIZING_FAILED ||
			state == checkoutv1.CheckoutState_CHECKOUT_STATE_FAILED
	}
	if topic == messaging.EventTypeCheckoutPaymentFailed &&
		result.GetPaymentFailed() != nil &&
		result.GetPaymentFailed().GetPaymentSessionId() == row.PaymentSessionID.String {
		return state == checkoutv1.CheckoutState_CHECKOUT_STATE_COMMITTING_STOCK ||
			state == checkoutv1.CheckoutState_CHECKOUT_STATE_FINALIZING_PAID ||
			state == checkoutv1.CheckoutState_CHECKOUT_STATE_COMPLETED
	}
	return false
}

func nextTransition(row database.Checkout, topic string, result *checkoutv1.CheckoutMessage) (*transition, error) {
	current := checkoutv1.CheckoutState(checkoutv1.CheckoutState_value[row.State])
	context := result.GetContext()
	expectedStep := ""
	expectedVersion := row.Version
	switch current {
	case checkoutv1.CheckoutState_CHECKOUT_STATE_CREATING_ORDER:
		expectedStep = "create-order"
	case checkoutv1.CheckoutState_CHECKOUT_STATE_RESERVING_STOCK:
		expectedStep = "reserve-stock"
	case checkoutv1.CheckoutState_CHECKOUT_STATE_CREATING_SESSION:
		expectedStep = "create-session"
	case checkoutv1.CheckoutState_CHECKOUT_STATE_READY_TO_PAY:
		expectedStep = "create-session"
		expectedVersion--
	case checkoutv1.CheckoutState_CHECKOUT_STATE_COMMITTING_STOCK:
		expectedStep = "commit-stock"
	case checkoutv1.CheckoutState_CHECKOUT_STATE_RELEASING_STOCK:
		expectedStep = "cancel-stock"
	case checkoutv1.CheckoutState_CHECKOUT_STATE_FINALIZING_PAID:
		expectedStep = "finalize-paid"
	case checkoutv1.CheckoutState_CHECKOUT_STATE_FINALIZING_FAILED:
		expectedStep = "finalize-failed"
	case checkoutv1.CheckoutState_CHECKOUT_STATE_RECONCILING,
		checkoutv1.CheckoutState_CHECKOUT_STATE_NEEDS_REVIEW:
		expectedStep = "cancel-session"
	}
	validExpected := expectedStep != "" && context.GetOperationId() == uuid.NewSHA1(row.ID, []byte(expectedStep)).String() &&
		(context.GetWorkflowVersion() == uint64(expectedVersion) ||
			current == checkoutv1.CheckoutState_CHECKOUT_STATE_NEEDS_REVIEW && context.GetWorkflowVersion() == uint64(row.Version-1))
	if !validExpected {
		if (current != checkoutv1.CheckoutState_CHECKOUT_STATE_RECONCILING && current != checkoutv1.CheckoutState_CHECKOUT_STATE_NEEDS_REVIEW) ||
			context.GetOperationId() != uuid.NewSHA1(row.ID, []byte("create-session")).String() ||
			context.GetWorkflowVersion() == 0 || context.GetWorkflowVersion() > uint64(row.Version) {
			return nil, nil
		}
	}

	snapshot := new(checkoutv1.SubmitCheckoutRequest)
	if err := protojson.Unmarshal(row.Snapshot, snapshot); err != nil {
		return nil, err
	}
	switch current {
	case checkoutv1.CheckoutState_CHECKOUT_STATE_CREATING_ORDER:
		switch topic {
		case messaging.EventTypeCheckoutOrderCreated:
			if result.GetOrderCreated() == nil {
				break
			}
			return &transition{
				state:        checkoutv1.CheckoutState_CHECKOUT_STATE_RESERVING_STOCK,
				commandTopic: messaging.EventTypeCheckoutReserveStock, commandPayload: &checkoutv1.ReserveStock{
					MerchantId: row.MerchantID.String(), TotalCents: snapshot.GetTotalCents(), Items: snapshot.GetItems(),
				}, operation: "reserve-stock", deadline: preparationDeadline,
			}, nil
		case messaging.EventTypeCheckoutOrderRejected:
			if result.GetOrderRejected() == nil {
				break
			}
			return &transition{state: checkoutv1.CheckoutState_CHECKOUT_STATE_FAILED, reason: result.GetOrderRejected().GetReason()}, nil
		}
	case checkoutv1.CheckoutState_CHECKOUT_STATE_RESERVING_STOCK:
		switch topic {
		case messaging.EventTypeCheckoutStockReserved:
			if result.GetStockReserved() == nil {
				break
			}
			return &transition{
				state:        checkoutv1.CheckoutState_CHECKOUT_STATE_CREATING_SESSION,
				commandTopic: messaging.EventTypeCheckoutCreateSession, commandPayload: &checkoutv1.CreatePaymentSession{
					BuyerUserId: row.BuyerUserID.String(), BuyerEmail: snapshot.GetBuyerEmail(),
					MerchantId: row.MerchantID.String(), TotalCents: snapshot.GetTotalCents(),
					Currency: snapshot.GetCurrency(), ShippingAddress: snapshot.GetShippingAddress(),
					Items: snapshot.GetItems(), ReturnUrl: snapshot.GetReturnUrl(),
					ReservationOperationId: context.GetOperationId(),
				}, operation: "create-session", deadline: preparationDeadline,
			}, nil
		case messaging.EventTypeCheckoutStockRejected:
			if result.GetStockRejected() == nil {
				break
			}
			return &transition{
				state:          checkoutv1.CheckoutState_CHECKOUT_STATE_RELEASING_STOCK,
				commandTopic:   messaging.EventTypeCheckoutCancelStock,
				commandPayload: &checkoutv1.CancelStock{Reason: "reservation rejected"},
				operation:      "cancel-stock", deadline: preparationDeadline, reason: result.GetStockRejected().GetReason(),
			}, nil
		}
	case checkoutv1.CheckoutState_CHECKOUT_STATE_CREATING_SESSION:
		switch topic {
		case messaging.EventTypeCheckoutSessionReady:
			ready := result.GetPaymentSessionReady()
			if ready == nil || ready.GetPaymentSessionId() == "" || ready.GetReturnUrl() == "" {
				break
			}
			if ready.GetExpiresAtUnixMs() <= time.Now().UnixMilli() {
				break
			}
			return &transition{
				state:     checkoutv1.CheckoutState_CHECKOUT_STATE_READY_TO_PAY,
				sessionID: ready.GetPaymentSessionId(), returnURL: ready.GetReturnUrl(),
				deadline: time.Until(time.UnixMilli(ready.GetExpiresAtUnixMs())),
			}, nil
		case messaging.EventTypeCheckoutSessionRejected:
			if result.GetPaymentSessionRejected() == nil {
				break
			}
			return &transition{
				state:          checkoutv1.CheckoutState_CHECKOUT_STATE_RELEASING_STOCK,
				commandTopic:   messaging.EventTypeCheckoutCancelStock,
				commandPayload: &checkoutv1.CancelStock{Reason: "session creation rejected"},
				operation:      "cancel-stock", deadline: preparationDeadline, reason: result.GetPaymentSessionRejected().GetReason(),
			}, nil
		}
	case checkoutv1.CheckoutState_CHECKOUT_STATE_READY_TO_PAY:
		switch topic {
		case messaging.EventTypeCheckoutPaymentSucceeded:
			if result.GetPaymentSucceeded() == nil || result.GetPaymentSucceeded().GetPaymentSessionId() != row.PaymentSessionID.String {
				break
			}
			return &transition{
				state:        checkoutv1.CheckoutState_CHECKOUT_STATE_COMMITTING_STOCK,
				commandTopic: messaging.EventTypeCheckoutCommitStock, commandPayload: &checkoutv1.CommitStock{},
				operation: "commit-stock", deadline: preparationDeadline,
			}, nil
		case messaging.EventTypeCheckoutPaymentFailed, messaging.EventTypeCheckoutPaymentCancelled:
			if topic == messaging.EventTypeCheckoutPaymentFailed {
				if result.GetPaymentFailed() == nil || result.GetPaymentFailed().GetPaymentSessionId() != row.PaymentSessionID.String {
					break
				}
			} else if result.GetPaymentCancelled() == nil {
				break
			}
			return &transition{
				state:          checkoutv1.CheckoutState_CHECKOUT_STATE_RELEASING_STOCK,
				commandTopic:   messaging.EventTypeCheckoutCancelStock,
				commandPayload: &checkoutv1.CancelStock{Reason: "payment definitively failed"},
				operation:      "cancel-stock", deadline: preparationDeadline,
			}, nil
		case messaging.EventTypeCheckoutPaymentUncertain:
			if result.GetPaymentUncertain() != nil && result.GetPaymentUncertain().GetPaymentSessionId() == row.PaymentSessionID.String {
				return &transition{state: checkoutv1.CheckoutState_CHECKOUT_STATE_NEEDS_REVIEW}, nil
			}
		}
	case checkoutv1.CheckoutState_CHECKOUT_STATE_RECONCILING,
		checkoutv1.CheckoutState_CHECKOUT_STATE_NEEDS_REVIEW:
		switch topic {
		case messaging.EventTypeCheckoutPaymentSucceeded:
			if result.GetPaymentSucceeded() != nil && row.PaymentSessionID.Valid &&
				result.GetPaymentSucceeded().GetPaymentSessionId() == row.PaymentSessionID.String {
				return &transition{
					state:        checkoutv1.CheckoutState_CHECKOUT_STATE_COMMITTING_STOCK,
					commandTopic: messaging.EventTypeCheckoutCommitStock, commandPayload: &checkoutv1.CommitStock{},
					operation: "commit-stock", deadline: preparationDeadline,
				}, nil
			}
		case messaging.EventTypeCheckoutPaymentCancelled, messaging.EventTypeCheckoutPaymentFailed:
			if topic == messaging.EventTypeCheckoutPaymentCancelled && result.GetPaymentCancelled() == nil {
				break
			}
			if topic == messaging.EventTypeCheckoutPaymentFailed &&
				(result.GetPaymentFailed() == nil || result.GetPaymentFailed().GetPaymentSessionId() != row.PaymentSessionID.String) {
				break
			}
			return &transition{
				state:          checkoutv1.CheckoutState_CHECKOUT_STATE_RELEASING_STOCK,
				commandTopic:   messaging.EventTypeCheckoutCancelStock,
				commandPayload: &checkoutv1.CancelStock{Reason: "payment cancellation confirmed"},
				operation:      "cancel-stock", deadline: preparationDeadline,
			}, nil
		case messaging.EventTypeCheckoutPaymentUncertain:
			if result.GetPaymentUncertain() != nil {
				return &transition{state: checkoutv1.CheckoutState_CHECKOUT_STATE_NEEDS_REVIEW}, nil
			}
		}
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

func setCommandPayload(message *checkoutv1.CheckoutMessage, payload any) error {
	switch value := payload.(type) {
	case *checkoutv1.ReserveStock:
		message.Payload = &checkoutv1.CheckoutMessage_ReserveStock{ReserveStock: value}
	case *checkoutv1.CreatePaymentSession:
		message.Payload = &checkoutv1.CheckoutMessage_CreatePaymentSession{CreatePaymentSession: value}
	case *checkoutv1.CancelStock:
		message.Payload = &checkoutv1.CheckoutMessage_CancelStock{CancelStock: value}
	case *checkoutv1.CommitStock:
		message.Payload = &checkoutv1.CheckoutMessage_CommitStock{CommitStock: value}
	case *checkoutv1.FinalizeOrder:
		message.Payload = &checkoutv1.CheckoutMessage_FinalizeOrder{FinalizeOrder: value}
	default:
		return fmt.Errorf("unknown checkout command payload %T", payload)
	}
	return nil
}
