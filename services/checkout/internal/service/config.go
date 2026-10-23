package service

import (
	"errors"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	defaultGRPCAddr     = ":9099"
	defaultKafkaGroupID = "checkout-service"
)

type Config struct {
	GRPCAddr                    string
	KafkaGroupID                string
	RedisAddr                   string
	PaymentGatewayURL           string
	PaymentGatewayAPIKey        string
	PaymentGatewayWebhookSecret string
	CheckoutReturnBaseURL       string
	PaymentDeadlineGrace        time.Duration
}

func LoadConfig() Config {
	config := Config{
		GRPCAddr:                    strings.TrimSpace(os.Getenv("GRPC_ADDR")),
		RedisAddr:                   strings.TrimSpace(os.Getenv("REDIS_ADDR")),
		PaymentGatewayURL:           strings.TrimSpace(os.Getenv("PAYMENT_GATEWAY_URL")),
		PaymentGatewayAPIKey:        os.Getenv("PAYMENT_GATEWAY_API_KEY"),
		PaymentGatewayWebhookSecret: os.Getenv("PAYMENT_GATEWAY_WEBHOOK_SECRET"),
		CheckoutReturnBaseURL:       strings.TrimSpace(os.Getenv("CHECKOUT_RETURN_BASE_URL")),
		PaymentDeadlineGrace:        2 * time.Minute,
		KafkaGroupID:                strings.TrimSpace(os.Getenv("KAFKA_GROUP_ID")),
	}
	if raw := os.Getenv("PAYMENT_GATEWAY_DEADLINE_GRACE"); raw != "" {
		grace, err := time.ParseDuration(raw)
		if err != nil {
			grace = -1
		}
		config.PaymentDeadlineGrace = grace
	}
	if config.GRPCAddr == "" {
		config.GRPCAddr = defaultGRPCAddr
	}
	if config.KafkaGroupID == "" {
		config.KafkaGroupID = defaultKafkaGroupID
	}
	if config.RedisAddr == "" {
		config.RedisAddr = "checkout-redis:6379"
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
	if config.PaymentGatewayURL != "" {
		if config.PaymentDeadlineGrace <= 0 {
			return errors.New("PAYMENT_GATEWAY_DEADLINE_GRACE must be positive")
		}

		endpoint, err := url.Parse(config.PaymentGatewayURL)
		if err != nil || (endpoint.Scheme != "http" && endpoint.Scheme != "https") || endpoint.Hostname() == "" || endpoint.User != nil || endpoint.RawQuery != "" || endpoint.Fragment != "" {
			return errors.New("PAYMENT_GATEWAY_URL must be an HTTP endpoint")
		}
		if config.PaymentGatewayAPIKey == "" || config.PaymentGatewayWebhookSecret == "" {
			return errors.New("gateway API key and webhook secret are required")
		}
		returnURL, err := url.Parse(config.CheckoutReturnBaseURL)
		if err != nil || returnURL.Scheme != "https" || returnURL.Hostname() == "" || returnURL.User != nil {
			return errors.New("CHECKOUT_RETURN_BASE_URL must be HTTPS")
		}
	}
	return nil
}
