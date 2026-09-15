package service

import (
	"errors"
	"os"
	"strings"
)

const (
	defaultSearchGRPCAddr                   = ":9098"
	defaultSearchProductCreatedKafkaGroupID = "search-product-created"
	defaultListingsIndex                    = "listings"
)

type Config struct {
	GRPCAddr                   string
	MeiliURL                   string
	MeiliAPIKey                string
	ProductCreatedKafkaGroupID string
	ListingsIndex              string
}

func LoadConfig() Config {
	cfg := Config{
		GRPCAddr:                   strings.TrimSpace(os.Getenv("GRPC_ADDR")),
		MeiliURL:                   strings.TrimSpace(os.Getenv("MEILI_URL")),
		MeiliAPIKey:                strings.TrimSpace(os.Getenv("MEILI_MASTER_KEY")),
		ProductCreatedKafkaGroupID: strings.TrimSpace(os.Getenv("KAFKA_PRODUCT_CREATED_GROUP_ID")),
		ListingsIndex:              strings.TrimSpace(os.Getenv("MEILI_LISTINGS_INDEX")),
	}
	if cfg.GRPCAddr == "" {
		cfg.GRPCAddr = defaultSearchGRPCAddr
	}
	if cfg.ProductCreatedKafkaGroupID == "" {
		cfg.ProductCreatedKafkaGroupID = defaultSearchProductCreatedKafkaGroupID
	}
	if cfg.ListingsIndex == "" {
		cfg.ListingsIndex = defaultListingsIndex
	}
	return cfg
}

func ValidateConfig(cfg Config) error {
	if strings.TrimSpace(cfg.GRPCAddr) == "" {
		return errors.New("GRPC_ADDR is required")
	}
	if strings.TrimSpace(cfg.MeiliURL) == "" {
		return errors.New("MEILI_URL is required")
	}
	if strings.TrimSpace(cfg.MeiliAPIKey) == "" {
		return errors.New("MEILI_MASTER_KEY is required")
	}
	if strings.TrimSpace(cfg.ProductCreatedKafkaGroupID) == "" {
		return errors.New("KAFKA_PRODUCT_CREATED_GROUP_ID is required")
	}
	if strings.TrimSpace(cfg.ListingsIndex) == "" {
		return errors.New("MEILI_LISTINGS_INDEX is required")
	}
	return nil
}
