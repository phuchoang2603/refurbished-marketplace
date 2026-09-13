package service

import (
	"errors"
	"os"
	"strings"
)

const (
	defaultInventoryGRPCAddr     = ":9097"
	defaultInventoryKafkaGroupID = "inventory-service"
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
		cfg.GRPCAddr = defaultInventoryGRPCAddr
	}
	if cfg.KafkaGroupID == "" {
		cfg.KafkaGroupID = defaultInventoryKafkaGroupID
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
