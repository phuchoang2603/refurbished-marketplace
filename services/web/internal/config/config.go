package config

import (
	"errors"
	"os"
	"strings"

	"refurbished-marketplace/shared/runtime"
)

const (
	defaultHTTPAddr       = ":8080"
	defaultGatewayBaseURL = "http://localhost:8097"
)

type Config struct {
	HTTPAddr        string
	UsersAddr       string
	ProductsAddr    string
	OrdersAddr      string
	CartAddr        string
	PaymentAddr     string
	JWTSecret       string
	GatewayBaseURL  string
	PublicBaseURL   string
	CallbackBaseURL string
}

func LoadConfig() Config {
	gatewayBaseURL := strings.TrimRight(os.Getenv("HOSTED_PAYMENT_BASE_URL"), "/")
	if gatewayBaseURL == "" {
		gatewayBaseURL = strings.TrimRight(os.Getenv("HOSTED_PAYMENT_SIMULATOR_BASE_URL"), "/")
	}
	if gatewayBaseURL == "" {
		gatewayBaseURL = defaultGatewayBaseURL
	}

	return Config{
		HTTPAddr:        runtime.EnvOr("HTTP_ADDR", defaultHTTPAddr),
		UsersAddr:       strings.TrimSpace(os.Getenv("USERS_SVC_ADDR")),
		ProductsAddr:    strings.TrimSpace(os.Getenv("PRODUCTS_SVC_ADDR")),
		OrdersAddr:      strings.TrimSpace(os.Getenv("ORDERS_SVC_ADDR")),
		CartAddr:        strings.TrimSpace(os.Getenv("CART_SVC_ADDR")),
		PaymentAddr:     strings.TrimSpace(os.Getenv("PAYMENT_SVC_ADDR")),
		JWTSecret:       strings.TrimSpace(os.Getenv("JWT_SECRET")),
		GatewayBaseURL:  gatewayBaseURL,
		PublicBaseURL:   strings.TrimRight(strings.TrimSpace(os.Getenv("PUBLIC_BASE_URL")), "/"),
		CallbackBaseURL: strings.TrimRight(strings.TrimSpace(os.Getenv("HOSTED_PAYMENT_CALLBACK_BASE_URL")), "/"),
	}
}

func ValidateConfig(cfg Config) error {
	if strings.TrimSpace(cfg.HTTPAddr) == "" {
		return errors.New("HTTP_ADDR is required")
	}
	if strings.TrimSpace(cfg.UsersAddr) == "" {
		return errors.New("USERS_SVC_ADDR is required")
	}
	if strings.TrimSpace(cfg.ProductsAddr) == "" {
		return errors.New("PRODUCTS_SVC_ADDR is required")
	}
	if strings.TrimSpace(cfg.OrdersAddr) == "" {
		return errors.New("ORDERS_SVC_ADDR is required")
	}
	if strings.TrimSpace(cfg.CartAddr) == "" {
		return errors.New("CART_SVC_ADDR is required")
	}
	if strings.TrimSpace(cfg.PaymentAddr) == "" {
		return errors.New("PAYMENT_SVC_ADDR is required")
	}
	if strings.TrimSpace(cfg.JWTSecret) == "" {
		return errors.New("JWT_SECRET is required")
	}
	return nil
}
