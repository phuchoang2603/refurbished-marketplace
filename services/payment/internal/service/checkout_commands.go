package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/phuchoang2603/refurbished-marketplace/services/payment/internal/database"
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
		return errors.New("payment checkout command missing identity")
	}
	checkoutID, checkoutError := uuid.Parse(messageContext.GetCheckoutId())
	orderID, orderError := uuid.Parse(messageContext.GetOrderId())
	if checkoutError != nil || orderError != nil || checkoutID == uuid.Nil || orderID == uuid.Nil {
		return errors.New("payment checkout command has invalid checkout or order")
	}
	tx, err := service.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	queries := database.New(tx)
	if _, err := queries.InsertPaymentInboxMessage(ctx, messageContext.GetMessageId()); errors.Is(err, sql.ErrNoRows) {
		return tx.Commit()
	} else if err != nil {
		return err
	}
	var topic string
	var resultPayload any
	switch record.Topic {
	case messaging.EventTypeCheckoutCreateSession:
		if command.GetCreatePaymentSession() == nil {
			return errors.New("missing create payment session payload")
		}
		topic, resultPayload, err = createCheckoutSession(ctx, queries, checkoutID, orderID, messageContext, command.GetCreatePaymentSession())
	case messaging.EventTypeCheckoutCancelSession:
		if command.GetCancelPaymentSession() == nil {
			return errors.New("missing cancellation payload")
		}
		topic, resultPayload, err = cancelCheckoutSession(ctx, queries, checkoutID, orderID)
	default:
		return fmt.Errorf("unsupported checkout payment command topic %s", record.Topic)
	}
	if err != nil {
		return err
	}
	return emitCheckoutPaymentResult(ctx, tx, queries, checkoutID, orderID, messageContext, topic, resultPayload)
}

func createCheckoutSession(ctx context.Context, queries *database.Queries, checkoutID, orderID uuid.UUID, messageContext *checkoutv1.MessageContext, request *checkoutv1.CreatePaymentSession) (string, any, error) {
	buyerID, buyerError := uuid.Parse(request.GetBuyerUserId())
	merchantID, merchantError := uuid.Parse(request.GetMerchantId())
	if buyerError != nil || merchantError != nil || buyerID == uuid.Nil || merchantID == uuid.Nil ||
		request.GetTotalCents() <= 0 || request.GetCurrency() == "" || request.GetReturnUrl() == "" ||
		request.GetShippingAddress() == nil || len(request.GetItems()) == 0 || request.GetReservationOperationId() == "" {
		return "", nil, ErrInvalidSessionFacts
	}
	address := request.GetShippingAddress()
	shipping, err := json.Marshal(map[string]string{
		"name": address.GetName(), "line1": address.GetLine1(), "line2": address.GetLine2(),
		"city": address.GetCity(), "region": address.GetRegion(),
		"postal_code": address.GetPostalCode(), "country": address.GetCountry(),
	})
	if err != nil || !usableShippingAddress(shipping) {
		return "", nil, ErrInvalidSessionFacts
	}
	lines, err := json.Marshal(request.GetItems())
	if err != nil {
		return "", nil, err
	}
	buyer, err := partySnapshotJSON(buyerID.String(), request.GetBuyerEmail())
	if err != nil {
		return "", nil, err
	}
	merchant, err := partySnapshotJSON(merchantID.String(), "")
	if err != nil {
		return "", nil, err
	}
	encoded, err := (proto.MarshalOptions{Deterministic: true}).Marshal(request)
	if err != nil {
		return "", nil, err
	}
	hash := sha256.Sum256(encoded)
	existing, err := queries.GetCheckoutPaymentSession(ctx, orderID)
	if err == nil {
		if existing.CheckoutID != checkoutID || existing.RequestHash != nil && !bytes.Equal(existing.RequestHash, hash[:]) {
			return "", nil, ErrSessionMismatch
		}
		if existing.Status == "ABSENT" {
			return messaging.EventTypeCheckoutSessionRejected, &checkoutv1.PaymentSessionRejected{Reason: "session was cancelled before creation"}, nil
		}
		intent, err := queries.GetPaymentIntentByOrderIDForUpdate(ctx, orderID)
		if err != nil {
			return "", nil, err
		}
		if intent.Status != HostedPaymentSessionStatusPending {
			return messaging.EventTypeCheckoutSessionRejected, &checkoutv1.PaymentSessionRejected{Reason: "session is terminal"}, nil
		}
		return messaging.EventTypeCheckoutSessionReady, &checkoutv1.PaymentSessionReady{
			PaymentSessionId: intent.PaymentSessionID.String, ReturnUrl: intent.ReturnUrl,
			ExpiresAtUnixMs: intent.ExpiresAt.Time.UnixMilli(),
		}, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", nil, err
	}
	operationID, err := uuid.Parse(messageContext.GetOperationId())
	if err != nil {
		return "", nil, err
	}
	expiresAt := time.Now().UTC().Add(30 * time.Minute)
	sessionID := uuid.NewString()
	if _, err := queries.CreateHostedPaymentSession(ctx, database.CreateHostedPaymentSessionParams{
		OrderID: orderID, BuyerUserID: buyerID, Currency: request.GetCurrency(), ShippingAddress: shipping,
		Status: HostedPaymentSessionStatusPending, PaymentSessionID: sql.NullString{String: sessionID, Valid: true},
		ReturnUrl: request.GetReturnUrl(), ExpiresAt: sql.NullTime{Time: expiresAt, Valid: true},
		LineItems: lines, Buyer: buyer, Merchant: merchant,
	}); err != nil {
		return "", nil, err
	}
	if _, err := queries.CreatePaymentTransaction(ctx, database.CreatePaymentTransactionParams{
		ID: uuid.New(), OrderID: orderID, MerchantID: merchantID,
		AmountCents: request.GetTotalCents(), Currency: request.GetCurrency(),
		Status: PaymentTxStatusInitialized, IdempotencyKey: "order:" + orderID.String(),
	}); err != nil {
		return "", nil, err
	}
	if err := queries.InsertCheckoutPaymentSession(ctx, database.InsertCheckoutPaymentSessionParams{
		OrderID: orderID, CheckoutID: checkoutID, RequestHash: hash[:],
		CreateOperationID: uuid.NullUUID{UUID: operationID, Valid: true},
		CreateVersion:     sql.NullInt64{Int64: int64(messageContext.GetWorkflowVersion()), Valid: true},
		Status:            "ACTIVE",
	}); err != nil {
		return "", nil, err
	}
	return messaging.EventTypeCheckoutSessionReady, &checkoutv1.PaymentSessionReady{
		PaymentSessionId: sessionID, ReturnUrl: request.GetReturnUrl(), ExpiresAtUnixMs: expiresAt.UnixMilli(),
	}, nil
}

func cancelCheckoutSession(ctx context.Context, queries *database.Queries, checkoutID, orderID uuid.UUID) (string, any, error) {
	existing, err := queries.GetCheckoutPaymentSession(ctx, orderID)
	if errors.Is(err, sql.ErrNoRows) {
		if err := queries.InsertCheckoutPaymentSession(ctx, database.InsertCheckoutPaymentSessionParams{
			OrderID: orderID, CheckoutID: checkoutID, Status: "ABSENT",
		}); err != nil {
			return "", nil, err
		}
		return messaging.EventTypeCheckoutPaymentCancelled, &checkoutv1.PaymentCancelled{}, nil
	}
	if err != nil {
		return "", nil, err
	}
	if existing.CheckoutID != checkoutID {
		return "", nil, ErrSessionMismatch
	}
	intent, err := queries.GetPaymentIntentByOrderIDForUpdate(ctx, orderID)
	if err != nil {
		return "", nil, err
	}
	switch intent.Status {
	case HostedPaymentSessionStatusSucceeded:
		return messaging.EventTypeCheckoutPaymentSucceeded, &checkoutv1.PaymentSucceeded{PaymentSessionId: intent.PaymentSessionID.String}, nil
	case HostedPaymentSessionStatusFailed:
		return messaging.EventTypeCheckoutPaymentCancelled, &checkoutv1.PaymentCancelled{}, nil
	default:
		return messaging.EventTypeCheckoutPaymentUncertain, &checkoutv1.PaymentUncertain{
			PaymentSessionId: intent.PaymentSessionID.String, Reason: "gateway capture cannot be ruled out",
		}, nil
	}
}

func emitCheckoutPaymentResult(ctx context.Context, tx *sql.Tx, queries *database.Queries, checkoutID, orderID uuid.UUID, original *checkoutv1.MessageContext, topic string, resultPayload any) error {
	resultID := uuid.New()
	result := &checkoutv1.CheckoutMessage{Context: &checkoutv1.MessageContext{
		MessageId: resultID.String(), CheckoutId: checkoutID.String(), OrderId: orderID.String(),
		OperationId: original.GetOperationId(), CausationId: original.GetMessageId(),
		WorkflowVersion: original.GetWorkflowVersion(),
	}}
	switch payload := resultPayload.(type) {
	case *checkoutv1.PaymentSessionReady:
		result.Payload = &checkoutv1.CheckoutMessage_PaymentSessionReady{PaymentSessionReady: payload}
	case *checkoutv1.PaymentSessionRejected:
		result.Payload = &checkoutv1.CheckoutMessage_PaymentSessionRejected{PaymentSessionRejected: payload}
	case *checkoutv1.PaymentCancelled:
		result.Payload = &checkoutv1.CheckoutMessage_PaymentCancelled{PaymentCancelled: payload}
	case *checkoutv1.PaymentUncertain:
		result.Payload = &checkoutv1.CheckoutMessage_PaymentUncertain{PaymentUncertain: payload}
	case *checkoutv1.PaymentSucceeded:
		result.Payload = &checkoutv1.CheckoutMessage_PaymentSucceeded{PaymentSucceeded: payload}
	case *checkoutv1.PaymentFailed:
		result.Payload = &checkoutv1.CheckoutMessage_PaymentFailed{PaymentFailed: payload}
	default:
		return errors.New("invalid checkout payment result")
	}
	encoded, err := proto.Marshal(result)
	if err != nil {
		return err
	}
	if _, err := queries.CreatePaymentOutbox(ctx, database.CreatePaymentOutboxParams{
		ID: resultID, AggregateID: checkoutID, EventType: topic,
		Payload: encoded, Tracingspancontext: sharedtrace.SerializeContext(ctx),
	}); err != nil {
		return err
	}
	return tx.Commit()
}
