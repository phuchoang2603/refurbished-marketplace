// Package grpcserver exposes the payment gRPC API.
package grpcserver

import (
	"refurbished-marketplace/services/payment/internal/service"
	paymentv1 "refurbished-marketplace/shared/proto/payment/v1"
)

type Server struct {
	paymentv1.UnimplementedPaymentServiceServer
	svc *service.Service
}

func New(svc *service.Service) *Server {
	return &Server{svc: svc}
}
