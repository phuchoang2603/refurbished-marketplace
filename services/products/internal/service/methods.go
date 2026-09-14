package service

import (
	"context"
	"errors"
	"time"

	"github.com/phuchoang2603/refurbished-marketplace/services/products/internal/catalog"

	"github.com/google/uuid"
)

type Product struct {
	ID          uuid.UUID
	Name        string
	Description string
	PriceCents  int64
	MerchantID  uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (s *Service) GetProductByID(ctx context.Context, id uuid.UUID) (Product, error) {
	if err := validateProductID(id); err != nil {
		return Product{}, err
	}

	p, err := s.store.GetListingByID(ctx, id)
	if err != nil {
		if errors.Is(err, catalog.ErrListingNotFound) {
			return Product{}, ErrProductNotFound
		}
		return Product{}, err
	}

	return mapListing(p), nil
}

const maxProductsByIDs = 100

func (s *Service) GetProductsByIDs(ctx context.Context, ids []uuid.UUID) ([]Product, error) {
	if len(ids) > maxProductsByIDs {
		return nil, ErrInvalidBatchSize
	}
	if len(ids) == 0 {
		return []Product{}, nil
	}

	seen := make(map[uuid.UUID]struct{}, len(ids))
	unique := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if err := validateProductID(id); err != nil {
			return nil, err
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		unique = append(unique, id)
	}

	rows, err := s.store.GetListingsByIDs(ctx, unique)
	if err != nil {
		return nil, err
	}

	result := make([]Product, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapListing(row))
	}
	return result, nil
}

func (s *Service) ListProducts(ctx context.Context, limit, offset int32) ([]Product, error) {
	if err := validateListPagination(limit, offset); err != nil {
		return nil, err
	}

	rows, err := s.store.ListListings(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	result := make([]Product, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapListing(row))
	}
	return result, nil
}
