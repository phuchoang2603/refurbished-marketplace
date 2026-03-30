package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"refurbished-marketplace/services/users/internal/database"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (s *Service) CreateUser(ctx context.Context, email string, password string) (User, error) {
	cleanEmail := strings.TrimSpace(strings.ToLower(email))
	if !strings.Contains(cleanEmail, "@") || len(cleanEmail) < 3 {
		return User{}, ErrInvalidEmail
	}

	if len(password) < 8 {
		return User{}, ErrInvalidPassword
	}

	if _, err := s.queries.GetUserByEmail(ctx, cleanEmail); err == nil {
		return User{}, ErrEmailTaken
	} else if !errors.Is(err, sql.ErrNoRows) {
		return User{}, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}

	created, err := s.queries.CreateUser(ctx, database.CreateUserParams{
		ID:           uuid.New(),
		Email:        cleanEmail,
		PasswordHash: string(hash),
	})
	if err != nil {
		return User{}, err
	}

	return mapDBUser(created), nil
}

func (s *Service) GetUserByID(ctx context.Context, id uuid.UUID) (User, error) {
	u, err := s.queries.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrUserNotFound
		}
		return User{}, err
	}

	return mapDBUser(u), nil
}

func mapDBUser(u database.User) User {
	return User{
		ID:           u.ID,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}
