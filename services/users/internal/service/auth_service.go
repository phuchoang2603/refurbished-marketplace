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

func (s *Service) Login(ctx context.Context, email string, password string) (Tokens, error) {
	cleanEmail := strings.TrimSpace(strings.ToLower(email))
	if !strings.Contains(cleanEmail, "@") || len(cleanEmail) < 3 {
		return Tokens{}, ErrInvalidCredentials
	}

	u, err := s.queries.GetUserByEmail(ctx, cleanEmail)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Tokens{}, ErrInvalidCredentials
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

	session, err := s.queries.GetRefreshTokenByID(ctx, refreshID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Tokens{}, ErrInvalidToken
		}
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
		if errors.Is(err, sql.ErrNoRows) {
			return Tokens{}, ErrUserNotFound
		}
		return Tokens{}, err
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

	session, err := s.queries.GetRefreshTokenByID(ctx, refreshID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrInvalidToken
		}
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

func (s *Service) issueTokenPair(ctx context.Context, userID uuid.UUID, email string) (Tokens, error) {
	now := time.Now().UTC()
	accessExpiresAt := now.Add(s.cfg.JWTAccessTTL)
	refreshExpiresAt := now.Add(s.cfg.JWTRefreshTTL)

	accessToken, err := s.signToken("access", uuid.NewString(), userID, email, accessExpiresAt)
	if err != nil {
		return Tokens{}, err
	}

	refreshID := uuid.New()
	refreshToken, err := s.signToken("refresh", refreshID.String(), userID, email, refreshExpiresAt)
	if err != nil {
		return Tokens{}, err
	}

	_, err = s.queries.CreateRefreshToken(ctx, database.CreateRefreshTokenParams{
		ID:        refreshID,
		TokenHash: hashToken(refreshToken),
		UserID:    userID,
		ExpiresAt: refreshExpiresAt,
	})
	if err != nil {
		return Tokens{}, err
	}

	return Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.cfg.JWTAccessTTL.Seconds()),
	}, nil
}
