package service

import (
	"context"
	"errors"
	"fmt"

	"refurbished-marketplace/services/payment/internal/database"
	"refurbished-marketplace/shared/dberrors"
	"refurbished-marketplace/shared/messaging"

	ordersv1 "refurbished-marketplace/shared/proto/orders/v1"

	"github.com/google/uuid"
)

func (s *Service) KafkaOrdersCreatedHandler() messaging.KafkaHandler {
	return func(ctx context.Context, msg messaging.KafkaMessage) error {
		return s.HandleOrdersCreated(ctx, messaging.KafkaMessageID(msg), msg.Value)
	}
}

func (s *Service) HandleOrdersCreated(ctx context.Context, messageID string, value []byte) error {
	if messageID == "" {
		return errors.New("messageID is required")
	}

	var msg ordersv1.OrderCreated
	if err := messaging.UnmarshalKafkaProtobuf(value, &msg); err != nil {
		return fmt.Errorf("decode orders.created payload: %w", err)
	}
	if msg.GetOrderId() == "" {
		return errors.New("invalid orders.created payload: missing order_id")
	}

	if _, err := s.queries.InsertPaymentInboxMessage(ctx, messageID); err != nil {
		if dberrors.IsNoRows(err) {
			return nil
		}
		return err
	}

	orderID, merchantID, err := parseOrderCreatedUUIDs(&msg)
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
		MerchantID:     merchantID,
		AmountCents:    msg.GetTotalCents(),
		Currency:       intent.Currency,
		Status:         PaymentTxStatusInitialized,
		IdempotencyKey: "order:" + orderID.String(),
	})
	if err != nil {
		if dberrors.IsNoRows(err) || isPostgresUniqueViolation(err) {
			return nil
		}
		return err
	}
	return nil
}
