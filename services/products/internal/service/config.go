package service

import (
	"errors"
	"os"
	"strings"
)

const defaultProductsGRPCAddr = ":9092"

type Config struct {
	GRPCAddr string
}

func LoadConfig() Config {
	cfg := Config{
		GRPCAddr: strings.TrimSpace(os.Getenv("GRPC_ADDR")),
	}
	if cfg.GRPCAddr == "" {
		cfg.GRPCAddr = defaultProductsGRPCAddr
	}
	return cfg
}

func ValidateConfig(cfg Config) error {
	if strings.TrimSpace(cfg.GRPCAddr) == "" {
		return errors.New("GRPC_ADDR is required")
	}
	return nil
}
