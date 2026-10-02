package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"math"

	"github.com/google/uuid"
	"github.com/phuchoang2603/refurbished-marketplace/services/inventory/internal/database"
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
	if messageContext == nil || messageContext.GetMessageId() == "" || messageContext.GetWorkflowVersion() == 0 {
		return errors.New("inventory command missing identity")
	}
	checkoutID, checkoutError := uuid.Parse(messageContext.GetCheckoutId())
	orderID, orderError := uuid.Parse(messageContext.GetOrderId())
	if checkoutError != nil || orderError != nil || checkoutID == uuid.Nil || orderID == uuid.Nil {
		return errors.New("inventory command missing checkout or order")
	}

	tx, err := service.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	queries := database.New(tx)
	if err := queries.LockInventoryReservationOrder(ctx, orderID); err != nil {
		return err
	}
	if _, err := queries.InsertInventoryInboxMessage(ctx, messageContext.GetMessageId()); errors.Is(err, sql.ErrNoRows) {
		return tx.Commit()
	} else if err != nil {
		return err
	}

	var resultTopic string
	var resultPayload any
	switch record.Topic {
	case messaging.EventTypeCheckoutReserveStock:
		if command.GetReserveStock() == nil {
			return errors.New("missing reserve stock")
		}
		resultTopic, resultPayload, err = reserveCheckoutStock(ctx, queries, checkoutID, orderID, command.GetReserveStock())
	case messaging.EventTypeCheckoutCancelStock:
		if command.GetCancelStock() == nil {
			return errors.New("missing cancel stock")
		}
		resultTopic, resultPayload, err = settleCheckoutStock(ctx, queries, checkoutID, orderID, false)
	case messaging.EventTypeCheckoutCommitStock:
		if command.GetCommitStock() == nil {
			return errors.New("missing commit stock")
		}
		resultTopic, resultPayload, err = settleCheckoutStock(ctx, queries, checkoutID, orderID, true)
	default:
		return fmt.Errorf("unsupported inventory checkout topic %s", record.Topic)
	}
	if err != nil {
		return err
	}
	resultID := uuid.New()
	result := &checkoutv1.CheckoutMessage{Context: &checkoutv1.MessageContext{
		MessageId: resultID.String(), CheckoutId: checkoutID.String(), OrderId: orderID.String(),
		OperationId: messageContext.GetOperationId(), CausationId: messageContext.GetMessageId(),
		WorkflowVersion: messageContext.GetWorkflowVersion(),
	}}
	switch payload := resultPayload.(type) {
	case *checkoutv1.StockReserved:
		result.Payload = &checkoutv1.CheckoutMessage_StockReserved{StockReserved: payload}
	case *checkoutv1.StockRejected:
		result.Payload = &checkoutv1.CheckoutMessage_StockRejected{StockRejected: payload}
	case *checkoutv1.StockReleased:
		result.Payload = &checkoutv1.CheckoutMessage_StockReleased{StockReleased: payload}
	case *checkoutv1.StockCommitted:
		result.Payload = &checkoutv1.CheckoutMessage_StockCommitted{StockCommitted: payload}
	}
	payload, err := proto.Marshal(result)
	if err != nil {
		return err
	}
	if _, err := queries.CreateInventoryOutbox(ctx, database.CreateInventoryOutboxParams{
		ID: resultID, AggregateID: checkoutID, EventType: resultTopic,
		Payload: payload, Tracingspancontext: sharedtrace.SerializeContext(ctx),
	}); err != nil {
		return err
	}
	return tx.Commit()
}

func reserveCheckoutStock(ctx context.Context, queries *database.Queries, checkoutID, orderID uuid.UUID, command *checkoutv1.ReserveStock) (string, any, error) {
	merchantID, err := uuid.Parse(command.GetMerchantId())
	if err != nil || merchantID == uuid.Nil || command.GetTotalCents() <= 0 || len(command.GetItems()) == 0 {
		return "", nil, errors.New("invalid reservation facts")
	}
	encoded, err := (proto.MarshalOptions{Deterministic: true}).Marshal(command)
	if err != nil {
		return "", nil, err
	}
	hash := sha256.Sum256(encoded)
	existing, err := queries.GetCheckoutReservation(ctx, orderID)
	if err == nil {
		if existing.CheckoutID != checkoutID || existing.ReserveHash != nil && !bytes.Equal(existing.ReserveHash, hash[:]) {
			return "", nil, errors.New("reservation intent conflicts with original checkout")
		}
		switch existing.Status {
		case ReservationStatusReserved:
			return messaging.EventTypeCheckoutStockReserved, &checkoutv1.StockReserved{}, nil
		case "ABSENT", ReservationStatusReleased, ReservationStatusCommitted, "REJECTED":
			return messaging.EventTypeCheckoutStockRejected, &checkoutv1.StockRejected{Reason: "reservation is already settled or cancelled"}, nil
		}
		return "", nil, errors.New("unknown reservation status")
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", nil, err
	}
	quantities := make(map[uuid.UUID]int32, len(command.GetItems()))
	var total int64
	for _, line := range command.GetItems() {
		if line == nil || line.GetQuantity() <= 0 || line.GetUnitPriceCents() <= 0 || line.GetName() == "" {
			return "", nil, errors.New("invalid reservation line")
		}
		productID, err := uuid.Parse(line.GetProductId())
		if err != nil || productID == uuid.Nil || quantities[productID] > math.MaxInt32-line.GetQuantity() ||
			line.GetUnitPriceCents() > (math.MaxInt64-total)/int64(line.GetQuantity()) {
			return "", nil, errors.New("invalid reservation quantity or amount")
		}
		quantities[productID] += line.GetQuantity()
		total += line.GetUnitPriceCents() * int64(line.GetQuantity())
	}
	if total != command.GetTotalCents() {
		return "", nil, errors.New("reservation total does not match lines")
	}
	items := make([]ReservationItemInput, 0, len(quantities))
	for productID, quantity := range quantities {
		items = append(items, ReservationItemInput{ProductID: productID, Quantity: quantity})
	}
	reserveError := reserveOrderItems(ctx, queries, orderID, items)
	status := ReservationStatusReserved
	resultTopic := messaging.EventTypeCheckoutStockReserved
	var payload any = &checkoutv1.StockReserved{}
	if errors.Is(reserveError, ErrInsufficientStock) || errors.Is(reserveError, ErrInventoryNotFound) {
		status = "REJECTED"
		resultTopic = messaging.EventTypeCheckoutStockRejected
		payload = &checkoutv1.StockRejected{Reason: reserveError.Error()}
	} else if reserveError != nil {
		return "", nil, reserveError
	}
	if err := queries.InsertCheckoutReservation(ctx, database.InsertCheckoutReservationParams{
		OrderID: orderID, CheckoutID: checkoutID, ReserveHash: hash[:], Status: status,
	}); err != nil {
		return "", nil, err
	}
	return resultTopic, payload, nil
}

func settleCheckoutStock(ctx context.Context, queries *database.Queries, checkoutID, orderID uuid.UUID, commit bool) (string, any, error) {
	existing, err := queries.GetCheckoutReservation(ctx, orderID)
	if errors.Is(err, sql.ErrNoRows) {
		if commit {
			return "", nil, errors.New("cannot commit absent stock")
		}
		if err := queries.InsertCheckoutReservation(ctx, database.InsertCheckoutReservationParams{
			OrderID: orderID, CheckoutID: checkoutID, Status: "ABSENT",
		}); err != nil {
			return "", nil, err
		}
		return messaging.EventTypeCheckoutStockReleased, &checkoutv1.StockReleased{}, nil
	}
	if err != nil {
		return "", nil, err
	}
	if existing.CheckoutID != checkoutID {
		return "", nil, errors.New("settlement conflicts with original checkout")
	}
	if commit && existing.Status == ReservationStatusCommitted {
		return messaging.EventTypeCheckoutStockCommitted, &checkoutv1.StockCommitted{}, nil
	}
	if !commit && (existing.Status == "ABSENT" || existing.Status == ReservationStatusReleased || existing.Status == "REJECTED") {
		return messaging.EventTypeCheckoutStockReleased, &checkoutv1.StockReleased{}, nil
	}
	if existing.Status != ReservationStatusReserved {
		return "", nil, errors.New("reservation already settled with conflicting result")
	}
	reservations, err := queries.GetOrderReservations(ctx, orderID)
	if err != nil || len(reservations) == 0 {
		return "", nil, errors.New("reserved order has no stock holds")
	}
	for _, reservation := range reservations {
		if reservation.Status != ReservationStatusReserved {
			return "", nil, errors.New("order has partially settled stock holds")
		}
		if commit {
			if _, err := queries.CommitInventoryReservedStock(ctx, database.CommitInventoryReservedStockParams{
				ProductID: reservation.ProductID, Quantity: reservation.Quantity,
			}); err != nil {
				return "", nil, err
			}
			if _, err := queries.MarkInventoryReservationCommitted(ctx, database.MarkInventoryReservationCommittedParams{
				OrderID: orderID, ProductID: reservation.ProductID,
			}); err != nil {
				return "", nil, err
			}
		} else {
			if _, err := queries.ReleaseInventoryReservedStock(ctx, database.ReleaseInventoryReservedStockParams{
				ProductID: reservation.ProductID, Quantity: reservation.Quantity,
			}); err != nil {
				return "", nil, err
			}
			if _, err := queries.MarkInventoryReservationReleased(ctx, database.MarkInventoryReservationReleasedParams{
				OrderID: orderID, ProductID: reservation.ProductID,
			}); err != nil {
				return "", nil, err
			}
		}
	}
	next := ReservationStatusReleased
	topic := messaging.EventTypeCheckoutStockReleased
	var payload any = &checkoutv1.StockReleased{}
	if commit {
		next = ReservationStatusCommitted
		topic = messaging.EventTypeCheckoutStockCommitted
		payload = &checkoutv1.StockCommitted{}
	}
	updated, err := queries.SetCheckoutReservationStatus(ctx, database.SetCheckoutReservationStatusParams{
		OrderID: orderID, Status: next, Status_2: ReservationStatusReserved,
	})
	if err != nil || updated != 1 {
		return "", nil, errors.New("reservation settlement status changed unexpectedly")
	}
	return topic, payload, nil
}
