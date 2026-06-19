package grpcserver

import (
	"context"

	"refurbished-marketplace/services/orders/internal/service"
	"refurbished-marketplace/shared/grpcerr"
	ordersv1 "refurbished-marketplace/shared/proto/orders/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func mapOrder(o service.Order) *ordersv1.Order {
	var status ordersv1.OrderStatus
	switch o.Status {
	case service.OrderStatusPending:
		status = ordersv1.OrderStatus_ORDER_STATUS_PENDING
	case service.OrderStatusPaid:
		status = ordersv1.OrderStatus_ORDER_STATUS_PAID
	case service.OrderStatusFailed:
		status = ordersv1.OrderStatus_ORDER_STATUS_FAILED
	default:
		status = ordersv1.OrderStatus_ORDER_STATUS_UNSPECIFIED
	}
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
		MerchantId:  o.MerchantID.String(),
		Status:      ordersv1.OrderStatus(status),
		TotalCents:  o.TotalCents,
		CreatedAt:   timestamppb.New(o.CreatedAt),
		UpdatedAt:   timestamppb.New(o.UpdatedAt),
		Items:       items,
	}
}

func (s *Server) CreateOrder(ctx context.Context, req *ordersv1.CreateOrderRequest) (*ordersv1.Order, error) {
	buyerID, err := grpcerr.ParseUUID(req.GetBuyerUserId(), "buyer user id")
	if err != nil {
		return nil, err
	}
	merchantID, err := grpcerr.ParseUUID(req.GetMerchantId(), "merchant id")
	if err != nil {
		return nil, err
	}
	items := make([]service.OrderItemInput, 0, len(req.GetItems()))
	for _, item := range req.GetItems() {
		productID, err := grpcerr.ParseUUID(item.GetProductId(), "product id")
		if err != nil {
			return nil, err
		}
		items = append(items, service.OrderItemInput{ProductID: productID, Quantity: item.GetQuantity(), UnitPriceCents: item.GetUnitPriceCents()})
	}

	order, err := s.svc.CreateOrder(ctx, buyerID, merchantID, items, req.TotalCents)
	if err != nil {
		return nil, grpcerr.Map(
			err,
			grpcerr.Mapping{Err: service.ErrInvalidBuyerID, Code: codes.InvalidArgument},
			grpcerr.Mapping{Err: service.ErrInvalidMerchantID, Code: codes.InvalidArgument},
			grpcerr.Mapping{Err: service.ErrInvalidProductID, Code: codes.InvalidArgument},
			grpcerr.Mapping{Err: service.ErrInvalidQuantity, Code: codes.InvalidArgument},
			grpcerr.Mapping{Err: service.ErrInvalidUnitPriceCents, Code: codes.InvalidArgument},
			grpcerr.Mapping{Err: service.ErrInvalidTotalCents, Code: codes.InvalidArgument},
		)
	}

	return mapOrder(order), nil
}

func (s *Server) GetOrderByID(ctx context.Context, req *ordersv1.GetOrderByIDRequest) (*ordersv1.Order, error) {
	id, err := grpcerr.ParseUUID(req.GetId(), "id")
	if err != nil {
		return nil, err
	}

	order, err := s.svc.GetOrderByID(ctx, id)
	if err != nil {
		return nil, grpcerr.Map(err, grpcerr.Mapping{Err: service.ErrOrderNotFound, Code: codes.NotFound, Message: "order not found"})
	}

	return mapOrder(order), nil
}

func (s *Server) ListOrdersByBuyer(ctx context.Context, req *ordersv1.ListOrdersByBuyerRequest) (*ordersv1.ListOrdersByBuyerResponse, error) {
	buyerID, err := grpcerr.ParseUUID(req.GetBuyerUserId(), "buyer user id")
	if err != nil {
		return nil, err
	}

	orders, err := s.svc.ListOrdersByBuyer(ctx, buyerID, req.GetLimit(), req.GetOffset())
	if err != nil {
		return nil, grpcerr.Map(
			err,
			grpcerr.Mapping{Err: service.ErrInvalidBuyerID, Code: codes.InvalidArgument},
			grpcerr.Mapping{Err: service.ErrInvalidPagination, Code: codes.InvalidArgument},
		)
	}

	out := make([]*ordersv1.Order, 0, len(orders))
	for _, order := range orders {
		out = append(out, mapOrder(order))
	}

	return &ordersv1.ListOrdersByBuyerResponse{Orders: out}, nil
}

func (s *Server) UpdateOrderStatus(ctx context.Context, req *ordersv1.UpdateOrderStatusRequest) (*ordersv1.Order, error) {
	id, err := grpcerr.ParseUUID(req.GetId(), "id")
	if err != nil {
		return nil, err
	}

	order, err := s.svc.UpdateOrderStatus(ctx, id, req.GetStatus().String())
	if err != nil {
		return nil, grpcerr.Map(
			err,
			grpcerr.Mapping{Err: service.ErrInvalidStatus, Code: codes.InvalidArgument},
			grpcerr.Mapping{Err: service.ErrOrderNotFound, Code: codes.NotFound, Message: "order not found"},
		)
	}

	return mapOrder(order), nil
}
