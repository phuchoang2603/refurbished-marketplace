package grpcserver

import (
	"context"
	"errors"

	"refurbished-marketplace/services/orders/internal/service"
	ordersv1 "refurbished-marketplace/shared/proto/orders/v1"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func mapOrder(o service.Order) *ordersv1.Order {
	items := make([]*ordersv1.OrderItem, 0, len(o.Items))
	for _, item := range o.Items {
		items = append(items, &ordersv1.OrderItem{
			Id:             item.ID.String(),
			OrderId:        item.OrderID.String(),
			ProductId:      item.ProductID.String(),
			Quantity:       item.Quantity,
			UnitPriceCents: item.UnitPriceCents,
			LineTotalCents: item.LineTotalCents,
			CreatedAt:      timestamppb.New(item.CreatedAt),
		})
	}

	return &ordersv1.Order{
		Id:          o.ID.String(),
		BuyerUserId: o.BuyerUserID.String(),
		Status:      o.Status,
		TotalCents:  o.TotalCents,
		CreatedAt:   timestamppb.New(o.CreatedAt),
		UpdatedAt:   timestamppb.New(o.UpdatedAt),
		Items:       items,
	}
}

func (s *Server) CreateOrder(ctx context.Context, req *ordersv1.CreateOrderRequest) (*ordersv1.Order, error) {
	buyerID, err := uuid.Parse(req.GetBuyerUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid buyer user id")
	}
	items := make([]service.OrderItemInput, 0, len(req.GetItems()))
	for _, item := range req.GetItems() {
		productID, err := uuid.Parse(item.GetProductId())
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid product id")
		}
		items = append(items, service.OrderItemInput{ProductID: productID, Quantity: item.GetQuantity(), UnitPriceCents: item.GetUnitPriceCents()})
	}

	order, err := s.svc.CreateOrder(ctx, buyerID, items, req.TotalCents)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidBuyerID), errors.Is(err, service.ErrInvalidProductID), errors.Is(err, service.ErrInvalidQuantity):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return mapOrder(order), nil
}

func (s *Server) GetOrderByID(ctx context.Context, req *ordersv1.GetOrderByIDRequest) (*ordersv1.Order, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	order, err := s.svc.GetOrderByID(ctx, id)
	if err != nil {
		if errors.Is(err, service.ErrOrderNotFound) {
			return nil, status.Error(codes.NotFound, "order not found")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return mapOrder(order), nil
}

func (s *Server) ListOrdersByBuyer(ctx context.Context, req *ordersv1.ListOrdersByBuyerRequest) (*ordersv1.ListOrdersByBuyerResponse, error) {
	buyerID, err := uuid.Parse(req.GetBuyerUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid buyer user id")
	}

	orders, err := s.svc.ListOrdersByBuyer(ctx, buyerID, req.GetLimit(), req.GetOffset())
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidBuyerID), errors.Is(err, service.ErrInvalidQuantity):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	out := make([]*ordersv1.Order, 0, len(orders))
	for _, order := range orders {
		out = append(out, mapOrder(order))
	}

	return &ordersv1.ListOrdersByBuyerResponse{Orders: out}, nil
}

func (s *Server) UpdateOrderStatus(ctx context.Context, req *ordersv1.UpdateOrderStatusRequest) (*ordersv1.Order, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	order, err := s.svc.UpdateOrderStatus(ctx, id, req.GetStatus())
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidStatus):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case errors.Is(err, service.ErrOrderNotFound):
			return nil, status.Error(codes.NotFound, "order not found")
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return mapOrder(order), nil
}
