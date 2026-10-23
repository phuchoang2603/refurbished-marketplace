package config

import (
	"errors"
	"os"
	"strings"

	"github.com/phuchoang2603/refurbished-marketplace/shared/runtime"
)

const (
	defaultHTTPAddr = ":8080"
)

type Config struct {
	HTTPAddr      string
	UsersAddr     string
	ProductsAddr  string
	InventoryAddr string
	SearchAddr    string
	OrdersAddr    string
	CartAddr      string
	CheckoutAddr  string
	JWTSecret     string
	PublicBaseURL string
}

func LoadConfig() Config {
	return Config{
		HTTPAddr:      runtime.EnvOr("HTTP_ADDR", defaultHTTPAddr),
		UsersAddr:     strings.TrimSpace(os.Getenv("USERS_SVC_ADDR")),
		ProductsAddr:  strings.TrimSpace(os.Getenv("PRODUCTS_SVC_ADDR")),
		InventoryAddr: strings.TrimSpace(os.Getenv("INVENTORY_SVC_ADDR")),
		SearchAddr:    strings.TrimSpace(os.Getenv("SEARCH_SVC_ADDR")),
		OrdersAddr:    strings.TrimSpace(os.Getenv("ORDERS_SVC_ADDR")),
		CartAddr:      strings.TrimSpace(os.Getenv("CART_SVC_ADDR")),
		CheckoutAddr:  strings.TrimSpace(os.Getenv("CHECKOUT_SVC_ADDR")),
		JWTSecret:     strings.TrimSpace(os.Getenv("JWT_SECRET")),
		PublicBaseURL: strings.TrimRight(strings.TrimSpace(os.Getenv("PUBLIC_BASE_URL")), "/"),
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
	if strings.TrimSpace(cfg.InventoryAddr) == "" {
		return errors.New("INVENTORY_SVC_ADDR is required")
	}
	if strings.TrimSpace(cfg.SearchAddr) == "" {
		return errors.New("SEARCH_SVC_ADDR is required")
	}
	if strings.TrimSpace(cfg.OrdersAddr) == "" {
		return errors.New("ORDERS_SVC_ADDR is required")
	}
	if strings.TrimSpace(cfg.CartAddr) == "" {
		return errors.New("CART_SVC_ADDR is required")
	}
	if strings.TrimSpace(cfg.CheckoutAddr) == "" {
		return errors.New("CHECKOUT_SVC_ADDR is required")
	}
	if strings.TrimSpace(cfg.JWTSecret) == "" {
		return errors.New("JWT_SECRET is required")
	}
	return nil
}
