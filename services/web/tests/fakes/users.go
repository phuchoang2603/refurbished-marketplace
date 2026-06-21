package fakes

import (
	"context"

	usersv1 "refurbished-marketplace/shared/proto/users/v1"
)

type UsersService struct {
	LoginFn      func(context.Context, string, string) (*usersv1.TokenResponse, error)
	LogoutFn     func(context.Context, string) (*usersv1.LogoutResponse, error)
	CreateUserFn func(context.Context, string, string) (*usersv1.User, error)
}

func (f *UsersService) Login(ctx context.Context, email, password string) (*usersv1.TokenResponse, error) {
	if f.LoginFn != nil {
		return f.LoginFn(ctx, email, password)
	}
	return nil, nil
}

func (f *UsersService) Logout(ctx context.Context, refreshToken string) (*usersv1.LogoutResponse, error) {
	if f.LogoutFn != nil {
		return f.LogoutFn(ctx, refreshToken)
	}
	return &usersv1.LogoutResponse{}, nil
}

func (f *UsersService) CreateUser(ctx context.Context, email, password string) (*usersv1.User, error) {
	if f.CreateUserFn != nil {
		return f.CreateUserFn(ctx, email, password)
	}
	return &usersv1.User{}, nil
}
