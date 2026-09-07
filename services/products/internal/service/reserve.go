package service

import (
	"context"
	"errors"

	sharedlog "github.com/phuchoang2603/refurbished-marketplace/shared/observe/log"

	"github.com/google/uuid"
)

func (s *Service) ReserveStock(ctx context.Context, orderID, merchantID uuid.UUID, totalCents int64, items []ReservationItemInput) error {
	if orderID == uuid.Nil || merchantID == uuid.Nil {
		return ErrInvalidMerchantID
	}
	if totalCents <= 0 || len(items) == 0 {
		return ErrInvalidQuantity
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	q := s.queries.WithTx(tx)
	defer func() {
		_ = tx.Rollback()
	}()

	existing, err := q.CountInventoryReservationsByOrderID(ctx, orderID)
	if err != nil {
		return err
	}
	if existing > 0 {
		return tx.Commit()
	}

	if err := reserveOrderItems(ctx, q, orderID, items); err != nil {
		if errors.Is(err, ErrInventoryNotFound) || errors.Is(err, ErrInsufficientStock) {
			if outboxErr := createInventoryReservationFailedOutbox(ctx, q, orderID); outboxErr != nil {
				return outboxErr
			}
			if err := tx.Commit(); err != nil {
				return err
			}
			sharedlog.WarnContext(
				ctx, "inventory reservation failed",
				sharedlog.KeyOrderID, orderID.String(),
				sharedlog.KeyMerchantID, merchantID.String(),
				sharedlog.KeyErr, err,
			)
			return err
		}
		return err
	}

	if err := createInventoryReservedOutbox(ctx, q, orderID, merchantID, totalCents); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	sharedlog.InfoContext(
		ctx, "inventory reserved",
		sharedlog.KeyOrderID, orderID.String(),
		sharedlog.KeyMerchantID, merchantID.String(),
		"total_cents", totalCents,
		"item_count", len(items),
	)
	return nil
}
