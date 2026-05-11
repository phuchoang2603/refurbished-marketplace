package service

import (
	"context"
	"errors"
	"fmt"

	"refurbished-marketplace/services/payment/internal/database"
	"refurbished-marketplace/shared/dberrors"
	"refurbished-marketplace/shared/messaging"
	ordersv1 "refurbished-marketplace/shared/proto/orders/v1"
	paymentv1 "refurbished-marketplace/shared/proto/payment/v1"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

// InitiatePayment upserts a payment intent for an order (web confirmation).
func (s *Service) InitiatePayment(ctx context.Context, p InitiatePaymentParams) error {
	p.Currency = defaultPaymentCurrency(p.Currency)
	_, err := s.queries.UpsertPaymentIntent(ctx, database.UpsertPaymentIntentParams{
		OrderID:         p.OrderID,
		BuyerUserID:     p.BuyerUserID,
		PaymentToken:    p.PaymentToken,
		Currency:        p.Currency,
		BillingAddress:  p.BillingAddress,
		ShippingAddress: p.ShippingAddress,
		Status:          PaymentTxStatusInitialized,
	})
	return err
}

// ApplyGatewayWebhook updates the transaction and appends payment_outbox in one DB transaction.
func (s *Service) ApplyGatewayWebhook(ctx context.Context, transactionID uuid.UUID, gatewayTransactionID string, succeeded bool, failureReason string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	q := database.New(tx)

	newStatus := PaymentTxStatusFailed
	if succeeded {
		newStatus = PaymentTxStatusSucceeded
	}

	updated, err := q.UpdatePaymentTransactionGatewayResult(ctx, database.UpdatePaymentTransactionGatewayResultParams{
		ID:                   transactionID,
		Status:               newStatus,
		GatewayTransactionID: optionalNullString(gatewayTransactionID),
		FailureReason:        optionalNullString(failureReason),
	})
	if err != nil {
		if dberrors.IsNoRows(err) {
			row, loadErr := loadPaymentTransaction(ctx, q, transactionID)
			if loadErr != nil {
				return loadErr
			}
			if paymentTransactionIsTerminal(row.Status) {
				return tx.Commit()
			}
		}
		return err
	}

	eventType := messaging.EventTypePaymentItemFailed
	if succeeded {
		eventType = messaging.EventTypePaymentItemSucceeded
	}
	payload, err := proto.Marshal(&paymentv1.PaymentItemOutbox{
		OrderId:     updated.OrderID.String(),
		OrderItemId: updated.OrderItemID.String(),
	})
	if err != nil {
		return err
	}

	if _, err := q.CreatePaymentOutbox(ctx, database.CreatePaymentOutboxParams{
		ID:          uuid.New(),
		AggregateID: updated.OrderItemID,
		EventType:   eventType,
		Payload:     payload,
	}); err != nil {
		return err
	}

	return tx.Commit()
}

// GetPaymentTransaction loads a transaction by primary key.
func (s *Service) GetPaymentTransaction(ctx context.Context, id uuid.UUID) (PaymentTransactionView, error) {
	row, err := loadPaymentTransaction(ctx, s.queries, id)
	if err != nil {
		return PaymentTransactionView{}, err
	}
	return paymentTransactionViewFromDB(row), nil
}

// HandleOrdersItemCreated records inbox dedupe and creates a per-item payment transaction when intent exists.
func (s *Service) HandleOrdersItemCreated(ctx context.Context, messageID string, value []byte) error {
	if messageID == "" {
		return errors.New("messageID is required")
	}

	var msg ordersv1.OrderItemCreated
	if err := messaging.UnmarshalKafkaProtobuf(value, &msg); err != nil {
		return fmt.Errorf("decode orders.item.created payload: %w", err)
	}
	if msg.GetOrderItemId() == "" || msg.GetOrderId() == "" {
		return errors.New("invalid orders.item.created payload: missing order_id or order_item_id")
	}

	if _, err := s.queries.InsertPaymentInboxMessage(ctx, messageID); err != nil {
		if dberrors.IsNoRows(err) {
			return nil
		}
		return err
	}

	orderID, orderItemID, merchantID, err := parseOrderItemCreatedUUIDs(&msg)
	if err != nil {
		return err
	}

	intent, err := loadPaymentIntentByOrderID(ctx, s.queries, orderID)
	if err != nil {
		return err
	}

	_, err = s.queries.CreatePaymentTransaction(ctx, database.CreatePaymentTransactionParams{
		ID:             uuid.New(),
		OrderID:        orderID,
		OrderItemID:    orderItemID,
		MerchantID:     merchantID,
		AmountCents:    msg.GetLineTotalCents(),
		Currency:       intent.Currency,
		Status:         PaymentTxStatusInitialized,
		IdempotencyKey: "order_item:" + orderItemID.String(),
	})
	if err != nil {
		if dberrors.IsNoRows(err) || isPostgresUniqueViolation(err) {
			return nil
		}
		return err
	}
	return nil
}
