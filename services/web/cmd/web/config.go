package main

import (
	"fmt"
	"os"
)

type config struct {
	addr         string
	usersAddr    string
	productsAddr string
	ordersAddr   string
	cartAddr     string
	paymentAddr  string
	jwtSecret    string
}

func loadConfig() (config, error) {
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	usersAddr, err := requiredEnv("USERS_SVC_ADDR")
	if err != nil {
		return config{}, err
	}
	productsAddr, err := requiredEnv("PRODUCTS_SVC_ADDR")
	if err != nil {
		return config{}, err
	}
	ordersAddr, err := requiredEnv("ORDERS_SVC_ADDR")
	if err != nil {
		return config{}, err
	}
	cartAddr, err := requiredEnv("CART_SVC_ADDR")
	if err != nil {
		return config{}, err
	}
	paymentAddr, err := requiredEnv("PAYMENT_SVC_ADDR")
	if err != nil {
		return config{}, err
	}
	jwtSecret, err := requiredEnv("JWT_SECRET")
	if err != nil {
		return config{}, err
	}

	return config{
		addr:         addr,
		usersAddr:    usersAddr,
		productsAddr: productsAddr,
		ordersAddr:   ordersAddr,
		cartAddr:     cartAddr,
		paymentAddr:  paymentAddr,
		jwtSecret:    jwtSecret,
	}, nil
}

func requiredEnv(name string) (string, error) {
	value := os.Getenv(name)
	if value == "" {
		return "", fmt.Errorf("%s is required", name)
	}
	return value, nil
}
