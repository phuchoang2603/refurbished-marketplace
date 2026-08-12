package grpcserver

import (
	"github.com/phuchoang2603/refurbished-marketplace/services/payment/internal/service"
	paymentv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/payment/v1"
)

type Server struct {
	paymentv1.UnimplementedPaymentServiceServer
	svc *service.Service
}

func New(svc *service.Service) *Server {
	return &Server{svc: svc}
}
