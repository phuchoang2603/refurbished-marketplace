package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"refurbished-marketplace/services/payment/internal/database"
	"refurbished-marketplace/shared/messaging"

	productsv1 "refurbished-marketplace/shared/proto/products/v1"

	"github.com/google/uuid"
)

func (s *Service) KafkaInventoryReservedHandler() messaging.KafkaHandler {
	return func(ctx context.Context, msg messaging.KafkaMessage) error {
		messageID := messaging.KafkaMessageID(msg)
		if messageID == "" {
			return errors.New("messageID is required")
		}

		var payload productsv1.InventoryReserved
		if err := messaging.UnmarshalKafkaProtobuf(msg.Value, &payload); err != nil {
			return fmt.Errorf("decode inventory.reserved payload: %w", err)
		}
		if payload.GetOrderId() == "" {
			return errors.New("invalid inventory.reserved payload: missing order_id")
		}

		if _, err := s.queries.InsertPaymentInboxMessage(ctx, messageID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil
			}
			return err
		}

		orderID, merchantID, err := parseOrderUUIDs(&payload)
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
			AmountCents:    payload.GetTotalCents(),
			Currency:       intent.Currency,
			Status:         PaymentTxStatusInitialized,
			IdempotencyKey: "order:" + orderID.String(),
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) || isPostgresUniqueViolation(err) {
				return nil
			}
			return err
		}
		return nil
	}
}
