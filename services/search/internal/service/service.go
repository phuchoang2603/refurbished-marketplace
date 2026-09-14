package service

import (
	"errors"

	"github.com/meilisearch/meilisearch-go"
)

var (
	ErrInvalidListLimit  = errors.New("invalid list limit")
	ErrInvalidListOffset = errors.New("invalid list offset")
	ErrInvalidMerchantID = errors.New("invalid merchant id")
)

type Service struct {
	client meilisearch.ServiceManager
	index  meilisearch.IndexManager
}

func New(client meilisearch.ServiceManager, indexUID string) *Service {
	return &Service{client: client, index: client.Index(indexUID)}
}
