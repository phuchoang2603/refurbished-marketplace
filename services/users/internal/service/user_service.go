package service

import (
	"context"
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

func (s *Service) CreateUser(ctx context.Context, email, password string) (User, error) {
	cleanEmail := normalizeEmail(email)
	if !isValidEmailShape(cleanEmail) {
		return User{}, ErrInvalidEmail
	}

	if len(password) < 8 {
		return User{}, ErrInvalidPassword
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
		if isPostgresUniqueViolation(err) {
			return User{}, ErrEmailTaken
		}
		return User{}, err
	}

	return mapDBUser(database.User(created)), nil
}

func (s *Service) GetUserByID(ctx context.Context, id uuid.UUID) (User, error) {
	u, err := s.queries.GetUserByID(ctx, id)
	if err != nil {
		return User{}, mapNotFound(err, ErrUserNotFound)
	}

	return mapDBUser(database.User(u)), nil
}
