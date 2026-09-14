package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/phuchoang2603/refurbished-marketplace/services/products/internal/catalog"
	"github.com/phuchoang2603/refurbished-marketplace/shared/messaging"
	sharedtrace "github.com/phuchoang2603/refurbished-marketplace/shared/observe/trace"
	productsv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/products/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Service) CreateProduct(ctx context.Context, name, description string, priceCents int64, merchantID uuid.UUID, initialStock *int32) (Product, error) {
	name = normalizeProductName(name)
	if name == "" {
		return Product{}, ErrInvalidProductName
	}
	if priceCents <= 0 {
		return Product{}, ErrInvalidPrice
	}
	if merchantID == uuid.Nil {
		return Product{}, ErrInvalidMerchantID
	}
	if initialStock == nil || *initialStock < 0 {
		return Product{}, ErrInvalidInitialStock
	}

	now := time.Now().UTC()
	listing := catalog.Listing{
		ID:          uuid.New(),
		Name:        name,
		Description: normalizeProductDescription(description, name),
		PriceCents:  priceCents,
		MerchantID:  merchantID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	eventID := uuid.New()
	payload, err := proto.Marshal(&productsv1.ProductCreated{
		EventId: eventID.String(), SchemaVersion: 1, OccurredAt: timestamppb.New(listing.CreatedAt),
		ProductId: listing.ID.String(), ProductVersion: 1, Name: listing.Name,
		Description: listing.Description, PriceCents: listing.PriceCents, MerchantId: merchantID.String(),
		InitialQty: initialStock,
	})
	if err != nil {
		return Product{}, err
	}
	if err := s.store.CreateListingAndOutbox(ctx, listing, catalog.OutboxEvent{
		ID:             eventID,
		AggregateID:    listing.ID,
		EventType:      messaging.EventTypeProductCreated,
		Payload:        payload,
		TracingContext: sharedtrace.SerializeContext(ctx),
	}); err != nil {
		return Product{}, err
	}
	return mapListing(listing), nil
}
