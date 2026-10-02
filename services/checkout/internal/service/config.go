package service

import (
	"errors"
	"os"
	"strings"
)

const (
	defaultGRPCAddr     = ":9099"
	defaultKafkaGroupID = "checkout-service"
)

type Config struct {
	GRPCAddr     string
	KafkaGroupID string
}

func LoadConfig() Config {
	config := Config{
		GRPCAddr:     strings.TrimSpace(os.Getenv("GRPC_ADDR")),
		KafkaGroupID: strings.TrimSpace(os.Getenv("KAFKA_GROUP_ID")),
	}
	if config.GRPCAddr == "" {
		config.GRPCAddr = defaultGRPCAddr
	}
	if config.KafkaGroupID == "" {
		config.KafkaGroupID = defaultKafkaGroupID
	}
	return config
}

func ValidateConfig(config Config) error {
	if strings.TrimSpace(config.GRPCAddr) == "" {
		return errors.New("GRPC_ADDR is required")
	}
	if strings.TrimSpace(config.KafkaGroupID) == "" {
		return errors.New("KAFKA_GROUP_ID is required")
	}
	return nil
}
