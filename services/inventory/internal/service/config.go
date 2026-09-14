package service

import (
	"errors"
	"os"
	"strings"
)

const (
	defaultInventoryGRPCAddr                   = ":9097"
	defaultInventoryKafkaGroupID               = "inventory-service"
	defaultInventoryProductCreatedKafkaGroupID = "inventory-product-created"
)

type Config struct {
	GRPCAddr                   string
	KafkaGroupID               string
	ProductCreatedKafkaGroupID string
}

func LoadConfig() Config {
	cfg := Config{
		GRPCAddr:                   strings.TrimSpace(os.Getenv("GRPC_ADDR")),
		KafkaGroupID:               strings.TrimSpace(os.Getenv("KAFKA_GROUP_ID")),
		ProductCreatedKafkaGroupID: strings.TrimSpace(os.Getenv("KAFKA_PRODUCT_CREATED_GROUP_ID")),
	}
	if cfg.GRPCAddr == "" {
		cfg.GRPCAddr = defaultInventoryGRPCAddr
	}
	if cfg.KafkaGroupID == "" {
		cfg.KafkaGroupID = defaultInventoryKafkaGroupID
	}
	if cfg.ProductCreatedKafkaGroupID == "" {
		cfg.ProductCreatedKafkaGroupID = defaultInventoryProductCreatedKafkaGroupID
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
	if strings.TrimSpace(cfg.ProductCreatedKafkaGroupID) == "" {
		return errors.New("KAFKA_PRODUCT_CREATED_GROUP_ID is required")
	}
	if cfg.KafkaGroupID == cfg.ProductCreatedKafkaGroupID {
		return errors.New("KAFKA_PRODUCT_CREATED_GROUP_ID must differ from KAFKA_GROUP_ID")
	}
	return nil
}
