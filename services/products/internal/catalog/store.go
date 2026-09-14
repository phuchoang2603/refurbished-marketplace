package catalog

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	listingsCollection = "listings"
	outboxCollection   = "catalog_outbox"
)

var ErrListingNotFound = errors.New("listing not found")

type Store struct {
	client   *mongo.Client
	db       *mongo.Database
	listings *mongo.Collection
	outbox   *mongo.Collection
}

type Listing struct {
	ID          uuid.UUID
	Name        string
	Description string
	PriceCents  int64
	MerchantID  uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type OutboxEvent struct {
	ID             uuid.UUID
	AggregateID    uuid.UUID
	EventType      string
	Payload        []byte
	TracingContext string
}

func Open(ctx context.Context, uri string) (*Store, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("connect mongodb: %w", err)
	}
	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(ctx)
		return nil, fmt.Errorf("ping mongodb: %w", err)
	}
	return New(client.Database(databaseNameFromURI(uri))), nil
}

func databaseNameFromURI(uri string) string {
	u, err := url.Parse(uri)
	if err != nil {
		return "catalog"
	}
	name := strings.Trim(u.Path, "/")
	if name == "" {
		return "catalog"
	}
	return name
}

func New(db *mongo.Database) *Store {
	return &Store{
		client:   db.Client(),
		db:       db,
		listings: db.Collection(listingsCollection),
		outbox:   db.Collection(outboxCollection),
	}
}

func (s *Store) Close(ctx context.Context) error {
	if s == nil || s.client == nil {
		return nil
	}
	return s.client.Disconnect(ctx)
}

func (s *Store) Database() *mongo.Database {
	return s.db
}

func (s *Store) CreateListingAndOutbox(ctx context.Context, listing Listing, event OutboxEvent) error {
	session, err := s.client.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sessCtx context.Context) (any, error) {
		if _, err := s.listings.InsertOne(sessCtx, listingDocFrom(listing)); err != nil {
			return nil, err
		}
		if _, err := s.outbox.InsertOne(sessCtx, outboxDocFrom(event)); err != nil {
			return nil, err
		}
		return nil, nil
	})
	return err
}

func (s *Store) GetListingByID(ctx context.Context, id uuid.UUID) (Listing, error) {
	var doc listingDoc
	err := s.listings.FindOne(ctx, bson.M{"_id": id.String()}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return Listing{}, ErrListingNotFound
	}
	if err != nil {
		return Listing{}, err
	}
	return doc.toListing()
}

func (s *Store) GetListingsByIDs(ctx context.Context, ids []uuid.UUID) ([]Listing, error) {
	rawIDs := make([]string, 0, len(ids))
	for _, id := range ids {
		rawIDs = append(rawIDs, id.String())
	}
	cursor, err := s.listings.Find(ctx, bson.M{"_id": bson.M{"$in": rawIDs}})
	if err != nil {
		return nil, err
	}
	defer func() { _ = cursor.Close(ctx) }()

	var docs []listingDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}
	out := make([]Listing, 0, len(docs))
	for _, doc := range docs {
		listing, err := doc.toListing()
		if err != nil {
			return nil, err
		}
		out = append(out, listing)
	}
	return out, nil
}

func (s *Store) ListListings(ctx context.Context, limit, offset int32) ([]Listing, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}, {Key: "_id", Value: -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))
	cursor, err := s.listings.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer func() { _ = cursor.Close(ctx) }()

	var docs []listingDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}
	out := make([]Listing, 0, len(docs))
	for _, doc := range docs {
		listing, err := doc.toListing()
		if err != nil {
			return nil, err
		}
		out = append(out, listing)
	}
	return out, nil
}

type listingDoc struct {
	ID          string    `bson:"_id"`
	Name        string    `bson:"name"`
	Description string    `bson:"description"`
	PriceCents  int64     `bson:"price_cents"`
	MerchantID  string    `bson:"merchant_id"`
	CreatedAt   time.Time `bson:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at"`
}

type outboxDoc struct {
	OID                string `bson:"_id"`
	ID                 string `bson:"id"`
	AggregateID        string `bson:"aggregate_id"`
	EventType          string `bson:"event_type"`
	Payload            []byte `bson:"payload"`
	TracingSpanContext string `bson:"tracingspancontext"`
}

func listingDocFrom(l Listing) listingDoc {
	return listingDoc{
		ID:          l.ID.String(),
		Name:        l.Name,
		Description: l.Description,
		PriceCents:  l.PriceCents,
		MerchantID:  l.MerchantID.String(),
		CreatedAt:   l.CreatedAt,
		UpdatedAt:   l.UpdatedAt,
	}
}

func outboxDocFrom(e OutboxEvent) outboxDoc {
	id := e.ID.String()
	return outboxDoc{
		OID:                id,
		ID:                 id,
		AggregateID:        e.AggregateID.String(),
		EventType:          e.EventType,
		Payload:            e.Payload,
		TracingSpanContext: e.TracingContext,
	}
}

func (d listingDoc) toListing() (Listing, error) {
	id, err := uuid.Parse(d.ID)
	if err != nil {
		return Listing{}, fmt.Errorf("listing id: %w", err)
	}
	merchantID, err := uuid.Parse(d.MerchantID)
	if err != nil {
		return Listing{}, fmt.Errorf("merchant id: %w", err)
	}
	return Listing{
		ID:          id,
		Name:        d.Name,
		Description: d.Description,
		PriceCents:  d.PriceCents,
		MerchantID:  merchantID,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}, nil
}
