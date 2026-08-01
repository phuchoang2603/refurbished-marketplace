package service

import (
	"context"
	"database/sql"
	"errors"
	"time"

	sharedlog "refurbished-marketplace/shared/observe/log"

	"github.com/google/uuid"
)

const sessionExpiryBatchSize int32 = 100

// ExpireDueSessions marks PENDING hosted sessions past expires_at as EXPIRED
// and emits payment.failed when a payment transaction already exists.
func (s *Service) ExpireDueSessions(ctx context.Context) error {
	due, err := s.queries.ListExpiredPendingHostedSessions(ctx, sessionExpiryBatchSize)
	if err != nil {
		return err
	}
	for _, intent := range due {
		if err := s.expireOneSession(ctx, intent.OrderID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) expireOneSession(ctx context.Context, orderID uuid.UUID) error {
	expired, err := s.queries.ExpireHostedPaymentSession(ctx, orderID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}

	txRow, err := s.queries.GetPaymentTransactionByOrderID(ctx, orderID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			sharedlog.InfoContext(
				ctx, "hosted payment session expired; awaiting payment transaction",
				sharedlog.KeyOrderID, orderID.String(),
				sharedlog.KeyStatus, expired.Status,
			)
			return nil
		}
		return err
	}

	return s.applyTerminalOutcome(ctx, txRow.ID, expired.Status, expired.FailureReason)
}

// RunSessionExpiryLoop periodically expires due hosted payment sessions until ctx is cancelled.
func (s *Service) RunSessionExpiryLoop(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		return
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	run := func() {
		if err := s.ExpireDueSessions(ctx); err != nil && !errors.Is(err, context.Canceled) {
			sharedlog.ErrorContext(ctx, "expire due payment sessions", "err", err)
		}
	}
	run()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			run()
		}
	}
}
