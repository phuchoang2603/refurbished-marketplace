package grpcserver

import (
	"github.com/phuchoang2603/refurbished-marketplace/services/products/internal/service"
	productsv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/products/v1"
)

type Server struct {
	productsv1.UnimplementedProductsServiceServer
	svc *service.Service
}

func New(svc *service.Service) *Server {
	return &Server{svc: svc}
}
