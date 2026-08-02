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
// and emits payment.failed when a payment transaction already exists. It also
// repairs EXPIRED sessions that still have a non-terminal payment transaction.
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

	repair, err := s.queries.ListExpiredHostedSessionsNeedingTerminalApply(ctx, sessionExpiryBatchSize)
	if err != nil {
		return err
	}
	for _, intent := range repair {
		if err := s.applyExpiredSessionTerminalOutcome(ctx, intent.OrderID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) expireOneSession(ctx context.Context, orderID uuid.UUID) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()
	q := s.queries.WithTx(tx)

	expired, err := q.ExpireHostedPaymentSession(ctx, orderID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}

	txRow, err := q.GetPaymentTransactionByOrderID(ctx, orderID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			sharedlog.InfoContext(
				ctx, "hosted payment session expired; awaiting payment transaction",
				sharedlog.KeyOrderID, orderID.String(),
				sharedlog.KeyStatus, expired.Status,
			)
			return tx.Commit()
		}
		return err
	}

	if err := s.applyTerminalOutcomeWithQueries(ctx, q, txRow.ID, expired.Status, expired.FailureReason); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Service) applyExpiredSessionTerminalOutcome(ctx context.Context, orderID uuid.UUID) error {
	intent, err := loadPaymentIntentByOrderID(ctx, s.queries, orderID)
	if err != nil {
		return err
	}
	if intent.Status != HostedPaymentSessionStatusExpired {
		return nil
	}

	txRow, err := s.queries.GetPaymentTransactionByOrderID(ctx, orderID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}
	if paymentTransactionIsTerminal(txRow.Status) {
		return nil
	}

	return s.applyTerminalOutcome(ctx, txRow.ID, intent.Status, intent.FailureReason)
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
