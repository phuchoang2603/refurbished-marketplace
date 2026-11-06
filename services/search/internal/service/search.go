package service

import (
	"context"
	"encoding/json"
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

	query = strings.TrimSpace(query)
	req := &meilisearch.SearchRequest{
		Limit:  int64(limit),
		Offset: int64(offset),
	}
	if query == "" {
		req.Sort = []string{"created_at:desc"}
	}
	if merchantID != "" {
		req.Filter = fmt.Sprintf(`merchant_id = "%s"`, merchantID)
	}

	res, err := s.index.SearchRawWithContext(ctx, query, req)
	if err != nil {
		return nil, err
	}

	var response struct {
		Hits []listingDocument `json:"hits"`
	}
	if err := json.Unmarshal(*res, &response); err != nil {
		return nil, fmt.Errorf("decode search hits: %w", err)
	}

	out := make([]*searchv1.ListingHit, 0, len(response.Hits))
	for _, doc := range response.Hits {
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
