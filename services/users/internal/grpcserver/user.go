package grpcserver

import (
	"context"

	"refurbished-marketplace/services/users/internal/service"
	"refurbished-marketplace/shared/grpcerr"
	usersv1 "refurbished-marketplace/shared/proto/users/v1"

	"google.golang.org/grpc/codes"
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
		return nil, grpcerr.Map(
			err,
			grpcerr.Mapping{Err: service.ErrInvalidEmail, Code: codes.InvalidArgument},
			grpcerr.Mapping{Err: service.ErrInvalidPassword, Code: codes.InvalidArgument},
			grpcerr.Mapping{Err: service.ErrEmailTaken, Code: codes.AlreadyExists},
		)
	}

	return mapUser(u), nil
}

func (s *Server) GetUserByID(ctx context.Context, req *usersv1.GetUserByIDRequest) (*usersv1.User, error) {
	id, err := grpcerr.ParseUUID(req.GetId(), "id")
	if err != nil {
		return nil, err
	}

	u, err := s.svc.GetUserByID(ctx, id)
	if err != nil {
		return nil, grpcerr.Map(err, grpcerr.Mapping{Err: service.ErrUserNotFound, Code: codes.NotFound, Message: "user not found"})
	}

	return mapUser(u), nil
}
