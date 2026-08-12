package service

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/phuchoang2603/refurbished-marketplace/shared/runtime"
)

const (
	defaultPaymentGRPCAddr             = ":9096"
	defaultPaymentKafkaGroupID         = "payment-service"
	defaultPaymentSessionSweepInterval = time.Minute
)

type Config struct {
	GRPCAddr             string
	KafkaGroupID         string
	SessionSweepInterval time.Duration
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
	cfg.SessionSweepInterval = runtime.ParseDurationEnv("PAYMENT_SESSION_SWEEP_INTERVAL", defaultPaymentSessionSweepInterval)
	if raw := strings.TrimSpace(os.Getenv("PAYMENT_SESSION_SWEEP_INTERVAL")); raw == "0" {
		cfg.SessionSweepInterval = 0
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
	if cfg.SessionSweepInterval < 0 {
		return errors.New("PAYMENT_SESSION_SWEEP_INTERVAL must be >= 0")
	}
	return nil
}
