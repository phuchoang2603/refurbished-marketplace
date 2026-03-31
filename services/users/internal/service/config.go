package service

import (
	"errors"
	authconfig "refurbished-marketplace/shared/auth/config"
	"strings"
	"time"
)

type Config struct {
	JWTSecret     string
	JWTIssuer     string
	JWTAudience   string
	JWTAccessTTL  time.Duration
	JWTRefreshTTL time.Duration
}

func DefaultConfig(secret string) Config {
	base := authconfig.DefaultConfig(secret)
	return Config{
		JWTSecret:     base.JWTSecret,
		JWTIssuer:     base.JWTIssuer,
		JWTAudience:   base.JWTAudience,
		JWTAccessTTL:  authconfig.DefaultJWTAccessTTL,
		JWTRefreshTTL: authconfig.DefaultJWTRefreshTTL,
	}
}

func ValidateConfig(cfg Config) error {
	if strings.TrimSpace(cfg.JWTSecret) == "" {
		return errors.New("JWT_SECRET is required")
	}
	if cfg.JWTAccessTTL <= 0 || cfg.JWTRefreshTTL <= 0 {
		return errors.New("jwt ttl must be positive")
	}
	if cfg.JWTIssuer == "" || cfg.JWTAudience == "" {
		return errors.New("jwt issuer and audience are required")
	}
	return nil
}
