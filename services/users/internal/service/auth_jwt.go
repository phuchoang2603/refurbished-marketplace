package service

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
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
	parsed, err := jwt.Parse(raw, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, ErrInvalidToken
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return jwt.RegisteredClaims{}, ErrTokenExpired
		}
		return jwt.RegisteredClaims{}, ErrInvalidToken
	}

	claimsMap, ok := parsed.Claims.(jwt.MapClaims)
	if !ok || !parsed.Valid {
		return jwt.RegisteredClaims{}, ErrInvalidToken
	}

	typ, _ := claimsMap["typ"].(string)
	iss, _ := claimsMap["iss"].(string)
	aud, _ := claimsMap["aud"].(string)
	sub, _ := claimsMap["sub"].(string)
	jti, _ := claimsMap["jti"].(string)

	if typ != expectedType || iss != s.cfg.JWTIssuer || aud != s.cfg.JWTAudience || sub == "" || jti == "" {
		return jwt.RegisteredClaims{}, ErrInvalidToken
	}

	return jwt.RegisteredClaims{
		Issuer:   iss,
		Subject:  sub,
		Audience: jwt.ClaimStrings{aud},
		ID:       jti,
	}, nil
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
