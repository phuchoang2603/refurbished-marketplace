package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Tokens struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	ExpiresIn    int64
}

func (s *Service) Login(ctx context.Context, email string, password string) (Tokens, error) {
	cleanEmail := normalizeEmail(email)
	if !isValidEmailShape(cleanEmail) {
		return Tokens{}, ErrInvalidCredentials
	}

	u, err := s.queries.GetUserByEmail(ctx, cleanEmail)
	if err != nil {
		if mapped := mapInvalidCredentials(err); mapped != nil {
			return Tokens{}, mapped
		}
		return Tokens{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return Tokens{}, ErrInvalidCredentials
	}

	return s.issueTokenPair(ctx, u.ID, u.Email)
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (Tokens, error) {
	claims, err := s.parseToken(refreshToken, "refresh")
	if err != nil {
		return Tokens{}, err
	}

	refreshID, err := uuid.Parse(claims.ID)
	if err != nil {
		return Tokens{}, ErrInvalidToken
	}

	session, err := loadRefreshSession(ctx, s.queries, refreshID)
	if err != nil {
		return Tokens{}, err
	}

	if session.RevokedAt.Valid {
		return Tokens{}, ErrTokenRevoked
	}
	if time.Now().UTC().After(session.ExpiresAt) {
		return Tokens{}, ErrTokenExpired
	}
	if hashToken(refreshToken) != session.TokenHash {
		return Tokens{}, ErrInvalidToken
	}

	if err := s.queries.RevokeRefreshToken(ctx, refreshID); err != nil {
		return Tokens{}, err
	}

	u, err := s.queries.GetUserByID(ctx, session.UserID)
	if err != nil {
		return Tokens{}, mapNotFound(err, ErrUserNotFound)
	}

	return s.issueTokenPair(ctx, u.ID, u.Email)
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	claims, err := s.parseToken(refreshToken, "refresh")
	if err != nil {
		return err
	}

	refreshID, err := uuid.Parse(claims.ID)
	if err != nil {
		return ErrInvalidToken
	}

	session, err := loadRefreshSession(ctx, s.queries, refreshID)
	if err != nil {
		return err
	}

	if session.RevokedAt.Valid {
		return nil
	}

	if hashToken(refreshToken) != session.TokenHash {
		return ErrInvalidToken
	}

	if err := s.queries.RevokeRefreshToken(ctx, refreshID); err != nil {
		return err
	}

	return nil
}
