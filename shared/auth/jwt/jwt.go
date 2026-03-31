package jwt

import (
	"errors"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

type Claims struct {
	Subject  string
	Type     string
	ID       string
	Issuer   string
	Audience string
}

func ParseAndValidate(raw, secret, expectedType, expectedIssuer, expectedAudience string) (Claims, error) {
	parsed, err := jwtlib.Parse(raw, func(t *jwtlib.Token) (any, error) {
		if t.Method != jwtlib.SigningMethodHS256 {
			return nil, ErrInvalidToken
		}
		return []byte(secret), nil
	})
	if err != nil {
		if errors.Is(err, jwtlib.ErrTokenExpired) {
			return Claims{}, ErrExpiredToken
		}
		return Claims{}, ErrInvalidToken
	}

	claimsMap, ok := parsed.Claims.(jwtlib.MapClaims)
	if !ok || !parsed.Valid {
		return Claims{}, ErrInvalidToken
	}

	typ, _ := claimsMap["typ"].(string)
	iss, _ := claimsMap["iss"].(string)
	aud, _ := claimsMap["aud"].(string)
	sub, _ := claimsMap["sub"].(string)
	jti, _ := claimsMap["jti"].(string)

	if typ != expectedType || iss != expectedIssuer || aud != expectedAudience || sub == "" || jti == "" {
		return Claims{}, ErrInvalidToken
	}

	return Claims{Subject: sub, Type: typ, ID: jti, Issuer: iss, Audience: aud}, nil
}
