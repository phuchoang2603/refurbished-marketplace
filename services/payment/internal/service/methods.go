package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"refurbished-marketplace/services/payment/internal/database"
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

type HostedPaymentSessionView struct {
	OrderID          string
	PaymentSessionID string
	Currency         string
	Status           string
	ReturnURL        string
	CancelURL        string
	FailureReason    string
	ExpiresAt        time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type CreateHostedPaymentSessionParams struct {
	OrderID         uuid.UUID
	BuyerUserID     uuid.UUID
	Currency        string
	ShippingAddress json.RawMessage
	ReturnURL       string
	CancelURL       string
}

func (s *Service) CreateHostedPaymentSession(ctx context.Context, p CreateHostedPaymentSessionParams) (HostedPaymentSessionView, error) {
	p.Currency = defaultPaymentCurrency(p.Currency)

	intent, err := loadPaymentIntentByOrderID(ctx, s.queries, p.OrderID)
	if err == nil {
		return mapDBHostedPaymentSessionView(intent), nil
	}
	if !errors.Is(err, ErrIntentNotFound) {
		return HostedPaymentSessionView{}, err
	}

	expiresAt := time.Now().UTC().Add(30 * time.Minute)
	created, err := s.queries.CreateHostedPaymentSession(ctx, database.CreateHostedPaymentSessionParams{
		OrderID:          p.OrderID,
		BuyerUserID:      p.BuyerUserID,
		Currency:         p.Currency,
		ShippingAddress:  p.ShippingAddress,
		Status:           HostedPaymentSessionStatusPending,
		PaymentSessionID: optionalNullString(uuid.NewString()),
		ReturnUrl:        p.ReturnURL,
		CancelUrl:        p.CancelURL,
		ExpiresAt:        optionalNullTime(expiresAt),
		FailureReason:    sql.NullString{},
	})
	if err != nil {
		return HostedPaymentSessionView{}, err
	}

	return mapDBHostedPaymentSessionView(created), nil
}

func (s *Service) ApplyGatewayWebhook(ctx context.Context, orderID uuid.UUID, paymentSessionID, status, failureReason string) error {
	intent, err := loadPaymentIntentByOrderID(ctx, s.queries, orderID)
	if err != nil {
		return err
	}
	if !intent.PaymentSessionID.Valid || intent.PaymentSessionID.String != paymentSessionID {
		return ErrSessionMismatch
	}
	if hostedPaymentSessionIsTerminal(intent.Status) {
		return nil
	}

	updatedIntent, err := s.queries.UpdateHostedPaymentSessionOutcome(ctx, database.UpdateHostedPaymentSessionOutcomeParams{
		OrderID:          orderID,
		PaymentSessionID: optionalNullString(paymentSessionID),
		Status:           status,
		FailureReason:    optionalNullString(failureReason),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrSessionMismatch
		}
		return err
	}

	txRow, err := s.queries.GetPaymentTransactionByOrderID(ctx, orderID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}

	return s.applyTerminalOutcome(ctx, txRow.ID, updatedIntent.Status, updatedIntent.FailureReason)
}

func (s *Service) applyTerminalOutcome(ctx context.Context, transactionID uuid.UUID, hostedStatus string, failureReason sql.NullString) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	q := database.New(tx)
	defer func() {
		_ = tx.Rollback()
	}()

	newStatus := PaymentTxStatusFailed
	if hostedPaymentSessionMapsToSuccess(hostedStatus) {
		newStatus = PaymentTxStatusSucceeded
	}

	updated, err := q.UpdatePaymentTransactionGatewayResult(ctx, database.UpdatePaymentTransactionGatewayResultParams{
		ID:                   transactionID,
		Status:               newStatus,
		GatewayTransactionID: sql.NullString{},
		FailureReason:        failureReason,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
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
	if newStatus == PaymentTxStatusSucceeded {
		eventType = messaging.EventTypePaymentSucceeded
	}
	payload, err := proto.Marshal(&paymentv1.PaymentOutcome{OrderId: updated.OrderID.String()})
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

func (s *Service) GetHostedPaymentSessionByOrder(ctx context.Context, orderID uuid.UUID) (HostedPaymentSessionView, error) {
	row, err := loadPaymentIntentByOrderID(ctx, s.queries, orderID)
	if err != nil {
		return HostedPaymentSessionView{}, err
	}
	return mapDBHostedPaymentSessionView(row), nil
}

func (s *Service) GetPaymentTransaction(ctx context.Context, id uuid.UUID) (PaymentTransactionView, error) {
	row, err := loadPaymentTransaction(ctx, s.queries, id)
	if err != nil {
		return PaymentTransactionView{}, err
	}
	return mapDBPaymentTransactionView(row), nil
}
