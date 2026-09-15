package grpcserver

import (
	"context"

	"github.com/phuchoang2603/refurbished-marketplace/services/products/internal/service"
	"github.com/phuchoang2603/refurbished-marketplace/shared/err/grpcerr"
	productsv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/products/v1"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func mapProduct(p service.Product) *productsv1.Product {
	return &productsv1.Product{
		Id:          p.ID.String(),
		Name:        p.Name,
		Description: p.Description,
		PriceCents:  p.PriceCents,
		MerchantId:  p.MerchantID.String(),
		CreatedAt:   timestamppb.New(p.CreatedAt),
		UpdatedAt:   timestamppb.New(p.UpdatedAt),
	}
}

func (s *Server) CreateProduct(ctx context.Context, req *productsv1.CreateProductRequest) (*productsv1.Product, error) {
	merchantID, err := grpcerr.ParseUUID(req.GetMerchantId(), "merchant id")
	if err != nil {
		return nil, err
	}

	p, err := s.svc.CreateProduct(ctx, req.GetName(), req.GetDescription(), req.GetPriceCents(), merchantID, req.InitialStock)
	if err != nil {
		return nil, grpcerr.Map(
			err,
			grpcerr.Mapping{Err: service.ErrInvalidProductName, Code: codes.InvalidArgument},
			grpcerr.Mapping{Err: service.ErrInvalidInitialStock, Code: codes.InvalidArgument},
			grpcerr.Mapping{Err: service.ErrInvalidPrice, Code: codes.InvalidArgument},
			grpcerr.Mapping{Err: service.ErrInvalidMerchantID, Code: codes.InvalidArgument},
		)
	}

	return mapProduct(p), nil
}

func (s *Server) GetProductByID(ctx context.Context, req *productsv1.GetProductByIDRequest) (*productsv1.Product, error) {
	id, err := grpcerr.ParseUUID(req.GetId(), "id")
	if err != nil {
		return nil, err
	}

	p, err := s.svc.GetProductByID(ctx, id)
	if err != nil {
		return nil, grpcerr.Map(err, grpcerr.Mapping{Err: service.ErrProductNotFound, Code: codes.NotFound, Message: "product not found"})
	}

	return mapProduct(p), nil
}

func (s *Server) GetProductsByIDs(ctx context.Context, req *productsv1.GetProductsByIDsRequest) (*productsv1.GetProductsByIDsResponse, error) {
	ids := make([]uuid.UUID, 0, len(req.GetIds()))
	for _, raw := range req.GetIds() {
		id, err := grpcerr.ParseUUID(raw, "id")
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	products, err := s.svc.GetProductsByIDs(ctx, ids)
	if err != nil {
		return nil, grpcerr.Map(
			err,
			grpcerr.Mapping{Err: service.ErrInvalidBatchSize, Code: codes.InvalidArgument},
			grpcerr.Mapping{Err: service.ErrInvalidProductID, Code: codes.InvalidArgument},
		)
	}

	out := make([]*productsv1.Product, 0, len(products))
	for _, p := range products {
		out = append(out, mapProduct(p))
	}
	return &productsv1.GetProductsByIDsResponse{Products: out}, nil
}
