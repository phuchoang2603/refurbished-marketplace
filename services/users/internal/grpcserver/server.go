package grpcserver

import (
	"context"
	"errors"

	"refurbished-marketplace/services/users/internal/service"
	usersv1 "refurbished-marketplace/services/users/proto/v1"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	usersv1.UnimplementedUsersServiceServer
	svc *service.Service
}

func New(svc *service.Service) *Server {
	return &Server{svc: svc}
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

func (s *Server) Login(ctx context.Context, req *usersv1.LoginRequest) (*usersv1.TokenResponse, error) {
	tokens, err := s.svc.Login(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return mapTokens(tokens), nil
}

func (s *Server) Refresh(ctx context.Context, req *usersv1.RefreshRequest) (*usersv1.TokenResponse, error) {
	tokens, err := s.svc.Refresh(ctx, req.GetRefreshToken())
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidToken), errors.Is(err, service.ErrTokenExpired), errors.Is(err, service.ErrTokenRevoked):
			return nil, status.Error(codes.Unauthenticated, "invalid refresh token")
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return mapTokens(tokens), nil
}

func (s *Server) Logout(ctx context.Context, req *usersv1.LogoutRequest) (*usersv1.LogoutResponse, error) {
	err := s.svc.Logout(ctx, req.GetRefreshToken())
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidToken), errors.Is(err, service.ErrTokenExpired), errors.Is(err, service.ErrTokenRevoked):
			return nil, status.Error(codes.Unauthenticated, "invalid refresh token")
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return &usersv1.LogoutResponse{}, nil
}

func mapUser(u service.User) *usersv1.User {
	return &usersv1.User{
		Id:        u.ID.String(),
		Email:     u.Email,
		CreatedAt: timestamppb.New(u.CreatedAt),
		UpdatedAt: timestamppb.New(u.UpdatedAt),
	}
}

func mapTokens(t service.Tokens) *usersv1.TokenResponse {
	return &usersv1.TokenResponse{
		AccessToken:  t.AccessToken,
		RefreshToken: t.RefreshToken,
		TokenType:    t.TokenType,
		ExpiresIn:    t.ExpiresIn,
	}
}
