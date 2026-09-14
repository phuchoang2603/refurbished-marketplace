package grpcserver

import (
	"context"

	"github.com/phuchoang2603/refurbished-marketplace/services/search/internal/service"
	"github.com/phuchoang2603/refurbished-marketplace/shared/err/grpcerr"
	searchv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/search/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) SearchProducts(ctx context.Context, req *searchv1.SearchProductsRequest) (*searchv1.SearchProductsResponse, error) {
	listings, err := s.svc.SearchProducts(ctx, req.GetQuery(), req.GetMerchantId(), req.GetLimit(), req.GetOffset())
	if err != nil {
		mapped := grpcerr.Map(
			err,
			grpcerr.Mapping{Err: service.ErrInvalidListLimit, Code: codes.InvalidArgument},
			grpcerr.Mapping{Err: service.ErrInvalidListOffset, Code: codes.InvalidArgument},
			grpcerr.Mapping{Err: service.ErrInvalidMerchantID, Code: codes.InvalidArgument},
		)
		if status.Code(mapped) == codes.Internal {
			return nil, status.Error(codes.Unavailable, "search unavailable")
		}
		return nil, mapped
	}
	return &searchv1.SearchProductsResponse{Listings: listings}, nil
}
