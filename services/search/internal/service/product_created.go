package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/phuchoang2603/refurbished-marketplace/shared/messaging"
	sharedlog "github.com/phuchoang2603/refurbished-marketplace/shared/observe/log"
	productsv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/products/v1"
)

func (s *Service) KafkaProductCreatedHandler() messaging.KafkaHandler {
	return func(ctx context.Context, msg messaging.KafkaMessage) error {
		if msg.Topic != messaging.EventTypeProductCreated {
			return nil
		}
		return s.HandleProductCreated(ctx, msg.Value)
	}
}

func (s *Service) HandleProductCreated(ctx context.Context, value []byte) error {
	var msg productsv1.ProductCreated
	if err := messaging.UnmarshalKafkaProtobuf(value, &msg); err != nil {
		return fmt.Errorf("decode products.created payload: %w", err)
	}

	productID, err := uuid.Parse(msg.GetProductId())
	if err != nil || productID == uuid.Nil {
		sharedlog.Error("skip product created: invalid product id", "product_id", msg.GetProductId())
		return nil
	}
	merchantID, err := uuid.Parse(msg.GetMerchantId())
	if err != nil || merchantID == uuid.Nil {
		sharedlog.Error("skip product created: invalid merchant id", "merchant_id", msg.GetMerchantId())
		return nil
	}
	name := strings.TrimSpace(msg.GetName())
	if name == "" || msg.GetPriceCents() <= 0 {
		sharedlog.Error("skip product created: invalid catalog fields", "product_id", productID.String())
		return nil
	}

	createdAt := time.Now().UTC()
	if msg.GetOccurredAt() != nil && msg.GetOccurredAt().CheckValid() == nil {
		createdAt = msg.GetOccurredAt().AsTime().UTC()
	}

	return s.UpsertListing(ctx, listingDocument{
		ID:          productID.String(),
		Name:        name,
		Description: strings.TrimSpace(msg.GetDescription()),
		PriceCents:  msg.GetPriceCents(),
		MerchantID:  merchantID.String(),
		CreatedAt:   createdAt.Unix(),
	})
}
