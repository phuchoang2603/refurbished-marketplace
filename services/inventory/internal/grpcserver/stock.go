package grpcserver

import (
	"context"

	"github.com/phuchoang2603/refurbished-marketplace/services/inventory/internal/service"
	"github.com/phuchoang2603/refurbished-marketplace/shared/err/grpcerr"
	inventoryv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/inventory/v1"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
)

func mapStock(inv service.Inventory) *inventoryv1.Stock {
	return &inventoryv1.Stock{
		ProductId:    inv.ProductID.String(),
		AvailableQty: inv.AvailableQty,
		ReservedQty:  inv.ReservedQty,
	}
}

func (s *Server) GetStock(ctx context.Context, req *inventoryv1.GetStockRequest) (*inventoryv1.Stock, error) {
	productID, err := grpcerr.ParseUUID(req.GetProductId(), "product id")
	if err != nil {
		return nil, err
	}
	inv, err := s.svc.GetInventoryByProductID(ctx, productID)
	if err != nil {
		return nil, grpcerr.Map(err, grpcerr.Mapping{Err: service.ErrInventoryNotFound, Code: codes.NotFound, Message: "stock not found"})
	}
	return mapStock(inv), nil
}

func (s *Server) GetStocksByIDs(ctx context.Context, req *inventoryv1.GetStocksByIDsRequest) (*inventoryv1.GetStocksByIDsResponse, error) {
	ids := make([]uuid.UUID, 0, len(req.GetProductIds()))
	for _, raw := range req.GetProductIds() {
		id, err := grpcerr.ParseUUID(raw, "product id")
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	rows, err := s.svc.GetInventoriesByProductIDs(ctx, ids)
	if err != nil {
		return nil, grpcerr.Map(
			err,
			grpcerr.Mapping{Err: service.ErrInvalidBatchSize, Code: codes.InvalidArgument},
			grpcerr.Mapping{Err: service.ErrInvalidProductID, Code: codes.InvalidArgument},
		)
	}
	out := make([]*inventoryv1.Stock, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapStock(row))
	}
	return &inventoryv1.GetStocksByIDsResponse{Stocks: out}, nil
}

func (s *Server) ReserveStock(ctx context.Context, req *inventoryv1.ReserveStockRequest) (*inventoryv1.ReserveStockResponse, error) {
	orderID, err := grpcerr.ParseUUID(req.GetOrderId(), "order id")
	if err != nil {
		return nil, err
	}
	merchantID, err := grpcerr.ParseUUID(req.GetMerchantId(), "merchant id")
	if err != nil {
		return nil, err
	}
	items := make([]service.ReservationItemInput, 0, len(req.GetItems()))
	for _, item := range req.GetItems() {
		productID, err := grpcerr.ParseUUID(item.GetProductId(), "product id")
		if err != nil {
			return nil, err
		}
		items = append(items, service.ReservationItemInput{ProductID: productID, Quantity: item.GetQuantity()})
	}
	err = s.svc.ReserveStock(ctx, orderID, merchantID, req.GetTotalCents(), items)
	if err != nil {
		return nil, grpcerr.Map(
			err,
			grpcerr.Mapping{Err: service.ErrInvalidMerchantID, Code: codes.InvalidArgument},
			grpcerr.Mapping{Err: service.ErrInvalidQuantity, Code: codes.InvalidArgument},
			grpcerr.Mapping{Err: service.ErrInvalidProductID, Code: codes.InvalidArgument},
			grpcerr.Mapping{Err: service.ErrInventoryNotFound, Code: codes.FailedPrecondition},
			grpcerr.Mapping{Err: service.ErrInsufficientStock, Code: codes.FailedPrecondition},
		)
	}
	return &inventoryv1.ReserveStockResponse{}, nil
}
