package grpcserver

import (
	"github.com/phuchoang2603/refurbished-marketplace/services/inventory/internal/service"
	inventoryv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/inventory/v1"
)

type Server struct {
	inventoryv1.UnimplementedInventoryServiceServer
	svc *service.Service
}

func New(svc *service.Service) *Server {
	return &Server{svc: svc}
}
