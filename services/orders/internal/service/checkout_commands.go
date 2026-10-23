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

const (
	checkoutOrderCreated      = "CREATED"
	checkoutOrderRejected     = "REJECTED"
	checkoutOrderFailedAbsent = "FAILED_ABSENT"
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
	sharedtrace.AnnotateCheckout(ctx, checkoutID.String(), orderID.String())
	tx, err := service.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	queries := database.New(tx)
	if err := queries.LockCheckoutIdentity(ctx, checkoutID.String()); err != nil {
		return err
	}
	if _, err := queries.InsertOrdersInboxMessage(ctx, messageContext.GetMessageId()); errors.Is(err, sql.ErrNoRows) {
		return tx.Commit()
	} else if err != nil {
		return err
	}

	var resultPayload any
	var resultTopic string
	switch record.Topic {
	case messaging.EventTypeCheckoutPrepare:
		if command.GetPrepareCheckout() == nil {
			return errors.New("checkout preparation payload missing")
		}
		rejection, err := prepareCheckoutOrder(ctx, queries, checkoutID, orderID, command.GetPrepareCheckout())
		if err != nil {
			return err
		}
		if rejection != "" {
			resultPayload = &checkoutv1.CheckoutMessage_OrderRejected{OrderRejected: &checkoutv1.OrderRejected{Reason: rejection}}
			resultTopic = messaging.EventTypeCheckoutOrderRejected
		} else {
			resultPayload = &checkoutv1.CheckoutMessage_OrderCreated{OrderCreated: &checkoutv1.OrderCreated{}}
			resultTopic = messaging.EventTypeCheckoutOrderCreated
		}
	case messaging.EventTypeCheckoutFinalizeOrder:
		finalize := command.GetFinalizeOrder()
		if finalize == nil || (finalize.GetStatus() != checkoutv1.FinalOrderStatus_FINAL_ORDER_STATUS_PAID && finalize.GetStatus() != checkoutv1.FinalOrderStatus_FINAL_ORDER_STATUS_FAILED) {
			return errors.New("invalid checkout finalization")
		}
		if err := finalizeCheckoutOrder(ctx, queries, checkoutID, orderID, finalize.GetStatus()); err != nil {
			return err
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
	case *checkoutv1.CheckoutMessage_OrderRejected:
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

func prepareCheckoutOrder(ctx context.Context, queries *database.Queries, checkoutID, orderID uuid.UUID, request *checkoutv1.PrepareCheckout) (string, error) {
	encoded, err := (proto.MarshalOptions{Deterministic: true}).Marshal(request)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(encoded)
	existing, err := queries.GetCheckoutOrderCommand(ctx, checkoutID)
	if err == nil {
		if existing.OrderID != orderID {
			return "", ErrIdempotencyConflict
		}
		if existing.Status == checkoutOrderFailedAbsent {
			return "checkout already failed before order creation", nil
		}
		if !bytes.Equal(existing.RequestHash, hash[:]) {
			return "", ErrIdempotencyConflict
		}
		if existing.Status == checkoutOrderRejected {
			return existing.Reason.String, nil
		}
		return "", nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}

	buyerID, merchantID, items, err := parsePreparedOrder(checkoutID, request)
	if err != nil {
		if insertErr := queries.InsertCheckoutOrderCommand(ctx, database.InsertCheckoutOrderCommandParams{
			CheckoutID: checkoutID, OrderID: orderID, RequestHash: hash[:],
			Status: checkoutOrderRejected, Reason: sql.NullString{String: err.Error(), Valid: true},
		}); insertErr != nil {
			return "", insertErr
		}
		return err.Error(), nil
	}
	if _, err := queries.CreateOrder(ctx, database.CreateOrderParams{
		ID: orderID, BuyerUserID: buyerID, MerchantID: merchantID,
		Status: OrderStatusPending, TotalCents: request.GetTotalCents(), IdempotencyKey: checkoutID,
	}); err != nil {
		return "", err
	}
	if _, err := createOrderItems(ctx, queries, orderID, items); err != nil {
		return "", err
	}
	return "", queries.InsertCheckoutOrderCommand(ctx, database.InsertCheckoutOrderCommandParams{
		CheckoutID: checkoutID, OrderID: orderID, RequestHash: hash[:], Status: checkoutOrderCreated,
	})
}

func parsePreparedOrder(checkoutID uuid.UUID, request *checkoutv1.PrepareCheckout) (uuid.UUID, uuid.UUID, []OrderItemInput, error) {
	buyerID, buyerError := uuid.Parse(request.GetBuyerUserId())
	merchantID, merchantError := uuid.Parse(request.GetMerchantId())
	if buyerError != nil || merchantError != nil {
		return uuid.Nil, uuid.Nil, nil, errors.New("invalid checkout order identities")
	}
	items := make([]OrderItemInput, 0, len(request.GetItems()))
	for _, item := range request.GetItems() {
		if item == nil {
			return uuid.Nil, uuid.Nil, nil, errors.New("nil checkout item")
		}
		productID, err := uuid.Parse(item.GetProductId())
		if err != nil {
			return uuid.Nil, uuid.Nil, nil, ErrInvalidProductID
		}
		items = append(items, OrderItemInput{ProductID: productID, Quantity: item.GetQuantity(), UnitPriceCents: item.GetUnitPriceCents()})
	}
	if err := validateCreateOrderInput(buyerID, merchantID, checkoutID, items, request.GetTotalCents()); err != nil {
		return uuid.Nil, uuid.Nil, nil, err
	}
	return buyerID, merchantID, items, nil
}

func finalizeCheckoutOrder(ctx context.Context, queries *database.Queries, checkoutID, orderID uuid.UUID, finalStatus checkoutv1.FinalOrderStatus) error {
	failed := finalStatus == checkoutv1.FinalOrderStatus_FINAL_ORDER_STATUS_FAILED
	owner, err := queries.GetCheckoutOrderCommand(ctx, checkoutID)
	if errors.Is(err, sql.ErrNoRows) {
		if !failed {
			return errors.New("paid finalization requires an existing checkout order")
		}
		return queries.InsertCheckoutOrderCommand(ctx, database.InsertCheckoutOrderCommandParams{
			CheckoutID: checkoutID, OrderID: orderID, Status: checkoutOrderFailedAbsent,
		})
	}
	if err != nil {
		return err
	}
	if owner.OrderID != orderID {
		return errors.New("finalize order checkout binding conflict")
	}
	if owner.Status != checkoutOrderCreated {
		if !failed {
			return fmt.Errorf("paid finalization conflicts with %s checkout order", owner.Status)
		}
		return nil
	}
	status := OrderStatusPaid
	if failed {
		status = OrderStatusFailed
	}
	if _, err := queries.FinalizeCheckoutOrder(ctx, database.FinalizeCheckoutOrderParams{ID: orderID, Status: status}); err != nil {
		return fmt.Errorf("finalize order terminal conflict: %w", err)
	}
	return nil
}
