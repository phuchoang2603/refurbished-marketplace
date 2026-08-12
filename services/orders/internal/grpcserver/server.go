package grpcserver

import (
	"github.com/phuchoang2603/refurbished-marketplace/services/orders/internal/service"
	ordersv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/orders/v1"
)

type Server struct {
	ordersv1.UnimplementedOrdersServiceServer
	svc *service.Service
}

func New(svc *service.Service) *Server {
	return &Server{svc: svc}
}
