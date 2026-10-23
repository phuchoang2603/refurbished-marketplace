package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/phuchoang2603/refurbished-marketplace/services/checkout/internal/database"
	"github.com/phuchoang2603/refurbished-marketplace/shared/messaging"
	sharedlog "github.com/phuchoang2603/refurbished-marketplace/shared/observe/log"
	sharedtrace "github.com/phuchoang2603/refurbished-marketplace/shared/observe/trace"
	checkoutv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/checkout/v1"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type gatewayEvent struct {
	EventID        string    `json:"event_id"`
	Type           string    `json:"type"`
	PaymentID      string    `json:"payment_id"`
	OrderReference string    `json:"order_reference"`
	AttemptID      string    `json:"attempt_id"`
	Status         string    `json:"status"`
	Reason         string    `json:"reason"`
	OccurredAt     time.Time `json:"occurred_at"`
}

func verifyGatewaySignature(secret, header string, body []byte, now time.Time) bool {
	if secret == "" {
		return false
	}
	var timestamp, signature string
	for _, part := range strings.Split(header, ",") {
		name, value, ok := strings.Cut(strings.TrimSpace(part), "=")
		if !ok {
			return false
		}
		switch name {
		case "t":
			if timestamp != "" {
				return false
			}
			timestamp = value
		case "v1":
			if signature != "" {
				return false
			}
			signature = value
		default:
			return false
		}
	}
	seconds, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return false
	}
	signedAt := time.Unix(seconds, 0)
	if signedAt.Before(now.Add(-5*time.Minute)) || signedAt.After(now.Add(5*time.Minute)) {
		return false
	}
	received, err := hex.DecodeString(signature)
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(timestamp + "."))
	_, _ = mac.Write(body)
	return hmac.Equal(received, mac.Sum(nil))
}

func (service *Service) HandleGatewayWebhook(ctx context.Context, request *checkoutv1.HandleGatewayWebhookRequest) (*checkoutv1.HandleGatewayWebhookResponse, error) {
	rejected := &checkoutv1.HandleGatewayWebhookResponse{Result: checkoutv1.GatewayWebhookResult_GATEWAY_WEBHOOK_RESULT_REJECTED}
	if request == nil || len(request.GetRawBody()) > 64<<10 || !verifyGatewaySignature(service.config.PaymentGatewayWebhookSecret, request.GetSignatureHeader(), request.GetRawBody(), time.Now()) {
		return rejected, nil
	}
	var event gatewayEvent
	if json.Unmarshal(request.GetRawBody(), &event) != nil || event.EventID == "" || len(event.EventID) > 255 || event.PaymentID == "" || event.OrderReference == "" || event.OccurredAt.IsZero() {
		return rejected, nil
	}
	switch event.Type {
	case "payment.succeeded":
		if event.Status != "SUCCEEDED" {
			return rejected, nil
		}
	case "payment.expired":
		if event.Status != "EXPIRED" {
			return rejected, nil
		}
	case "attempt.declined":
		if event.Status != "DECLINED" {
			return rejected, nil
		}
	default:
		return rejected, nil
	}
	ctx, span := sharedtrace.Tracer("checkout").Start(ctx, "gateway.webhook")
	defer span.End()
	tx, err := service.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, status.Error(codes.Unavailable, "webhook store unavailable")
	}
	defer func() { _ = tx.Rollback() }()
	queries := database.New(tx)
	row, err := queries.LockCheckoutByGatewayPayment(ctx, sql.NullString{String: event.PaymentID, Valid: true})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, status.Error(codes.Unavailable, "gateway payment not recorded yet")
	}
	if err != nil {
		return nil, status.Error(codes.Unavailable, "webhook store unavailable")
	}
	if event.OrderReference != row.OrderID.UUID.String() {
		return rejected, nil
	}
	span.SetAttributes(attribute.String("order_id", row.OrderID.UUID.String()), attribute.String("gateway_payment_id", event.PaymentID))
	inserted, err := queries.InsertInbox(ctx, database.InsertInboxParams{MessageID: "gateway:" + event.EventID, CheckoutID: row.ID})
	if err != nil {
		return nil, status.Error(codes.Unavailable, "webhook store unavailable")
	}
	result := checkoutv1.GatewayWebhookResult_GATEWAY_WEBHOOK_RESULT_DUPLICATE
	if inserted != 0 {
		result = checkoutv1.GatewayWebhookResult_GATEWAY_WEBHOOK_RESULT_ACCEPTED
		err = service.applyGatewayOutcome(ctx, queries, row, event.Type, "gateway:"+event.EventID, event.OccurredAt)
	}
	if err == nil {
		err = tx.Commit()
	}
	if err != nil {
		return nil, status.Error(codes.Unavailable, "webhook store unavailable")
	}
	return &checkoutv1.HandleGatewayWebhookResponse{Result: result}, nil
}

func (service *Service) applyGatewayOutcome(ctx context.Context, queries *database.Queries, row database.Checkout, eventType, eventID string, occurredAt time.Time) error {
	if gatewayOutcomeContradicts(row, eventType) {
		return recordGatewayContradiction(ctx, queries, row, eventType, eventID)
	}
	var paymentStatus sql.NullString
	switch eventType {
	case "payment.succeeded":
		paymentStatus = sql.NullString{String: "SUCCEEDED", Valid: true}
	case "payment.expired":
		paymentStatus = sql.NullString{String: "EXPIRED", Valid: true}
	}
	if err := queries.RecordGatewayEvent(ctx, database.RecordGatewayEventParams{ID: row.ID, Column2: occurredAt, PaymentStatus: paymentStatus}); err != nil {
		return err
	}
	sharedlog.InfoContext(ctx, "gateway outcome accepted", "order_id", row.OrderID.UUID.String(), "gateway_payment_id", row.GatewayPaymentID.String, "event_type", eventType)
	settlement := gatewaySettlement(row, eventType)
	if settlement == nil {
		return nil
	}
	if err := queries.ResolvePaymentUncertain(ctx, row.ID); err != nil {
		return err
	}
	return deadlineCommand(ctx, queries, row, settlement.state, settlement.commandTopic, settlement.operation, settlement.commandPayload, time.Now().Add(preparationDeadline))
}

func gatewayOutcomeContradicts(row database.Checkout, eventType string) bool {
	state := checkoutv1.CheckoutState(checkoutv1.CheckoutState_value[row.State])
	switch eventType {
	case "payment.succeeded":
		if row.PaymentStatus.String == "EXPIRED" {
			return true
		}
		switch state {
		case checkoutv1.CheckoutState_CHECKOUT_STATE_RELEASING_STOCK,
			checkoutv1.CheckoutState_CHECKOUT_STATE_FINALIZING_FAILED,
			checkoutv1.CheckoutState_CHECKOUT_STATE_FAILED:
			return true
		}
	case "payment.expired":
		if row.PaymentStatus.String == "SUCCEEDED" {
			return true
		}
		switch state {
		case checkoutv1.CheckoutState_CHECKOUT_STATE_COMMITTING_STOCK,
			checkoutv1.CheckoutState_CHECKOUT_STATE_FINALIZING_PAID,
			checkoutv1.CheckoutState_CHECKOUT_STATE_COMPLETED:
			return true
		}
	}
	return false
}

func recordGatewayContradiction(ctx context.Context, queries *database.Queries, row database.Checkout, eventType, eventID string) error {
	if err := queries.InsertException(ctx, database.InsertExceptionParams{ID: uuid.New(), CheckoutID: row.ID, MessageID: sql.NullString{String: eventID, Valid: true}, Kind: "contradictory_payment", Details: "gateway outcome contradicts stock settlement"}); err != nil {
		return err
	}
	sharedlog.WarnContext(ctx, "contradictory gateway outcome", "order_id", row.OrderID.UUID.String(), "gateway_payment_id", row.GatewayPaymentID.String, "event_type", eventType)
	state := checkoutv1.CheckoutState(checkoutv1.CheckoutState_value[row.State])
	if state == checkoutv1.CheckoutState_CHECKOUT_STATE_FAILED || state == checkoutv1.CheckoutState_CHECKOUT_STATE_COMPLETED {
		return nil
	}
	_, err := queries.AdvanceCheckout(ctx, database.AdvanceCheckoutParams{ID: row.ID, Version: row.Version, State: checkoutv1.CheckoutState_CHECKOUT_STATE_NEEDS_REVIEW.String()})
	return err
}

func gatewaySettlement(row database.Checkout, eventType string) *transition {
	state := checkoutv1.CheckoutState(checkoutv1.CheckoutState_value[row.State])
	if state != checkoutv1.CheckoutState_CHECKOUT_STATE_READY_TO_PAY && (state != checkoutv1.CheckoutState_CHECKOUT_STATE_NEEDS_REVIEW || row.PaymentStatus.String != "OPEN") {
		return nil
	}
	switch eventType {
	case "payment.succeeded":
		return &transition{
			state:        checkoutv1.CheckoutState_CHECKOUT_STATE_COMMITTING_STOCK,
			commandTopic: messaging.EventTypeCheckoutCommitStock,
			operation:    "commit-stock", commandPayload: &checkoutv1.CommitStock{},
		}
	case "payment.expired":
		return &transition{
			state:        checkoutv1.CheckoutState_CHECKOUT_STATE_RELEASING_STOCK,
			commandTopic: messaging.EventTypeCheckoutCancelStock,
			operation:    "cancel-stock", commandPayload: &checkoutv1.CancelStock{Reason: "payment expired"},
		}
	default:
		return nil
	}
}
