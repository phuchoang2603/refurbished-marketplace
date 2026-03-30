package grpcserver

import (
	"context"
	"errors"

	"refurbished-marketplace/services/users/internal/service"
	usersv1 "refurbished-marketplace/shared/proto/users/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func mapTokens(t service.Tokens) *usersv1.TokenResponse {
	return &usersv1.TokenResponse{
		AccessToken:  t.AccessToken,
		RefreshToken: t.RefreshToken,
		TokenType:    t.TokenType,
		ExpiresIn:    t.ExpiresIn,
	}
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
