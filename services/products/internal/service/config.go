package service

import (
	"errors"
	"net/url"
	"os"
	"strings"
)

const defaultProductsGRPCAddr = ":9092"

type Config struct {
	GRPCAddr string
	MongoURI string
}

func LoadConfig() Config {
	cfg := Config{
		GRPCAddr: strings.TrimSpace(os.Getenv("GRPC_ADDR")),
		MongoURI: strings.TrimSpace(os.Getenv("MONGO_URI")),
	}
	if cfg.GRPCAddr == "" {
		cfg.GRPCAddr = defaultProductsGRPCAddr
	}
	if cfg.MongoURI == "" {
		cfg.MongoURI = mongoURIFromParts()
	}
	return cfg
}

func ValidateConfig(cfg Config) error {
	if strings.TrimSpace(cfg.GRPCAddr) == "" {
		return errors.New("GRPC_ADDR is required")
	}
	if strings.TrimSpace(cfg.MongoURI) == "" {
		return errors.New("Mongo connection (MONGO_URI or MONGO_ADDR, MONGO_USER, MONGO_PASSWORD) is required")
	}
	return nil
}

func mongoURIFromParts() string {
	addr := strings.TrimSpace(os.Getenv("MONGO_ADDR"))
	user := strings.TrimSpace(os.Getenv("MONGO_USER"))
	password := os.Getenv("MONGO_PASSWORD")
	database := strings.TrimSpace(os.Getenv("MONGO_DATABASE"))
	authSource := strings.TrimSpace(os.Getenv("MONGO_AUTH_SOURCE"))
	replicaSet := strings.TrimSpace(os.Getenv("MONGO_REPLICA_SET"))
	if addr == "" || user == "" || password == "" {
		return ""
	}
	if database == "" {
		database = "catalog"
	}
	if authSource == "" {
		authSource = "admin"
	}
	u := &url.URL{
		Scheme: "mongodb",
		User:   url.UserPassword(user, password),
		Host:   addr,
		Path:   "/" + database,
	}
	q := url.Values{}
	q.Set("authSource", authSource)
	if replicaSet != "" {
		q.Set("replicaSet", replicaSet)
	}
	u.RawQuery = q.Encode()
	return u.String()
}
