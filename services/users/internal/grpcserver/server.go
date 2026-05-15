package grpcserver

import (
	"refurbished-marketplace/services/users/internal/service"
	usersv1 "refurbished-marketplace/shared/proto/users/v1"
)

type Server struct {
	usersv1.UnimplementedUsersServiceServer
	svc *service.Service
}

func New(svc *service.Service) *Server {
	return &Server{svc: svc}
}
