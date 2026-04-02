package grpcserver

import (
	"context"
	"errors"

	"refurbished-marketplace/services/cart/internal/service"
	cartv1 "refurbished-marketplace/shared/proto/cart/v1"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func mapCart(c service.Cart) *cartv1.Cart {
	items := make([]*cartv1.CartItem, 0, len(c.Items))
	for _, item := range c.Items {
		items = append(items, &cartv1.CartItem{ProductId: item.ProductID, Quantity: item.Quantity})
	}
	return &cartv1.Cart{CartId: c.CartID, Items: items, CreatedAt: timestamppb.New(c.CreatedAt), UpdatedAt: timestamppb.New(c.UpdatedAt)}
}

func (s *Server) GetCart(ctx context.Context, req *cartv1.GetCartRequest) (*cartv1.Cart, error) {
	if _, err := uuid.Parse(req.GetCartId()); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid cart id")
	}
	c, err := s.cart.GetCart(ctx, req.GetCartId())
	if err != nil {
		if errors.Is(err, service.ErrInvalidCartID) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, "internal error")
	}
	return mapCart(c), nil
}

func (s *Server) AddCartItem(ctx context.Context, req *cartv1.AddCartItemRequest) (*cartv1.Cart, error) {
	if _, err := uuid.Parse(req.GetCartId()); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid cart id")
	}
	if _, err := uuid.Parse(req.GetProductId()); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid product id")
	}
	c, err := s.cart.AddCartItem(ctx, req.GetCartId(), req.GetProductId(), req.GetQuantity())
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidProductID), errors.Is(err, service.ErrInvalidQuantity), errors.Is(err, service.ErrInvalidCartID):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}
	return mapCart(c), nil
}

func (s *Server) SetCartItemQuantity(ctx context.Context, req *cartv1.SetCartItemQuantityRequest) (*cartv1.Cart, error) {
	if _, err := uuid.Parse(req.GetCartId()); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid cart id")
	}
	if _, err := uuid.Parse(req.GetProductId()); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid product id")
	}
	c, err := s.cart.SetCartItemQuantity(ctx, req.GetCartId(), req.GetProductId(), req.GetQuantity())
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidProductID), errors.Is(err, service.ErrInvalidQuantity), errors.Is(err, service.ErrInvalidCartID), errors.Is(err, service.ErrItemNotFound):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}
	return mapCart(c), nil
}

func (s *Server) RemoveCartItem(ctx context.Context, req *cartv1.RemoveCartItemRequest) (*cartv1.Cart, error) {
	if _, err := uuid.Parse(req.GetCartId()); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid cart id")
	}
	if _, err := uuid.Parse(req.GetProductId()); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid product id")
	}
	c, err := s.cart.RemoveCartItem(ctx, req.GetCartId(), req.GetProductId())
	if err != nil {
		switch {
		case errors.Is(err, service.ErrItemNotFound), errors.Is(err, service.ErrInvalidCartID):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}
	return mapCart(c), nil
}

func (s *Server) ClearCart(ctx context.Context, req *cartv1.ClearCartRequest) (*cartv1.Empty, error) {
	if _, err := uuid.Parse(req.GetCartId()); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid cart id")
	}
	if err := s.cart.ClearCart(ctx, req.GetCartId()); err != nil {
		if errors.Is(err, service.ErrInvalidCartID) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &cartv1.Empty{}, nil
}
