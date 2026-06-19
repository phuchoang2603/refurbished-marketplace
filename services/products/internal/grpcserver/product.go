package grpcserver

import (
	"context"

	"refurbished-marketplace/services/products/internal/service"
	"refurbished-marketplace/shared/grpcerr"
	productsv1 "refurbished-marketplace/shared/proto/products/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func mapProduct(p service.Product) *productsv1.Product {
	out := &productsv1.Product{
		Id:          p.ID.String(),
		Name:        p.Name,
		Description: p.Description,
		PriceCents:  p.PriceCents,
		MerchantId:  p.MerchantID.String(),
		CreatedAt:   timestamppb.New(p.CreatedAt),
		UpdatedAt:   timestamppb.New(p.UpdatedAt),
	}
	if p.AvailableQty != nil {
		out.AvailableQty = p.AvailableQty
	}
	if p.ReservedQty != nil {
		out.ReservedQty = p.ReservedQty
	}
	return out
}

func (s *Server) CreateProduct(ctx context.Context, req *productsv1.CreateProductRequest) (*productsv1.Product, error) {
	merchantID, err := grpcerr.ParseUUID(req.GetMerchantId(), "merchant id")
	if err != nil {
		return nil, err
	}
	if req.InitialStock == nil {
		return nil, grpcerr.InvalidArgument("initial stock is required")
	}

	p, err := s.svc.CreateProduct(ctx, req.GetName(), req.GetDescription(), req.GetPriceCents(), merchantID, req.GetInitialStock())
	if err != nil {
		return nil, grpcerr.Map(
			err,
			grpcerr.Mapping{Err: service.ErrInvalidProductName, Code: codes.InvalidArgument},
			grpcerr.Mapping{Err: service.ErrInvalidPrice, Code: codes.InvalidArgument},
			grpcerr.Mapping{Err: service.ErrInvalidMerchantID, Code: codes.InvalidArgument},
			grpcerr.Mapping{Err: service.ErrInvalidQuantity, Code: codes.InvalidArgument},
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

func (s *Server) ListProducts(ctx context.Context, req *productsv1.ListProductsRequest) (*productsv1.ListProductsResponse, error) {
	products, err := s.svc.ListProducts(ctx, req.GetLimit(), req.GetOffset())
	if err != nil {
		return nil, grpcerr.Map(
			err,
			grpcerr.Mapping{Err: service.ErrInvalidListLimit, Code: codes.InvalidArgument},
			grpcerr.Mapping{Err: service.ErrInvalidListOffset, Code: codes.InvalidArgument},
		)
	}

	out := make([]*productsv1.Product, 0, len(products))
	for _, p := range products {
		out = append(out, mapProduct(p))
	}

	return &productsv1.ListProductsResponse{Products: out}, nil
}
