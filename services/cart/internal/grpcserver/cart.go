package grpcserver

import (
	"context"

	"github.com/phuchoang2603/refurbished-marketplace/services/cart/internal/service"
	"github.com/phuchoang2603/refurbished-marketplace/shared/err/grpcerr"
	cartv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/cart/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func mapCart(c service.Cart) *cartv1.Cart {
	items := make([]*cartv1.CartItem, 0, len(c.Items))
	for _, item := range c.Items {
		items = append(items, &cartv1.CartItem{
			ProductId:      item.ProductID,
			Quantity:       item.Quantity,
			MerchantId:     item.MerchantID,
			ProductName:    item.ProductName,
			UnitPriceCents: item.UnitPriceCents,
		})
	}
	return &cartv1.Cart{CartId: c.CartID, Items: items, CreatedAt: timestamppb.New(c.CreatedAt), UpdatedAt: timestamppb.New(c.UpdatedAt)}
}

func (s *Server) GetCart(ctx context.Context, req *cartv1.GetCartRequest) (*cartv1.Cart, error) {
	if _, err := grpcerr.ParseUUID(req.GetCartId(), "cart id"); err != nil {
		return nil, err
	}
	c, err := s.cart.GetCart(ctx, req.GetCartId())
	if err != nil {
		return nil, grpcerr.Map(err, grpcerr.Mapping{Err: service.ErrInvalidCartID, Code: codes.InvalidArgument})
	}
	return mapCart(c), nil
}

func cartWriteMappings(err error) error {
	return grpcerr.Map(
		err,
		grpcerr.Mapping{Err: service.ErrInvalidCartID, Code: codes.InvalidArgument},
		grpcerr.Mapping{Err: service.ErrInvalidProductID, Code: codes.InvalidArgument},
		grpcerr.Mapping{Err: service.ErrInvalidMerchantID, Code: codes.InvalidArgument},
		grpcerr.Mapping{Err: service.ErrInvalidQuantity, Code: codes.InvalidArgument},
		grpcerr.Mapping{Err: service.ErrInvalidSnapshot, Code: codes.InvalidArgument},
		grpcerr.Mapping{Err: service.ErrEmptyProductIDs, Code: codes.InvalidArgument},
	)
}

func (s *Server) AddCartItem(ctx context.Context, req *cartv1.AddCartItemRequest) (*cartv1.Cart, error) {
	if _, err := grpcerr.ParseUUID(req.GetCartId(), "cart id"); err != nil {
		return nil, err
	}
	if _, err := grpcerr.ParseUUID(req.GetProductId(), "product id"); err != nil {
		return nil, err
	}
	if _, err := grpcerr.ParseUUID(req.GetMerchantId(), "merchant id"); err != nil {
		return nil, err
	}
	c, err := s.cart.AddCartItem(ctx, req.GetCartId(), req.GetProductId(), req.GetMerchantId(), req.GetProductName(), req.GetQuantity(), req.GetUnitPriceCents())
	if err != nil {
		return nil, cartWriteMappings(err)
	}
	return mapCart(c), nil
}

func (s *Server) SetCartItemQuantity(ctx context.Context, req *cartv1.SetCartItemQuantityRequest) (*cartv1.Cart, error) {
	if _, err := grpcerr.ParseUUID(req.GetCartId(), "cart id"); err != nil {
		return nil, err
	}
	if _, err := grpcerr.ParseUUID(req.GetProductId(), "product id"); err != nil {
		return nil, err
	}
	if req.GetQuantity() > 0 {
		if _, err := grpcerr.ParseUUID(req.GetMerchantId(), "merchant id"); err != nil {
			return nil, err
		}
	}
	c, err := s.cart.SetCartItemQuantity(ctx, req.GetCartId(), req.GetProductId(), req.GetMerchantId(), req.GetProductName(), req.GetQuantity(), req.GetUnitPriceCents())
	if err != nil {
		return nil, cartWriteMappings(err)
	}
	return mapCart(c), nil
}

func (s *Server) RemoveCartItems(ctx context.Context, req *cartv1.RemoveCartItemsRequest) (*cartv1.Cart, error) {
	if _, err := grpcerr.ParseUUID(req.GetCartId(), "cart id"); err != nil {
		return nil, err
	}
	for _, productID := range req.GetProductIds() {
		if _, err := grpcerr.ParseUUID(productID, "product id"); err != nil {
			return nil, err
		}
	}
	c, err := s.cart.RemoveCartItems(ctx, req.GetCartId(), req.GetProductIds())
	if err != nil {
		return nil, cartWriteMappings(err)
	}
	return mapCart(c), nil
}

func (s *Server) ClearCart(ctx context.Context, req *cartv1.ClearCartRequest) (*cartv1.Empty, error) {
	if _, err := grpcerr.ParseUUID(req.GetCartId(), "cart id"); err != nil {
		return nil, err
	}
	if err := s.cart.ClearCart(ctx, req.GetCartId()); err != nil {
		return nil, grpcerr.Map(err, grpcerr.Mapping{Err: service.ErrInvalidCartID, Code: codes.InvalidArgument})
	}
	return &cartv1.Empty{}, nil
}
