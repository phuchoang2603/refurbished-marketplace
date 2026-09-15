package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/phuchoang2603/refurbished-marketplace/services/inventory/internal/database"
	"github.com/phuchoang2603/refurbished-marketplace/shared/messaging"
	productsv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/products/v1"
)

var ErrConflictingSeed = errors.New("conflicting inventory seed intent")

func (s *Service) HandleProductCreated(ctx context.Context, value []byte) error {
	var msg productsv1.ProductCreated
	if err := messaging.UnmarshalKafkaProtobuf(value, &msg); err != nil {
		return err
	}
	eventID, err := uuid.Parse(msg.GetEventId())
	if err != nil || eventID == uuid.Nil {
		return fmt.Errorf("invalid creation event id")
	}
	productID, err := uuid.Parse(msg.GetProductId())
	if err != nil || productID == uuid.Nil {
		return ErrInvalidProductID
	}
	merchantID, err := uuid.Parse(msg.GetMerchantId())
	if err != nil || merchantID == uuid.Nil {
		return ErrInvalidMerchantID
	}
	if msg.GetSchemaVersion() != 1 || msg.GetProductVersion() != 1 || msg.GetOccurredAt() == nil || msg.GetOccurredAt().CheckValid() != nil || strings.TrimSpace(msg.GetName()) == "" || msg.GetPriceCents() <= 0 {
		return fmt.Errorf("invalid ProductCreated metadata")
	}
	if msg.InitialQty == nil || msg.GetInitialQty() < 0 {
		return ErrInvalidQuantity
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	q := s.queries.WithTx(tx)
	if _, err := q.InsertInventoryInboxMessage(ctx, messaging.EventTypeProductCreated+"/"+eventID.String()); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return tx.Commit()
		}
		return err
	}
	if err := q.InsertInventorySeedIntent(ctx, database.InsertInventorySeedIntentParams{
		ProductID: productID, InitialQty: msg.GetInitialQty(), ProductVersion: int64(msg.GetProductVersion()),
	}); err != nil {
		return err
	}
	intent, err := q.GetInventorySeedIntent(ctx, productID)
	if err != nil {
		return err
	}
	if intent.InitialQty != msg.GetInitialQty() || intent.ProductVersion != int64(msg.GetProductVersion()) {
		return ErrConflictingSeed
	}
	if _, err := q.GetInventoryByProductID(ctx, productID); errors.Is(err, sql.ErrNoRows) {
		if _, err := q.CreateInventory(ctx, database.CreateInventoryParams{ProductID: productID, AvailableQty: msg.GetInitialQty()}); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return tx.Commit()
}
