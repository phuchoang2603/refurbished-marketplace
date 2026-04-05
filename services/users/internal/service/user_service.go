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
	XPos         float64
	YPos         float64
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (s *Service) CreateUser(ctx context.Context, email string, password string, xPos, yPos float64) (User, error) {
	cleanEmail := normalizeEmail(email)
	if !isValidEmailShape(cleanEmail) {
		return User{}, ErrInvalidEmail
	}

	if len(password) < 8 {
		return User{}, ErrInvalidPassword
	}

	if _, err := s.queries.GetUserByEmail(ctx, cleanEmail); err == nil {
		return User{}, ErrEmailTaken
	} else if err = mapNotFound(err, ErrUserNotFound); err != ErrUserNotFound {
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
		XPos:         xPos,
		YPos:         yPos,
	})
	if err != nil {
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
