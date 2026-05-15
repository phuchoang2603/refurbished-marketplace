package service

import (
	"database/sql"
	"errors"

	"refurbished-marketplace/services/users/internal/database"
)

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
	queries *database.Queries
	cfg     Config
}

func New(db *sql.DB, cfg Config) *Service {
	return &Service{queries: database.New(db), cfg: cfg}
}
