package service

import (
	"errors"
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
	return Config{
		JWTSecret:     secret,
		JWTIssuer:     "refurbished-marketplace",
		JWTAudience:   "refurbished-marketplace-api",
		JWTAccessTTL:  15 * time.Minute,
		JWTRefreshTTL: 168 * time.Hour,
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
