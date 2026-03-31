package service

import (
	"crypto/sha256"
	"encoding/hex"
	errorspkg "errors"
	sharedjwt "refurbished-marketplace/shared/auth/jwt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Tokens struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	ExpiresIn    int64
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
		if errorspkg.Is(err, sharedjwt.ErrExpiredToken) {
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

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
