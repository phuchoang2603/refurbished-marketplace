package service

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"refurbished-marketplace/services/users/internal/database"
	"strings"
	"time"

	sharedjwt "refurbished-marketplace/shared/auth/jwt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func normalizeEmail(email string) string {
	return strings.TrimSpace(strings.ToLower(email))
}

func isValidEmailShape(email string) bool {
	return strings.Contains(email, "@") && len(email) >= 3
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

func mapNotFound(err error, notFoundErr error) error {
	if err == nil {
		return nil
	}
	if err == sql.ErrNoRows {
		return notFoundErr
	}
	return err
}

func (s *Service) signToken(tokenType, jti string, userID uuid.UUID, email string, expiresAt time.Time) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss": s.cfg.JWTIssuer,
		"sub": userID.String(),
		"aud": s.cfg.JWTAudience,
		"exp": expiresAt.Unix(),
		"iat": time.Now().UTC().Unix(),
		"jti": jti,
		"typ": tokenType,
		"eml": email,
	})

	return token.SignedString([]byte(s.cfg.JWTSecret))
}

func (s *Service) parseToken(raw string, expectedType string) (jwt.RegisteredClaims, error) {
	claims, err := sharedjwt.ParseAndValidate(raw, s.cfg.JWTSecret, expectedType, s.cfg.JWTIssuer, s.cfg.JWTAudience)
	if err != nil {
		if errors.Is(err, sharedjwt.ErrExpiredToken) {
			return jwt.RegisteredClaims{}, ErrTokenExpired
		}
		return jwt.RegisteredClaims{}, ErrInvalidToken
	}

	return jwt.RegisteredClaims{
		Issuer:   claims.Issuer,
		Subject:  claims.Subject,
		Audience: jwt.ClaimStrings{claims.Audience},
		ID:       claims.ID,
	}, nil
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

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func mapInvalidToken(err error) error {
	if err == nil {
		return nil
	}
	if err == sql.ErrNoRows {
		return ErrInvalidToken
	}
	return err
}

func loadRefreshSession(ctx context.Context, queries queryStore, refreshID uuid.UUID) (database.RefreshToken, error) {
	session, err := queries.GetRefreshTokenByID(ctx, refreshID)
	if err != nil {
		if mapped := mapInvalidToken(err); mapped != nil {
			return database.RefreshToken{}, mapped
		}
		return database.RefreshToken{}, err
	}

	return session, nil
}

func mapInvalidCredentials(err error) error {
	if err == nil {
		return nil
	}
	if err == sql.ErrNoRows {
		return ErrInvalidCredentials
	}
	return err
}
