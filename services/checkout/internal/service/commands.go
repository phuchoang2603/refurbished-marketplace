package service

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/phuchoang2603/refurbished-marketplace/services/checkout/internal/database"
	sharedtrace "github.com/phuchoang2603/refurbished-marketplace/shared/observe/trace"
	checkoutv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/checkout/v1"
	"google.golang.org/protobuf/proto"
)

func enqueueCheckoutCommand(ctx context.Context, queries *database.Queries, row database.Checkout, topic, operation, causationID string, payload any) error {
	operationID := uuid.NewSHA1(row.ID, []byte(operation))
	command := &checkoutv1.CheckoutMessage{Context: &checkoutv1.MessageContext{
		CheckoutId: row.ID.String(), OrderId: row.OrderID.UUID.String(),
		MessageId: operationID.String(), OperationId: operationID.String(),
		CausationId: causationID, WorkflowVersion: uint64(row.Version + 1),
	}}
	if err := setCommandPayload(command, payload); err != nil {
		return err
	}
	encoded, err := proto.Marshal(command)
	if err != nil {
		return err
	}
	return queries.InsertOutbox(ctx, database.InsertOutboxParams{
		ID: operationID, AggregateID: row.ID, EventType: topic, Payload: encoded,
		Tracingspancontext: sql.NullString{String: sharedtrace.SerializeContext(ctx), Valid: true},
	})
}

func setCommandPayload(message *checkoutv1.CheckoutMessage, payload any) error {
	switch value := payload.(type) {
	case *checkoutv1.ReserveStock:
		message.Payload = &checkoutv1.CheckoutMessage_ReserveStock{ReserveStock: value}
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
