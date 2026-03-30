// Package service provides the core business logic for the users service. It defines the Service struct, which has methods for creating users, authenticating them, and managing refresh tokens. The service interacts with the database through a queryStore interface, which abstracts away the database operations. This allows for easier testing and separation of concerns.
package service

import (
	"context"
	"errors"

	"refurbished-marketplace/services/users/internal/database"

	"github.com/google/uuid"
)

type queryStore interface {
	CreateUser(ctx context.Context, arg database.CreateUserParams) (database.User, error)
	GetUserByEmail(ctx context.Context, email string) (database.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (database.User, error)
	CreateRefreshToken(ctx context.Context, arg database.CreateRefreshTokenParams) (database.RefreshToken, error)
	GetRefreshTokenByID(ctx context.Context, id uuid.UUID) (database.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, id uuid.UUID) error
}

var (
	ErrInvalidEmail       = errors.New("invalid email")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrEmailTaken         = errors.New("email already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token expired")
	ErrTokenRevoked       = errors.New("token revoked")
)

type Service struct {
	queries queryStore
	cfg     Config
}

func New(queries queryStore, cfg Config) *Service {
	return &Service{queries: queries, cfg: cfg}
}
