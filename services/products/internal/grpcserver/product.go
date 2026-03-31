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
		OwnerUserId: p.OwnerUserID.String(),
		Name:        p.Name,
		Description: p.Description,
		PriceCents:  p.PriceCents,
		Stock:       p.Stock,
		CreatedAt:   timestamppb.New(p.CreatedAt),
		UpdatedAt:   timestamppb.New(p.UpdatedAt),
	}
}

func (s *Server) CreateProduct(ctx context.Context, req *productsv1.CreateProductRequest) (*productsv1.Product, error) {
	ownerUserID, err := uuid.Parse(req.GetOwnerUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid owner user id")
	}

	p, err := s.svc.CreateProduct(ctx, ownerUserID, req.GetName(), req.GetDescription(), req.GetPriceCents(), req.GetStock())
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidOwner), errors.Is(err, service.ErrInvalidProductName), errors.Is(err, service.ErrInvalidPrice), errors.Is(err, service.ErrInvalidStock):
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

func (s *Server) UpdateProduct(ctx context.Context, req *productsv1.UpdateProductRequest) (*productsv1.Product, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	ownerUserID, err := uuid.Parse(req.GetOwnerUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid owner user id")
	}

	in := service.UpdateProductInput{}
	if req.Name != nil {
		v := req.GetName()
		in.Name = &v
	}
	if req.Description != nil {
		v := req.GetDescription()
		in.Description = &v
	}
	if req.PriceCents != nil {
		v := req.GetPriceCents()
		in.PriceCents = &v
	}
	if req.Stock != nil {
		v := req.GetStock()
		in.Stock = &v
	}

	p, err := s.svc.UpdateProduct(ctx, id, ownerUserID, in)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidOwner), errors.Is(err, service.ErrInvalidProductName), errors.Is(err, service.ErrInvalidPrice), errors.Is(err, service.ErrInvalidStock):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case errors.Is(err, service.ErrForbiddenProduct):
			return nil, status.Error(codes.PermissionDenied, "forbidden")
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return mapProduct(p), nil
}

func (s *Server) DeleteProduct(ctx context.Context, req *productsv1.DeleteProductRequest) (*productsv1.DeleteProductResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	ownerUserID, err := uuid.Parse(req.GetOwnerUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid owner user id")
	}

	err = s.svc.DeleteProduct(ctx, id, ownerUserID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidOwner):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case errors.Is(err, service.ErrForbiddenProduct):
			return nil, status.Error(codes.PermissionDenied, "forbidden")
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return &productsv1.DeleteProductResponse{}, nil
}
