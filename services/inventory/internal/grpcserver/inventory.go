package grpcserver

import (
	"context"
	"errors"

	"refurbished-marketplace/services/inventory/internal/service"
	inventoryv1 "refurbished-marketplace/shared/proto/inventory/v1"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func mapInventory(inv service.Inventory) *inventoryv1.Inventory {
	return &inventoryv1.Inventory{
		ProductId:    inv.ProductID.String(),
		AvailableQty: inv.AvailableQty,
		ReservedQty:  inv.ReservedQty,
		CreatedAt:    timestamppb.New(inv.CreatedAt),
		UpdatedAt:    timestamppb.New(inv.UpdatedAt),
	}
}

func (s *Server) CreateInventory(ctx context.Context, req *inventoryv1.CreateInventoryRequest) (*inventoryv1.Inventory, error) {
	productID, err := uuid.Parse(req.GetProductId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid product id")
	}

	inv, err := s.svc.CreateInventory(ctx, productID, req.GetAvailableQty())
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidProductID), errors.Is(err, service.ErrInvalidQuantity):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return mapInventory(inv), nil
}

func (s *Server) GetInventoryByProductID(ctx context.Context, req *inventoryv1.GetInventoryByProductIDRequest) (*inventoryv1.Inventory, error) {
	productID, err := uuid.Parse(req.GetProductId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid product id")
	}

	inv, err := s.svc.GetInventoryByProductID(ctx, productID)
	if err != nil {
		if errors.Is(err, service.ErrInventoryNotFound) {
			return nil, status.Error(codes.NotFound, "inventory not found")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return mapInventory(inv), nil
}

func (s *Server) ReserveStock(ctx context.Context, req *inventoryv1.ReserveStockRequest) (*inventoryv1.Inventory, error) {
	productID, err := uuid.Parse(req.GetProductId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid product id")
	}

	inv, err := s.svc.ReserveStock(ctx, productID, req.GetQuantity())
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidProductID), errors.Is(err, service.ErrInvalidQuantity):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case errors.Is(err, service.ErrInsufficientStock):
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return mapInventory(inv), nil
}

func (s *Server) CommitReservation(ctx context.Context, req *inventoryv1.CommitReservationRequest) (*inventoryv1.Inventory, error) {
	productID, err := uuid.Parse(req.GetProductId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid product id")
	}

	inv, err := s.svc.CommitReservation(ctx, productID, req.GetQuantity())
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidProductID), errors.Is(err, service.ErrInvalidQuantity):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case errors.Is(err, service.ErrInventoryNotFound):
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return mapInventory(inv), nil
}

func (s *Server) ReleaseReservation(ctx context.Context, req *inventoryv1.ReleaseReservationRequest) (*inventoryv1.Inventory, error) {
	productID, err := uuid.Parse(req.GetProductId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid product id")
	}

	inv, err := s.svc.ReleaseReservation(ctx, productID, req.GetQuantity())
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidProductID), errors.Is(err, service.ErrInvalidQuantity):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case errors.Is(err, service.ErrInventoryNotFound):
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return mapInventory(inv), nil
}
