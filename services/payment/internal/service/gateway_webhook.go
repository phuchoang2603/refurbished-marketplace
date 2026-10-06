package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

func (s *Service) ApplyGatewayWebhook(ctx context.Context, orderID uuid.UUID, paymentSessionID, status, failureReason string) error {
	checkoutSession, err := s.queries.GetCheckoutPaymentSession(ctx, orderID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrIntentNotFound
	}
	if err != nil {
		return err
	}
	return s.applyCheckoutGatewayWebhook(ctx, checkoutSession, paymentSessionID, status, failureReason)
}
