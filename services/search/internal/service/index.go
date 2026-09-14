package service

import (
	"context"
	"fmt"
	"time"

	"github.com/meilisearch/meilisearch-go"
)

type listingDocument struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	PriceCents  int64  `json:"price_cents"`
	MerchantID  string `json:"merchant_id"`
	CreatedAt   int64  `json:"created_at"`
}

func (s *Service) EnsureIndexSettings(ctx context.Context) error {
	settings := meiliSettings()
	task, err := s.index.UpdateSettingsWithContext(ctx, &settings)
	if err != nil {
		return fmt.Errorf("update listings index settings: %w", err)
	}
	if _, err := s.client.WaitForTaskWithContext(ctx, task.TaskUID, 100*time.Millisecond); err != nil {
		return fmt.Errorf("wait for listings index settings: %w", err)
	}
	return nil
}

func (s *Service) UpsertListing(ctx context.Context, doc listingDocument) error {
	primaryKey := "id"
	task, err := s.index.AddDocumentsWithContext(ctx, []listingDocument{doc}, &meilisearch.DocumentOptions{PrimaryKey: &primaryKey})
	if err != nil {
		return fmt.Errorf("upsert listing: %w", err)
	}
	if _, err := s.client.WaitForTaskWithContext(ctx, task.TaskUID, 50*time.Millisecond); err != nil {
		return fmt.Errorf("wait for listing upsert: %w", err)
	}
	return nil
}
