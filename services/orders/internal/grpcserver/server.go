package grpcserver

import (
	"refurbished-marketplace/services/orders/internal/service"
	ordersv1 "refurbished-marketplace/shared/proto/orders/v1"
)

type Server struct {
	ordersv1.UnimplementedOrdersServiceServer
	svc *service.Service
}

func New(svc *service.Service) *Server {
	return &Server{svc: svc}
}
