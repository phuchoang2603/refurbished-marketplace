package grpcserver

import (
	"context"
	"errors"

	"refurbished-marketplace/services/products/internal/service"
	productsv1 "refurbished-marketplace/shared/proto/products/v1"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	merchantID, err := uuid.Parse(req.GetMerchantId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid merchant id")
	}

	p, err := s.svc.CreateProduct(ctx, req.GetName(), req.GetDescription(), req.GetPriceCents(), merchantID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidProductName), errors.Is(err, service.ErrInvalidPrice), errors.Is(err, service.ErrInvalidMerchantID):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return mapProduct(p), nil
}

func (s *Server) GetProductByID(ctx context.Context, req *productsv1.GetProductByIDRequest) (*productsv1.Product, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	p, err := s.svc.GetProductByID(ctx, id)
	if err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			return nil, status.Error(codes.NotFound, "product not found")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return mapProduct(p), nil
}

func (s *Server) ListProducts(ctx context.Context, req *productsv1.ListProductsRequest) (*productsv1.ListProductsResponse, error) {
	products, err := s.svc.ListProducts(ctx, req.GetLimit(), req.GetOffset())
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidListLimit), errors.Is(err, service.ErrInvalidListOffset):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	out := make([]*productsv1.Product, 0, len(products))
	for _, p := range products {
		out = append(out, mapProduct(p))
	}

	return &productsv1.ListProductsResponse{Products: out}, nil
}
