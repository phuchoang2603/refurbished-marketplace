package grpcserver

import (
	"github.com/phuchoang2603/refurbished-marketplace/services/cart/internal/service"
	cartv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/cart/v1"
)

type Server struct {
	cartv1.UnimplementedCartServiceServer
	cart *service.Service
}

func New(cart *service.Service) *Server {
	return &Server{cart: cart}
}
