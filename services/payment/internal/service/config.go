package service

import (
	"errors"
	"os"
	"strings"
)

const (
	defaultPaymentGRPCAddr     = ":9096"
	defaultPaymentKafkaGroupID = "payment-service"
)

type Config struct {
	GRPCAddr     string
	KafkaGroupID string
}

func LoadConfig() Config {
	cfg := Config{
		GRPCAddr:     strings.TrimSpace(os.Getenv("GRPC_ADDR")),
		KafkaGroupID: strings.TrimSpace(os.Getenv("KAFKA_GROUP_ID")),
	}
	if cfg.GRPCAddr == "" {
		cfg.GRPCAddr = defaultPaymentGRPCAddr
	}
	if cfg.KafkaGroupID == "" {
		cfg.KafkaGroupID = defaultPaymentKafkaGroupID
	}
	return cfg
}

func ValidateConfig(cfg Config) error {
	if strings.TrimSpace(cfg.GRPCAddr) == "" {
		return errors.New("GRPC_ADDR is required")
	}
	if strings.TrimSpace(cfg.KafkaGroupID) == "" {
		return errors.New("KAFKA_GROUP_ID is required")
	}
	return nil
}
