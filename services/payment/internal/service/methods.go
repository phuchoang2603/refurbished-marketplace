package service

import (
	"context"
	"encoding/json"
	"time"

	"refurbished-marketplace/services/payment/internal/database"
	"refurbished-marketplace/shared/dberrors"
	"refurbished-marketplace/shared/messaging"

	paymentv1 "refurbished-marketplace/shared/proto/payment/v1"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

type PaymentTransactionView struct {
	ID                   string
	OrderID              string
	MerchantID           string
	AmountCents          int64
	Currency             string
	Status               string
	IdempotencyKey       string
	GatewayTransactionID string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type InitiatePaymentParams struct {
	OrderID         uuid.UUID
	BuyerUserID     uuid.UUID
	PaymentToken    string
	Currency        string
	BillingAddress  json.RawMessage
	ShippingAddress json.RawMessage
}

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
	q := database.New(tx)
	defer func() {
		_ = tx.Rollback()
	}()

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

	eventType := messaging.EventTypePaymentFailed
	if succeeded {
		eventType = messaging.EventTypePaymentSucceeded
	}
	payload, err := proto.Marshal(&paymentv1.PaymentOutcome{
		OrderId: updated.OrderID.String(),
	})
	if err != nil {
		return err
	}

	if _, err := q.CreatePaymentOutbox(ctx, database.CreatePaymentOutboxParams{
		ID:          uuid.New(),
		AggregateID: updated.OrderID,
		EventType:   eventType,
		Payload:     payload,
	}); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Service) GetPaymentTransaction(ctx context.Context, id uuid.UUID) (PaymentTransactionView, error) {
	row, err := loadPaymentTransaction(ctx, s.queries, id)
	if err != nil {
		return PaymentTransactionView{}, err
	}
	return mapDBPaymentTransactionView(row), nil
}
