package grpcserver

import (
	"refurbished-marketplace/services/inventory/internal/service"
	inventoryv1 "refurbished-marketplace/shared/proto/inventory/v1"
)

type Server struct {
	inventoryv1.UnimplementedInventoryServiceServer
	svc *service.Service
}

func New(svc *service.Service) *Server {
	return &Server{svc: svc}
}
