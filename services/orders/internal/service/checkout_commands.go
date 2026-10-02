package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/phuchoang2603/refurbished-marketplace/services/orders/internal/database"
	"github.com/phuchoang2603/refurbished-marketplace/shared/messaging"
	sharedtrace "github.com/phuchoang2603/refurbished-marketplace/shared/observe/trace"
	checkoutv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/checkout/v1"
	"google.golang.org/protobuf/proto"
)

func (service *Service) HandleCheckoutCommand(ctx context.Context, record messaging.KafkaMessage) error {
	command := new(checkoutv1.CheckoutMessage)
	if err := proto.Unmarshal(record.Value, command); err != nil {
		return err
	}
	messageContext := command.GetContext()
	if messageContext == nil || messageContext.GetMessageId() == "" || messageContext.GetOperationId() == "" || messageContext.GetWorkflowVersion() == 0 {
		return errors.New("checkout order command missing identity")
	}
	checkoutID, err := uuid.Parse(messageContext.GetCheckoutId())
	if err != nil || checkoutID == uuid.Nil {
		return errors.New("invalid checkout id")
	}
	orderID, err := uuid.Parse(messageContext.GetOrderId())
	if err != nil || orderID == uuid.Nil {
		return errors.New("invalid order id")
	}
	tx, err := service.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	queries := database.New(tx)
	if _, err := queries.InsertOrdersInboxMessage(ctx, messageContext.GetMessageId()); errors.Is(err, sql.ErrNoRows) {
		return tx.Commit()
	} else if err != nil {
		return err
	}

	var resultPayload any
	var resultTopic string
	switch record.Topic {
	case messaging.EventTypeCheckoutCreateOrder:
		if command.GetCreateOrder() == nil {
			return errors.New("order create payload missing")
		}
		if err := createCheckoutOrder(ctx, queries, checkoutID, orderID, command.GetCreateOrder()); err != nil {
			return err
		}
		resultPayload = &checkoutv1.CheckoutMessage_OrderCreated{OrderCreated: &checkoutv1.OrderCreated{}}
		resultTopic = messaging.EventTypeCheckoutOrderCreated
	case messaging.EventTypeCheckoutFinalizeOrder:
		finalize := command.GetFinalizeOrder()
		if finalize == nil || (finalize.GetStatus() != checkoutv1.FinalOrderStatus_FINAL_ORDER_STATUS_PAID && finalize.GetStatus() != checkoutv1.FinalOrderStatus_FINAL_ORDER_STATUS_FAILED) {
			return errors.New("invalid checkout finalization")
		}
		owner, err := queries.GetCheckoutOrderCommand(ctx, checkoutID)
		if err != nil || owner.OrderID != orderID {
			return fmt.Errorf("finalize order checkout binding: %w", err)
		}
		status := OrderStatusPaid
		if finalize.GetStatus() == checkoutv1.FinalOrderStatus_FINAL_ORDER_STATUS_FAILED {
			status = OrderStatusFailed
		}
		if _, err := queries.FinalizeCheckoutOrder(ctx, database.FinalizeCheckoutOrderParams{ID: orderID, Status: status}); err != nil {
			return fmt.Errorf("finalize order terminal conflict: %w", err)
		}
		resultPayload = &checkoutv1.CheckoutMessage_OrderFinalized{OrderFinalized: &checkoutv1.OrderFinalized{Status: finalize.GetStatus()}}
		resultTopic = messaging.EventTypeCheckoutOrderFinalized
	default:
		return fmt.Errorf("unsupported checkout order topic %s", record.Topic)
	}
	resultMessageID := uuid.New()
	result := &checkoutv1.CheckoutMessage{Context: &checkoutv1.MessageContext{
		MessageId: resultMessageID.String(), CheckoutId: checkoutID.String(), OrderId: orderID.String(),
		OperationId: messageContext.GetOperationId(), CausationId: messageContext.GetMessageId(),
		WorkflowVersion: messageContext.GetWorkflowVersion(),
	}}
	switch value := resultPayload.(type) {
	case *checkoutv1.CheckoutMessage_OrderCreated:
		result.Payload = value
	case *checkoutv1.CheckoutMessage_OrderFinalized:
		result.Payload = value
	}
	payload, err := proto.Marshal(result)
	if err != nil {
		return err
	}
	if _, err := queries.CreateOrderOutbox(ctx, database.CreateOrderOutboxParams{
		ID: resultMessageID, AggregateID: checkoutID, EventType: resultTopic,
		Payload: payload, Tracingspancontext: sharedtrace.SerializeContext(ctx),
	}); err != nil {
		return err
	}
	return tx.Commit()
}

func createCheckoutOrder(ctx context.Context, queries *database.Queries, checkoutID, orderID uuid.UUID, request *checkoutv1.CreateOrder) error {
	buyerID, buyerError := uuid.Parse(request.GetBuyerUserId())
	merchantID, merchantError := uuid.Parse(request.GetMerchantId())
	if buyerError != nil || merchantError != nil {
		return errors.New("invalid checkout order identities")
	}
	items := make([]OrderItemInput, 0, len(request.GetItems()))
	for _, item := range request.GetItems() {
		if item == nil {
			return errors.New("nil checkout item")
		}
		productID, err := uuid.Parse(item.GetProductId())
		if err != nil {
			return err
		}
		items = append(items, OrderItemInput{ProductID: productID, Quantity: item.GetQuantity(), UnitPriceCents: item.GetUnitPriceCents()})
	}
	if err := validateCreateOrderInput(buyerID, merchantID, checkoutID, items, request.GetTotalCents()); err != nil {
		return err
	}
	encoded, err := (proto.MarshalOptions{Deterministic: true}).Marshal(request)
	if err != nil {
		return err
	}
	hash := sha256.Sum256(encoded)
	existing, err := queries.GetCheckoutOrderCommand(ctx, checkoutID)
	if err == nil {
		if existing.OrderID != orderID || !bytes.Equal(existing.RequestHash, hash[:]) {
			return ErrIdempotencyConflict
		}
		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if _, err := queries.CreateOrder(ctx, database.CreateOrderParams{
		ID: orderID, BuyerUserID: buyerID, MerchantID: merchantID,
		Status: OrderStatusPending, TotalCents: request.GetTotalCents(), IdempotencyKey: checkoutID,
	}); err != nil {
		return err
	}
	if _, err := createOrderItems(ctx, queries, orderID, items); err != nil {
		return err
	}
	return queries.InsertCheckoutOrderCommand(ctx, database.InsertCheckoutOrderCommandParams{
		CheckoutID: checkoutID, OrderID: orderID, RequestHash: hash[:],
	})
}
