package catalog

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/phuchoang2603/refurbished-marketplace/services/products/internal/database"
)

var ErrListingNotFound = errors.New("listing not found")

type Store struct {
	db      *sql.DB
	queries *database.Queries
}

type Listing = database.Product

type OutboxEvent struct {
	ID             uuid.UUID
	AggregateID    uuid.UUID
	EventType      string
	Payload        []byte
	TracingContext string
}

func New(db *sql.DB) *Store {
	return &Store{db: db, queries: database.New(db)}
}

func (s *Store) Database() *sql.DB {
	return s.db
}

func (s *Store) CreateListingAndOutbox(ctx context.Context, listing Listing, event OutboxEvent) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	queries := s.queries.WithTx(tx)
	if err := queries.CreateListing(ctx, database.CreateListingParams(listing)); err != nil {
		return err
	}
	if err := queries.InsertOutboxEvent(ctx, database.InsertOutboxEventParams{
		ID:                 event.ID,
		AggregateID:        event.AggregateID,
		EventType:          event.EventType,
		Payload:            event.Payload,
		Tracingspancontext: event.TracingContext,
	}); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) GetListingByID(ctx context.Context, id uuid.UUID) (Listing, error) {
	listing, err := s.queries.GetListingByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return Listing{}, ErrListingNotFound
	}
	return listing, err
}

func (s *Store) GetListingsByIDs(ctx context.Context, ids []uuid.UUID) ([]Listing, error) {
	return s.queries.GetListingsByIDs(ctx, ids)
}
