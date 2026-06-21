package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"refurbished-marketplace/services/payment/internal/database"
	shareddb "refurbished-marketplace/shared/db"

	"github.com/google/uuid"
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
		PaymentSessionID: shareddb.OptionalNullString(uuid.NewString()),
		ReturnUrl:        p.ReturnURL,
		CancelUrl:        p.CancelURL,
		ExpiresAt:        shareddb.OptionalNullTime(expiresAt),
	})
	if err != nil {
		return HostedPaymentSessionView{}, err
	}

	return mapDBHostedPaymentSessionView(created), nil
}

func (s *Service) GetHostedPaymentSessionByOrder(ctx context.Context, orderID uuid.UUID) (HostedPaymentSessionView, error) {
	row, err := loadPaymentIntentByOrderID(ctx, s.queries, orderID)
	if err != nil {
		return HostedPaymentSessionView{}, err
	}
	return mapDBHostedPaymentSessionView(row), nil
}
