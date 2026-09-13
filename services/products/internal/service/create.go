package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/phuchoang2603/refurbished-marketplace/services/products/internal/database"
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
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Product{}, err
	}
	defer func() { _ = tx.Rollback() }()
	q := s.queries.WithTx(tx)
	created, err := q.CreateProduct(ctx, database.CreateProductParams{
		ID: uuid.New(), Name: name, Description: normalizeProductDescription(description, name),
		PriceCents: priceCents, MerchantID: merchantID,
	})
	if err != nil {
		return Product{}, err
	}
	eventID := uuid.New()
	payload, err := proto.Marshal(&productsv1.ProductCreated{
		EventId: eventID.String(), SchemaVersion: 1, OccurredAt: timestamppb.New(created.CreatedAt),
		ProductId: created.ID.String(), ProductVersion: 1, Name: created.Name,
		Description: created.Description, PriceCents: created.PriceCents, MerchantId: merchantID.String(),
		InitialQty: initialStock,
	})
	if err != nil {
		return Product{}, err
	}
	if err := q.CreateProductOutbox(ctx, database.CreateProductOutboxParams{
		ID: eventID, AggregateID: created.ID, EventType: messaging.EventTypeProductCreated,
		Payload: payload, Tracingspancontext: sharedtrace.SerializeContext(ctx),
	}); err != nil {
		return Product{}, err
	}
	if err := tx.Commit(); err != nil {
		return Product{}, err
	}
	return mapDBProduct(created), nil
}
