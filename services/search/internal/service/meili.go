package service

import (
	"github.com/meilisearch/meilisearch-go"
)

func meiliSettings() meilisearch.Settings {
	return meilisearch.Settings{
		SearchableAttributes: []string{"name", "description"},
		FilterableAttributes: []string{"merchant_id"},
		SortableAttributes:   []string{"created_at"},
	}
}
