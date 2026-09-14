package grpcserver

import (
	"github.com/phuchoang2603/refurbished-marketplace/services/search/internal/service"
	searchv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/search/v1"
)

type Server struct {
	searchv1.UnimplementedSearchServiceServer
	svc *service.Service
}

func New(svc *service.Service) *Server {
	return &Server{svc: svc}
}
