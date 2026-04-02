package grpcserver

import (
	"refurbished-marketplace/services/cart/internal/service"
	cartv1 "refurbished-marketplace/shared/proto/cart/v1"
)

type Server struct {
	cartv1.UnimplementedCartServiceServer
	cart *service.Service
}

func New(cart *service.Service) *Server {
	return &Server{cart: cart}
}
