// Package grpcserver provides the gRPC server implementation for the users service. It translates gRPC requests into calls to the underlying service layer and maps service responses back to gRPC responses.
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
