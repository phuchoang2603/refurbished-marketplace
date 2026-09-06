package service

import (
	"errors"
	"os"
	"strings"
)

const (
	defaultOrdersGRPCAddr     = ":9093"
	defaultOrdersKafkaGroupID = "orders-service"
)

type Config struct {
	GRPCAddr     string
	KafkaGroupID string
	ProductsAddr string
}

func LoadConfig() Config {
	cfg := Config{
		GRPCAddr:     strings.TrimSpace(os.Getenv("GRPC_ADDR")),
		KafkaGroupID: strings.TrimSpace(os.Getenv("KAFKA_GROUP_ID")),
		ProductsAddr: strings.TrimSpace(os.Getenv("PRODUCTS_SVC_ADDR")),
	}
	if cfg.GRPCAddr == "" {
		cfg.GRPCAddr = defaultOrdersGRPCAddr
	}
	if cfg.KafkaGroupID == "" {
		cfg.KafkaGroupID = defaultOrdersKafkaGroupID
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
	if strings.TrimSpace(cfg.ProductsAddr) == "" {
		return errors.New("PRODUCTS_SVC_ADDR is required")
	}
	return nil
}
