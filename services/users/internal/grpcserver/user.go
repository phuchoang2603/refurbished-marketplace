package grpcserver

import (
	"context"
	"errors"

	"refurbished-marketplace/services/users/internal/service"
	usersv1 "refurbished-marketplace/shared/proto/users/v1"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func mapUser(u service.User) *usersv1.User {
	return &usersv1.User{
		Id:        u.ID.String(),
		Email:     u.Email,
		CreatedAt: timestamppb.New(u.CreatedAt),
		UpdatedAt: timestamppb.New(u.UpdatedAt),
	}
}

func (s *Server) CreateUser(ctx context.Context, req *usersv1.CreateUserRequest) (*usersv1.User, error) {
	u, err := s.svc.CreateUser(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidEmail), errors.Is(err, service.ErrInvalidPassword):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case errors.Is(err, service.ErrEmailTaken):
			return nil, status.Error(codes.AlreadyExists, err.Error())
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return mapUser(u), nil
}

func (s *Server) GetUserByID(ctx context.Context, req *usersv1.GetUserByIDRequest) (*usersv1.User, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	u, err := s.svc.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return mapUser(u), nil
}
