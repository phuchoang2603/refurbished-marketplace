package grpcserver

import (
	"context"

	"refurbished-marketplace/services/users/internal/service"
	"refurbished-marketplace/shared/grpcerr"
	usersv1 "refurbished-marketplace/shared/proto/users/v1"

	"google.golang.org/grpc/codes"
)

func mapTokens(t service.Tokens) *usersv1.TokenResponse {
	return &usersv1.TokenResponse{
		AccessToken:      t.AccessToken,
		RefreshToken:     t.RefreshToken,
		TokenType:        t.TokenType,
		ExpiresIn:        t.ExpiresIn,
		RefreshExpiresIn: t.RefreshExpiresIn,
	}
}

func (s *Server) Login(ctx context.Context, req *usersv1.LoginRequest) (*usersv1.TokenResponse, error) {
	tokens, err := s.svc.Login(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, grpcerr.Map(err, grpcerr.Mapping{Err: service.ErrInvalidCredentials, Code: codes.Unauthenticated, Message: "invalid credentials"})
	}

	return mapTokens(tokens), nil
}

func (s *Server) Refresh(ctx context.Context, req *usersv1.RefreshRequest) (*usersv1.TokenResponse, error) {
	tokens, err := s.svc.Refresh(ctx, req.GetRefreshToken())
	if err != nil {
		return nil, grpcerr.Map(
			err,
			grpcerr.Mapping{Err: service.ErrInvalidToken, Code: codes.Unauthenticated, Message: "invalid refresh token"},
			grpcerr.Mapping{Err: service.ErrTokenExpired, Code: codes.Unauthenticated, Message: "invalid refresh token"},
			grpcerr.Mapping{Err: service.ErrTokenRevoked, Code: codes.Unauthenticated, Message: "invalid refresh token"},
		)
	}

	return mapTokens(tokens), nil
}

func (s *Server) Logout(ctx context.Context, req *usersv1.LogoutRequest) (*usersv1.LogoutResponse, error) {
	err := s.svc.Logout(ctx, req.GetRefreshToken())
	if err != nil {
		return nil, grpcerr.Map(
			err,
			grpcerr.Mapping{Err: service.ErrInvalidToken, Code: codes.Unauthenticated, Message: "invalid refresh token"},
			grpcerr.Mapping{Err: service.ErrTokenExpired, Code: codes.Unauthenticated, Message: "invalid refresh token"},
			grpcerr.Mapping{Err: service.ErrTokenRevoked, Code: codes.Unauthenticated, Message: "invalid refresh token"},
		)
	}

	return &usersv1.LogoutResponse{}, nil
}
