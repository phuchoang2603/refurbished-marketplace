package service

import (
	"errors"
	"os"
	"strings"
	"time"
)

const (
	defaultCartGRPCAddr = ":9094"
	defaultCartTTL      = 24 * time.Hour
)

type Config struct {
	GRPCAddr  string
	RedisAddr string
	CartTTL   time.Duration
}

func LoadConfig() Config {
	ttl := defaultCartTTL
	if raw := strings.TrimSpace(os.Getenv("CART_TTL")); raw != "" {
		if parsed, err := time.ParseDuration(raw); err == nil && parsed > 0 {
			ttl = parsed
		}
	}

	cfg := Config{
		GRPCAddr:  strings.TrimSpace(os.Getenv("GRPC_ADDR")),
		RedisAddr: strings.TrimSpace(os.Getenv("REDIS_ADDR")),
		CartTTL:   ttl,
	}
	if cfg.GRPCAddr == "" {
		cfg.GRPCAddr = defaultCartGRPCAddr
	}
	return cfg
}

func ValidateConfig(cfg Config) error {
	if strings.TrimSpace(cfg.GRPCAddr) == "" {
		return errors.New("GRPC_ADDR is required")
	}
	if strings.TrimSpace(cfg.RedisAddr) == "" {
		return errors.New("REDIS_ADDR is required")
	}
	if cfg.CartTTL <= 0 {
		return errors.New("CART_TTL must be positive")
	}
	return nil
}
