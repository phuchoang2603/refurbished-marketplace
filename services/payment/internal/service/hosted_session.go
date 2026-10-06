package service

import (
	"context"
	"time"

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
	FailureReason    string
	ExpiresAt        time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (s *Service) GetHostedPaymentSessionByOrder(ctx context.Context, orderID uuid.UUID) (HostedPaymentSessionView, error) {
	row, err := loadPaymentIntentByOrderID(ctx, s.queries, orderID)
	if err != nil {
		return HostedPaymentSessionView{}, err
	}
	return mapDBHostedPaymentSessionView(row), nil
}
