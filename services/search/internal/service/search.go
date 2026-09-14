package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/meilisearch/meilisearch-go"
	searchv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/search/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Service) SearchProducts(ctx context.Context, query, merchantID string, limit, offset int32) ([]*searchv1.ListingHit, error) {
	if err := validateListPagination(limit, offset); err != nil {
		return nil, err
	}
	merchantID = strings.TrimSpace(merchantID)
	if merchantID != "" {
		if _, err := uuid.Parse(merchantID); err != nil {
			return nil, ErrInvalidMerchantID
		}
	}

	req := &meilisearch.SearchRequest{
		Limit:  int64(limit),
		Offset: int64(offset),
		Sort:   []string{"created_at:desc"},
	}
	if merchantID != "" {
		req.Filter = fmt.Sprintf(`merchant_id = "%s"`, merchantID)
	}

	res, err := s.index.SearchWithContext(ctx, query, req)
	if err != nil {
		return nil, err
	}

	var docs []listingDocument
	if err := res.Hits.DecodeInto(&docs); err != nil {
		return nil, fmt.Errorf("decode search hits: %w", err)
	}

	out := make([]*searchv1.ListingHit, 0, len(docs))
	for _, doc := range docs {
		out = append(out, &searchv1.ListingHit{
			Id:          doc.ID,
			Name:        doc.Name,
			Description: doc.Description,
			PriceCents:  doc.PriceCents,
			MerchantId:  doc.MerchantID,
			CreatedAt:   timestamppb.New(unixSeconds(doc.CreatedAt)),
		})
	}
	return out, nil
}

func validateListPagination(limit, offset int32) error {
	if limit <= 0 || limit > 100 {
		return ErrInvalidListLimit
	}
	if offset < 0 {
		return ErrInvalidListOffset
	}
	return nil
}
